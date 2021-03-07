package main

import (
	"flag"
	"fmt"
	"net"
)

const (
	outputPanel = iota
	errorPanel
	historyPanel
	commandEntryPanel
)

// CliProcessor processes command-line arguments for the tpcli application
type CliProcessor struct {
	usingTCPBindSocket bool
	tcpBindAddress     *net.TCPAddr
	unixSocketPath     string
	panelStack         []int
	wantsDebugLogging  bool
	debugLogFilepath   string
}

// ProcessCliArguments processes os.Args, searching for requisite flags.  It validates any values passed
// to the flags, and sets defaults for optional flags that are not provided.  If a validation error occurs,
// return the error.
func ProcessCliArguments() (*CliProcessor, error) {
	processor := &CliProcessor{
		usingTCPBindSocket: false,
		tcpBindAddress:     nil,
		unixSocketPath:     "",
		panelStack:         make([]int, 3),
		wantsDebugLogging:  false,
		debugLogFilepath:   "",
	}

	tcpBindParameter := flag.String("tcp", "", "ip:tcp-port on which this application should listen for commands")
	unixBindParameter := flag.String("unix", "", "Path to unix socket on which this application should listen for commands")
	orderParameter := flag.String("order", "ohc", "Three letters representing panel stack order (o, h, e, and c)")
	debugParameter := flag.String("debug", "", "Path to debug log file if debugging is desired")

	flag.Parse()

	if err := processor.processBindParameter(*tcpBindParameter, *unixBindParameter); err != nil {
		return nil, err
	}

	if err := processor.processOrderParameter(*orderParameter); err != nil {
		return nil, err
	}

	if err := processor.processDebugParameter(*debugParameter); err != nil {
		return nil, err
	}

	return processor, nil
}

// WantsToLogToDebugFile is true if the user provided the -debug flag.
func (processor *CliProcessor) WantsToLogToDebugFile() bool {
	return processor.wantsDebugLogging
}

// DebugLogFileFullPath returns the full path to the debug log file the user provided.  If -debug was
// not supplied, the return result is undefined.
func (processor *CliProcessor) DebugLogFileFullPath() string {
	return processor.debugLogFilepath
}

// DesiredPanelStackingOrder returns a list of three elements, indicating the preferred panel stacking
// order.
func (processor *CliProcessor) DesiredPanelStackingOrder() []int {
	return processor.panelStack
}

// WantsToBindToTCPSocket returns true if the user wants to bind to a tcp socket.
func (processor *CliProcessor) WantsToBindToTCPSocket() bool {
	return processor.usingTCPBindSocket
}

// WantsToBindToUnixSocket returns true if the user wants to bind to a unix socket.
func (processor *CliProcessor) WantsToBindToUnixSocket() bool {
	return !processor.usingTCPBindSocket
}

// TCPSocketBindAddress returns the command-line provided TCP bind address.
func (processor *CliProcessor) TCPSocketBindAddress() *net.TCPAddr {
	return processor.tcpBindAddress
}

// UnixSocketBindEndpoint returns the full path to the unix socket file entry if a unix socket bind
// is desired.  If a unix socket bind is not desired, the result is undefined.
func (processor *CliProcessor) UnixSocketBindEndpoint() string {
	return ""
}

func (processor *CliProcessor) processBindParameter(tcpBindParameterValue string, unixBindParameterValue string) error {
	if tcpBindParameterValue == "" && unixBindParameterValue == "" {
		processor.usingTCPBindSocket = true
		processor.tcpBindAddress, _ = net.ResolveTCPAddr("tcp", "localhost:6000")
		return nil
	}

	if tcpBindParameterValue != "" {
		if unixBindParameterValue != "" {
			return fmt.Errorf("Cannot provide both -tcp and -unix")
		}

		addr, err := net.ResolveTCPAddr("tcp", tcpBindParameterValue)
		if err != nil {
			return fmt.Errorf("(%s) cannot be used as a bind address", tcpBindParameterValue)
		}

		processor.usingTCPBindSocket = true
		processor.tcpBindAddress = addr

	} else {
		processor.usingTCPBindSocket = false
		processor.unixSocketPath = unixBindParameterValue
	}

	return nil
}

func (processor *CliProcessor) processOrderParameter(orderParameterValue string) error {
	if len(orderParameterValue) != 3 {
		return fmt.Errorf("-order must be exactly three letters")
	}

	orderLettersGiven := make(map[rune]bool)
	for _, orderLetter := range string(orderParameterValue) {
		if _, letterAlreadyProvided := orderLettersGiven[orderLetter]; letterAlreadyProvided {
			return fmt.Errorf("In -order, a single letter cannot be provided more than once")
		}
		switch orderLetter {
		case 'o':
			processor.panelStack = append(processor.panelStack, outputPanel)
		case 'h':
			processor.panelStack = append(processor.panelStack, historyPanel)
		case 'e':
			processor.panelStack = append(processor.panelStack, errorPanel)
		case 'c':
			processor.panelStack = append(processor.panelStack, commandEntryPanel)
		default:
			return fmt.Errorf("In -order, only 'o', 'h', 'e', and 'c' are allowed")
		}
		orderLettersGiven[orderLetter] = true
	}

	if _, providedLetterH := orderLettersGiven['h']; providedLetterH {
		if _, providedLetterE := orderLettersGiven['e']; providedLetterE {
			return fmt.Errorf("-order must have exactly one of 'h' or 'e', but cannot have both")
		}
	}

	return nil
}

func (processor *CliProcessor) processDebugParameter(debugParameterValue string) error {
	if debugParameterValue != "" {
		processor.wantsDebugLogging = true
		processor.debugLogFilepath = debugParameterValue
	}
	return nil
}
