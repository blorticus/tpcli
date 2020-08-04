package tpcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// StackingOrder is
type StackingOrder int

// Various stacking orders
const (
	CommandErrorGeneral StackingOrder = iota
	CommandGeneralError
	GeneralCommandError
	GeneralErrorCommand
	ErrorCommandGeneral
	ErrorGeneralCommand
)

type panelTypes int

const (
	commandPanel panelTypes = iota
	errorOrHistoryPanel
	commandHistoryPanel
	generalOutputPanel
)

// Tpcli provides a TUI based text interface.  It has a command
// entry box with basic EMACS-like control keys (^k, ^a, ^e, ^b) used with a bash
// shell.  Another box shows the last 10 commands entered.  A third box is used for
// general output from the application logic.
type Tpcli struct {
	tviewApplication              *tview.Application
	commandInputPanel             *commandInputPanel
	generalOutputPanel            *outputPanel
	errorOrHistoryPanel           *outputPanel
	isUsingAHistoryPanel          bool
	userInputStringChannel        chan string
	panelTypesInOrder             []panelTypes
	functionToExecuteAfterUIExits func()
}

// NewUI constructs the UI interface elements
func NewUI() *Tpcli {
	ui := &Tpcli{
		userInputStringChannel:        make(chan string, 10),
		panelTypesInOrder:             []panelTypes{generalOutputPanel, errorOrHistoryPanel, commandPanel},
		functionToExecuteAfterUIExits: func() { os.Exit(0) },
		isUsingAHistoryPanel:          false,
	}

	ui.createTviewApplication().
		createPanelForErrorOrCommandHistory().
		createCommandInputPanel().
		createGeneralOutputPanel().
		composeIntoUIGridUsingStackOrder(ui.panelTypesInOrder).
		addGlobalKeybindings()

	return ui
}

// ChangeStackingOrderTo changes the panel stacking order to the provided ordering
func (ui *Tpcli) ChangeStackingOrderTo(newOrder StackingOrder) *Tpcli {
	switch newOrder {
	case CommandErrorGeneral:
		ui.panelTypesInOrder = []panelTypes{commandPanel, errorOrHistoryPanel, generalOutputPanel}
	case CommandGeneralError:
		ui.panelTypesInOrder = []panelTypes{commandPanel, generalOutputPanel, errorOrHistoryPanel}
	case GeneralCommandError:
		ui.panelTypesInOrder = []panelTypes{generalOutputPanel, commandPanel, errorOrHistoryPanel}
	case GeneralErrorCommand:
		ui.panelTypesInOrder = []panelTypes{generalOutputPanel, errorOrHistoryPanel, commandPanel}
	case ErrorCommandGeneral:
		ui.panelTypesInOrder = []panelTypes{errorOrHistoryPanel, commandPanel, generalOutputPanel}
	case ErrorGeneralCommand:
		ui.panelTypesInOrder = []panelTypes{errorOrHistoryPanel, generalOutputPanel, commandPanel}
	}

	return ui
}

// OnUIExit is
func (ui *Tpcli) OnUIExit(functionToExecuteAfterUIExits func()) *Tpcli {
	ui.functionToExecuteAfterUIExits = functionToExecuteAfterUIExits
	return ui
}

// UsingCommandHistoryPanel is
func (ui *Tpcli) UsingCommandHistoryPanel() *Tpcli {
	ui.isUsingAHistoryPanel = true
	return ui
}

// Start is
func (ui *Tpcli) Start() {
	go ui.tviewApplication.Run()
}

func (ui *Tpcli) exit() {
	ui.Stop()
	ui.functionToExecuteAfterUIExits()
}

// Stop is
func (ui *Tpcli) Stop() {
	ui.tviewApplication.Stop()
}

// ChannelOfEnteredCommands is
func (ui *Tpcli) ChannelOfEnteredCommands() <-chan string {
	return ui.userInputStringChannel
}

// ReplaceCommandStringWith is
func (ui *Tpcli) ReplaceCommandStringWith(newString string) {
	ui.commandInputPanel.ChangeCommandStringTo(newString)
}

// AddToGeneralOutputText is
func (ui *Tpcli) AddToGeneralOutputText(additionalContent string) {
	ui.generalOutputPanel.AppendText(additionalContent)
}

// AddToErrorText is
func (ui *Tpcli) AddToErrorText(additionalContent string) {
	if ui.isUsingAHistoryPanel {
		ui.generalOutputPanel.AppendText(additionalContent)
	} else {
		ui.errorOrHistoryPanel.AppendText(additionalContent)
	}
}

func (ui *Tpcli) createTviewApplication() *Tpcli {
	ui.tviewApplication = tview.NewApplication()
	return ui
}

func (ui *Tpcli) sendNextInputCommandToChannelWithoutBlocking(commandText string) {
	go func() { ui.userInputStringChannel <- commandText }()
}

func (ui *Tpcli) createCommandInputPanel() *Tpcli {
	ui.commandInputPanel = newCommandInputPanel(ui.tviewApplication)

	if ui.isUsingAHistoryPanel {
		ui.commandInputPanel.WhenACommandIsEntered(func(command string) {
			ui.sendNextInputCommandToChannelWithoutBlocking(command)
			ui.errorOrHistoryPanel.AppendText(command)
		})
	} else {
		ui.commandInputPanel.WhenACommandIsEntered(func(command string) {
			ui.sendNextInputCommandToChannelWithoutBlocking(command)
		})
	}

	return ui
}

func (ui *Tpcli) createGeneralOutputPanel() *Tpcli {
	ui.generalOutputPanel = newOutputPanel(ui.tviewApplication)
	return ui
}

