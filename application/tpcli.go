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

	tpcliStackingOrder, usingErrorPanel := mainApplication.DeriveTpcliPanelStackingOrderFromCliProcessorStackOrder(cliArgumentsProcessor.DesiredPanelStackingOrder())

	ui := tpcli.NewUI()
	ui.ChangeStackingOrderTo(tpcliStackingOrder)

	if !usingErrorPanel {
		ui.UsingCommandHistoryPanel()
	}

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

	go ui.Start()
	go broker.StartListening()

	for {
		messageFromPeer := <-channelOfMessagesFromPeer
		ui.FmtToGeneralOutput("Received message from peer.  Type = (%d), Message = (%s)\n", messageFromPeer.Type, messageFromPeer.Message)
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

func (app *application) DeriveTpcliPanelStackingOrderFromCliProcessorStackOrder(appStackOrder []int) (tpcliStackingOrder tpcli.StackingOrder, usingErrorPanel bool) {
	switch appStackOrder[0] {
	case outputPanel:
		switch appStackOrder[1] {
		case errorPanel:
			return tpcli.GeneralErrorCommand, true
		case historyPanel:
			return tpcli.GeneralErrorCommand, false
		case commandEntryPanel:
			if appStackOrder[2] == errorPanel {
				return tpcli.GeneralCommandError, true
			}
			return tpcli.GeneralCommandError, false
		}
	case errorPanel:
		switch appStackOrder[1] {
		case outputPanel:
			return tpcli.ErrorGeneralCommand, true
		case commandEntryPanel:
			return tpcli.ErrorCommandGeneral, true
		}
	case historyPanel:
		switch appStackOrder[1] {
		case outputPanel:
			return tpcli.ErrorGeneralCommand, false
		case commandEntryPanel:
			return tpcli.ErrorCommandGeneral, false
		}
	case commandEntryPanel:
		switch appStackOrder[1] {
		case outputPanel:
			if appStackOrder[2] == errorPanel {
				return tpcli.CommandGeneralError, true
			}
			return tpcli.CommandGeneralError, false
		case errorPanel:
			return tpcli.CommandErrorGeneral, true
		case historyPanel:
			return tpcli.CommandErrorGeneral, false
		}
	}

	return tpcli.GeneralErrorCommand, false
}

func (app *application) activateDebugLoggingUsingFile(fileName string) {
	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0640)
	panicIfError(err)

	app.debugLogger = log.New(fileHandle, "", 0)
}

func (app *application) deactivateDebugLogging() {
	app.debugLogger = log.New(ioutil.Discard, "", 0)
}

func (app *application) setPanelStackingOrderToUserSpecifiedValue(order string) {

}

func (app *application) setPanelStackingOrderToDefault() {

}

func (app *application) setEhPanelChoiceToUserSpecifiedValue(value string) {

}

func (app *application) setEhPanelChoiceToDefault() {

}

func (app *application) bindCommandSocket(bindType string, bindValue string) {

}
