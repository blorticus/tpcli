package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

// PeerMessageJSON is the json package type mapping for a peer message
type PeerMessageJSON struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// PeerMessageType represents types of peer message
type PeerMessageType int

// Peer message types
const (
	ProtocolError PeerMessageType = iota
	InputCommandReceived
	InputCommandReplacement
	GeneralOutput
	ErrorOuput
)

// PeerMessage represents a message delivered to or received from a remote peer
type PeerMessage struct {
	Type    PeerMessageType
	Message string
}

type bindType int

const (
	unixBindType bindType = iota
	tcpBindType
)

// PeerCommunicationBroker handles communication via a TCP or Unix stream socket peer.  This system only allows a single
// peer to be connected at a time.  If a peer is connected and another peer attempts to connect, the connection
// will be rejected.
type PeerCommunicationBroker struct {
	bindType                         bindType
	unixBindSocketPath               string
	tcpListenAddress                 *net.TCPAddr
	channelOfMessagesFromPeers       chan *PeerMessage
	incomingPeerAcceptHandler        func(broker *PeerCommunicationBroker, peerConnection net.Conn)
	peerClosureHandler               func(broker *PeerCommunicationBroker, peerConnection net.Conn)
	peerCommunicationErrorHandler    func(broker *PeerCommunicationBroker, peerConnection net.Conn, err error)
	generalCommunicationErrorHandler func(broker *PeerCommunicationBroker, err error)
}

// BindUsingUnixSocket attempts to bind to an existing Unix stream socket and, on success, returns
// a MessageBroker.  This action will attempt to remove the socketFilePath if it exists.  A failure to do so
// will result in an error.
func BindUsingUnixSocket(socketFilePath string) *PeerCommunicationBroker {
	broker := &PeerCommunicationBroker{
		bindType:                   unixBindType,
		unixBindSocketPath:         socketFilePath,
		tcpListenAddress:           nil,
		channelOfMessagesFromPeers: make(chan *PeerMessage, 10),
	}

	return broker.setAllHandlersToEmpty()
}

// BindUsingTCPSocket attempts to bind to a TCP socket for listening and, on success, returns
// a MessageBroker.
func BindUsingTCPSocket(tcpAddress *net.TCPAddr) *PeerCommunicationBroker {
	broker := &PeerCommunicationBroker{
		bindType:                   tcpBindType,
		unixBindSocketPath:         "",
		tcpListenAddress:           tcpAddress,
		channelOfMessagesFromPeers: make(chan *PeerMessage, 10),
	}

	return broker.setAllHandlersToEmpty()
}

func (broker *PeerCommunicationBroker) setAllHandlersToEmpty() *PeerCommunicationBroker {
	broker.incomingPeerAcceptHandler = func(broker *PeerCommunicationBroker, conn net.Conn) {
		go func() {
			for {
				<-broker.channelOfMessagesFromPeers
			}
		}()
	}

	broker.peerClosureHandler = func(*PeerCommunicationBroker, net.Conn) {}
	broker.peerCommunicationErrorHandler = func(*PeerCommunicationBroker, net.Conn, error) {}
	broker.generalCommunicationErrorHandler = func(*PeerCommunicationBroker, error) {}

	return broker
}

// OnIncomingPeerAccept sets a callback which is executed each time an incoming peer is accepted.  It does not fire when a peer
// connection is rejected (because another peer is already connected).  The peerIncomingMessageChannel is a channel of messages received
// from the peer.  Incoming peer messages cannot be of type InputcommandReceived, so these are filtered out (and an error is sent
// to the peer if messages of this type -- or any invalid type -- are received).
func (broker *PeerCommunicationBroker) OnIncomingPeerAccept(callback func(broker *PeerCommunicationBroker, peerConnection net.Conn)) *PeerCommunicationBroker {
	broker.incomingPeerAcceptHandler = callback
	return broker
}

// OnPeerClosure sets a callback which is executed each time that a peer connection closes.
func (broker *PeerCommunicationBroker) OnPeerClosure(callback func(broker *PeerCommunicationBroker, peerConnection net.Conn)) *PeerCommunicationBroker {
	broker.peerClosureHandler = callback
	return broker
}