func (ui *Tpcli) createPanelForErrorOrCommandHistory() *Tpcli {
	ui.errorOrHistoryPanel = newOutputPanel(ui.tviewApplication).SetTitleTo("Command History")
	return ui
}

func (ui *Tpcli) composeIntoUIGridUsingStackOrder(panelOrderByType []panelTypes) *Tpcli {
	grid := tview.NewGrid()

	rowSizes := make([]int, 3)

	for i, panelType := range panelOrderByType {
		switch panelType {
		case generalOutputPanel:
			rowSizes[i] = 0

		case errorOrHistoryPanel:
			rowSizes[i] = 12

		case commandPanel:
			rowSizes[i] = 3
		}
	}

	grid.
		SetRows(rowSizes...).
		SetColumns(0)

	// The SetRows() must be completed before laying these out
	for i, panelType := range panelOrderByType {
		switch panelType {
		case generalOutputPanel:
			grid.AddItem(ui.generalOutputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, false)
		case errorOrHistoryPanel:
			grid.AddItem(ui.errorOrHistoryPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, false)
		case commandPanel:
			grid.AddItem(ui.commandInputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, true)
		}
	}

	ui.tviewApplication.SetRoot(grid, true)

	return ui
}

func (ui *Tpcli) addGlobalKeybindings() *Tpcli {
	ui.tviewApplication.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			switch ui.tviewApplication.GetFocus() {
			case ui.errorOrHistoryPanel.BackingTviewObject():
				ui.tviewApplication.SetFocus(ui.commandInputPanel.BackingTviewObject())
			case ui.commandInputPanel.BackingTviewObject():
				ui.tviewApplication.SetFocus(ui.generalOutputPanel.BackingTviewObject())
			default:
				ui.tviewApplication.SetFocus(ui.errorOrHistoryPanel.BackingTviewObject())
			}
			return nil
		case tcell.KeyESC:
			ui.exit()
		case tcell.KeyCtrlQ:
			ui.exit()
		}

		return event
	})

	return ui
}

// UserInputStringCommandChannel retrieves a string channel that will contain user input
// provided in the command input box
func (ui *Tpcli) UserInputStringCommandChannel() <-chan string {
	return ui.userInputStringChannel
}

type commandInputPanel struct {
	promptTextWithTrailingSpace string
	parentTviewApplication      *tview.Application
	tviewInputField             *tview.InputField
	userCommandReadlineHistory  *readlineHistory
	callbackOnEnteredCommand    func(string)
}

func newCommandInputPanel(parentTviewApplication *tview.Application) *commandInputPanel {
	panel := &commandInputPanel{
		parentTviewApplication:      parentTviewApplication,
		promptTextWithTrailingSpace: "Enter command> ",
		userCommandReadlineHistory:  newReadlineHistory(200),
	}

	panel.createPanelTviewInputField()

	return panel
}

func (panel *commandInputPanel) ChangePromptTo(promptWithoutTrainingSpace string) *commandInputPanel {
	panel.promptTextWithTrailingSpace = promptWithoutTrainingSpace + " "
	return panel
}

func (panel *commandInputPanel) BackingTviewObject() tview.Primitive {
	return panel.tviewInputField
}

func (panel *commandInputPanel) WhenACommandIsEntered(doThis func(commandWithoutTrailingNewline string)) *commandInputPanel {
	panel.callbackOnEnteredCommand = doThis
	return panel
}

func (panel *commandInputPanel) ChangeCommandStringTo(newString string) {
	panel.tviewInputField.SetText(newString)
}

func (panel *commandInputPanel) createPanelTviewInputField() {
	panel.tviewInputField = tview.NewInputField().
		SetLabel(panel.promptTextWithTrailingSpace).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldWidth(100).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				userProvidedCommandTextTrimmed := strings.TrimSpace(panel.tviewInputField.GetText())

				panel.userCommandReadlineHistory.AddItem(userProvidedCommandTextTrimmed)
				panel.userCommandReadlineHistory.ResetIteration()
				panel.callbackOnEnteredCommand(userProvidedCommandTextTrimmed)
				panel.tviewInputField.SetText("")
			}
		})

	panel.tviewInputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp {
			if commandFromHistory, thereWereMoreCommandsInHistory := panel.userCommandReadlineHistory.Up(); thereWereMoreCommandsInHistory {
				panel.tviewInputField.SetText(commandFromHistory)
			}
			return nil
		} else if event.Key() == tcell.KeyDown {
			if commandFromHistory, wasNotYetAtFirstCommand := panel.userCommandReadlineHistory.Down(); wasNotYetAtFirstCommand {
				panel.tviewInputField.SetText(commandFromHistory)
			}
			return nil
		} else {
			return event
		}
	})
}

type outputPanel struct {
	textView *tview.TextView
}

func newOutputPanel(parentTviewApplication *tview.Application) *outputPanel {
	textView := tview.NewTextView()

	textView.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	textView.SetChangedFunc(func() {
		parentTviewApplication.Draw()
	})

	return &outputPanel{
		textView: textView,
	}
}

func (panel *outputPanel) BackingTviewObject() tview.Primitive {
	return panel.textView
}

func (panel *outputPanel) SetTitleTo(newTitle string) *outputPanel {
	panel.textView.SetTitle(newTitle)
	return panel
}

func (panel *outputPanel) AppendText(s string) {
	if panel.textView.GetText(false) == "" {
		fmt.Fprintf(panel.textView, s)
	} else {
		fmt.Fprintf(panel.textView, "\n%s", s)
	}
}

func (panel *outputPanel) Write(p []byte) (int, error) {
	panel.AppendText(string(p))
	return len(p), nil
}

func (panel *outputPanel) Clear() {
	panel.textView.SetText("")
}

type uiPanel interface {
	BackingTviewObject() tview.Primitive
}
