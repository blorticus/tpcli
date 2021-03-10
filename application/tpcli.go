package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/blorticus/tpcli"
)

func main() {
	mainApplication := &application{}

	cliArgumentsProcessor, err := ProcessCliArguments()
	mainApplication.dieIfError(err)

	if cliArgumentsProcessor.WantsToLogToDebugFile() {
		mainApplication.activateDebugLoggingUsingFile(cliArgumentsProcessor.DebugLogFileFullPath())
	} else {
		mainApplication.deactivateDebugLogging()
	}

	var broker *PeerCommunicationBroker
	if cliArgumentsProcessor.WantsToBindToTCPSocket() {
		broker = BindUsingTCPSocket(cliArgumentsProcessor.TCPSocketBindAddress())
	} else {
		broker = BindUsingUnixSocket(cliArgumentsProcessor.DebugLogFileFullPath())
	}

	channelOfMessagesFromPeer := broker.ChannelOfMessagesFromPeers()

	tpcliStackingOrder, usingCommandHistoryPanel := mainApplication.DeriveTpcliPanelStackingOrderFromCliProcessorStackOrder(cliArgumentsProcessor.DesiredPanelStackingOrder())

	ui := tpcli.NewUI()
	ui.ChangeStackingOrderTo(tpcliStackingOrder)

	if usingCommandHistoryPanel {
		ui.UsingCommandHistoryPanel()
	}

	channelOfUserEnteredCommands := ui.ChannelOfEnteredCommands()

	broker.
		OnIncomingPeerAccept(func(broker *PeerCommunicationBroker, peerConnection net.Conn) {
			ui.FmtToGeneralOutput("Incoming connection from (%s)", peerConnection.RemoteAddr().String())
		}).
		OnPeerClosure(func(broker *PeerCommunicationBroker, peerConnection net.Conn) {
			ui.FmtToGeneralOutput("Connection closed for peer (%s)", peerConnection.RemoteAddr().String())
		}).
		OnGeneralCommunicationError(func(broker *PeerCommunicationBroker, err error) {
			ui.FmtToErrorOutput("General error: %s", err.Error())
		}).
		OnPeerCommunicationError(func(broker *PeerCommunicationBroker, peerConnection net.Conn, err error) {
			ui.FmtToErrorOutput("Peer communication error with peer (%s): %s", peerConnection.RemoteAddr().String(), err.Error())
		})

	ui.OnUIExit(func() {
		broker.SendMessageToPeer(&PeerMessage{
			Type:    UserExited,
			Message: "",
		})
		broker.Terminate()
		os.Exit(0)
	})

	go ui.Start()
	go broker.StartListening()

	for {
		select {
		case messageFromPeer := <-channelOfMessagesFromPeer:
			switch messageFromPeer.Type {
			case ProtocolError:
				ui.FmtToErrorOutput("Peer reports protocol error: %s", messageFromPeer.Message)
			case InputCommandReplacement:
				ui.ReplaceCommandStringWith(messageFromPeer.Message)
			case GeneralOutput:
				ui.AddStringToGeneralOutput(messageFromPeer.Message)
			case ErrorOuput:
				ui.AddStringToErrorOutput(messageFromPeer.Message)
			default:
				broker.SendMessageToPeer(&PeerMessage{
					Type:    ProtocolError,
					Message: fmt.Sprintf("invalid type (%s)", messageFromPeer.TypeAsString()),
				})
			}
		case userEnteredCommand := <-channelOfUserEnteredCommands:
			broker.SendMessageToPeer(&PeerMessage{Type: InputCommandReceived, Message: userEnteredCommand})
		}
	}
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type application struct {
	debugLogger *log.Logger
}

func (app *application) die(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

func (app *application) dieIfError(err error) {
	if err != nil {
		app.die(err.Error())
	}
}

func (app *application) DeriveTpcliPanelStackingOrderFromCliProcessorStackOrder(appStackOrder []int) (tpcliStackingOrder tpcli.StackingOrder, usingCommandHistoryPanel bool) {
	switch appStackOrder[0] {
	case outputPanel:
		switch appStackOrder[1] {
		case errorPanel:
			return tpcli.GeneralErrorCommand, false
		case historyPanel:
			return tpcli.GeneralErrorCommand, true
		case commandEntryPanel:
			if appStackOrder[2] == errorPanel {
				return tpcli.GeneralCommandError, false
			}
			return tpcli.GeneralCommandError, true
		}
	case errorPanel:
		switch appStackOrder[1] {
		case outputPanel:
			return tpcli.ErrorGeneralCommand, false
		case commandEntryPanel:
			return tpcli.ErrorCommandGeneral, false
		}
	case historyPanel:
		switch appStackOrder[1] {
		case outputPanel:
			return tpcli.ErrorGeneralCommand, true
		case commandEntryPanel:
			return tpcli.ErrorCommandGeneral, true
		}
	case commandEntryPanel:
		switch appStackOrder[1] {
		case outputPanel:
			if appStackOrder[2] == errorPanel {
				return tpcli.CommandGeneralError, false
			}
			return tpcli.CommandGeneralError, true
		case errorPanel:
			return tpcli.CommandErrorGeneral, false
		case historyPanel:
			return tpcli.CommandErrorGeneral, true
		}
	}

	return tpcli.GeneralErrorCommand, true
}

func (app *application) activateDebugLoggingUsingFile(fileName string) {
	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0640)
	panicIfError(err)

	app.debugLogger = log.New(fileHandle, "", 0)
}

func (app *application) deactivateDebugLogging() {
	app.debugLogger = log.New(ioutil.Discard, "", 0)
}