// OnPeerCommunicationError sets a callback which is executed each time there is an error with a remote peer.
func (broker *PeerCommunicationBroker) OnPeerCommunicationError(callback func(broker *PeerCommunicationBroker, peerConnection net.Conn, err error)) *PeerCommunicationBroker {
	broker.peerCommunicationErrorHandler = callback
	return broker
}

// OnGeneralCommunicationError sets a callback which is executed each time there is a general communication error, not specific to
// a connected peer.
func (broker *PeerCommunicationBroker) OnGeneralCommunicationError(callback func(broker *PeerCommunicationBroker, err error)) *PeerCommunicationBroker {
	broker.generalCommunicationErrorHandler = callback
	return broker
}

// ChannelOfMessagesFromPeers retrieves a channel onto which incoming PeerMessages are emitted
func (broker *PeerCommunicationBroker) ChannelOfMessagesFromPeers() <-chan *PeerMessage {
	return broker.channelOfMessagesFromPeers
}

// StartListening causes this broker to start listening for incoming peers and handling message between peers.  This should be invoked as a goroutine.
func (broker *PeerCommunicationBroker) StartListening() {
	var listener net.Listener
	var err error

	if broker.bindType == unixBindType {
		if err := os.RemoveAll(broker.unixBindSocketPath); err != nil {
			broker.generalCommunicationErrorHandler(broker, fmt.Errorf("On attempt to remove socket file path (%s): %s", broker.unixBindSocketPath, err.Error()))
			return
		}

		if listener, err = net.Listen("unix", broker.unixBindSocketPath); err != nil {
			broker.generalCommunicationErrorHandler(broker, err)
			return
		}
	} else {
		if listener, err = net.Listen("tcp", broker.tcpListenAddress.String()); err != nil {
			broker.generalCommunicationErrorHandler(broker, err)
			return
		}
	}
	defer listener.Close()

	peerConnection, err := listener.Accept()
	if err != nil {
		broker.generalCommunicationErrorHandler(broker, err)
		return
	}
	defer peerConnection.Close()

	broker.incomingPeerAcceptHandler(broker, peerConnection)

	jsonDecoder := json.NewDecoder(peerConnection)

	var nextJSONMessage PeerMessageJSON
	var nextPeerMessage *PeerMessage

	for {
		if err = jsonDecoder.Decode(&nextJSONMessage); err != nil {
			if err == io.EOF {
				broker.peerClosureHandler(broker, peerConnection)
				break
			}
			broker.peerCommunicationErrorHandler(broker, peerConnection, fmt.Errorf("Error decoding incoming JSON: %s", err.Error()))
		} else {
			if nextPeerMessage, err = broker.convertPeerMessageJSONToMessageObject(&nextJSONMessage); err != nil {
				broker.peerCommunicationErrorHandler(broker, peerConnection, fmt.Errorf("Error decoding incoming JSON: %s", err.Error()))
			}

			broker.channelOfMessagesFromPeers <- nextPeerMessage
		}
	}
}

func (broker *PeerCommunicationBroker) convertPeerMessageJSONToMessageObject(jsonMessage *PeerMessageJSON) (*PeerMessage, error) {
	switch jsonMessage.Type {
	case "protocol_error":
		return &PeerMessage{Type: ProtocolError, Message: jsonMessage.Message}, nil
	case "input_command_received":
		return &PeerMessage{Type: InputCommandReceived, Message: jsonMessage.Message}, nil
	case "input_command_replacement":
		return &PeerMessage{Type: InputCommandReplacement, Message: jsonMessage.Message}, nil
	case "general_output":
		return &PeerMessage{Type: GeneralOutput, Message: jsonMessage.Message}, nil
	case "error_output":
		return &PeerMessage{Type: ErrorOuput, Message: jsonMessage.Message}, nil
	default:
		return nil, fmt.Errorf("Invalid type (%s) in peer message", jsonMessage.Type)
	}
}
