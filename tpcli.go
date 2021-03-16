package tpcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StackingOrder represents the order in which the three panels should be arranged vertically.
// All panels consume all horizontal space (i.e., they all use the same number of columns).
type StackingOrder int

// Various stacking orders.  "Command" is the command input panel.  "General" is the general
// output panel.  "Error" is the error (or command-history) panel.
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

// Tpcli provides a terminal text interfaces.  It creates three "panels": a command
// entry panel, a general output panel and third panel that is either for error
// output or which records the history of entered commands.  The command entry
// panel supports basic shell-emacs bindings (e.g., ^a to go to the start of the
// line, ^e to the end of the line) and arrow key readline-style history navigation.
type Tpcli struct {
	tviewApplication              *tview.Application
	commandInputPanel             *commandInputPanel
	generalOutputPanel            *outputPanel
	errorOrHistoryPanel           *outputPanel
	userInputStringChannel        chan string
	panelTypesInOrder             []panelTypes
	indexInOrderOfPanelWithFocus  int
	useErrorPanelAsCommandHistory bool
	functionToExecuteAfterUIExits func()
}

// NewUI constructs the UI interface elements for the Tpcli but does not start showing
// them (that happens on invocation of Start())
func NewUI() *Tpcli {
	ui := &Tpcli{
		userInputStringChannel:        make(chan string, 10),
		panelTypesInOrder:             []panelTypes{generalOutputPanel, errorOrHistoryPanel, commandPanel},
		indexInOrderOfPanelWithFocus:  2,
		functionToExecuteAfterUIExits: func() { os.Exit(0) },
		useErrorPanelAsCommandHistory: false,
	}

	return ui
}

// Write allows an instance of tpcli to be used as a Writer.  Any bytes provided will be interpreted
// as an ASCII string and will be written to the General Output panel.
func (ui *Tpcli) Write(p []byte) (n int, err error) {
	ui.AddStringToGeneralOutput(string(p))
	return len(p), nil
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

	for i, panelType := range ui.panelTypesInOrder {
		if panelType == commandPanel {
			ui.indexInOrderOfPanelWithFocus = i
			break
		}
	}

	return ui
}

// OnUIExit provides a function that is executed by the Tpcli immediately after it Stops, and the
// UI is terminated.  This function is executed when a UI exit is provided, including ^q or <esc>.
func (ui *Tpcli) OnUIExit(functionToExecuteAfterUIExits func()) *Tpcli {
	ui.functionToExecuteAfterUIExits = functionToExecuteAfterUIExits
	return ui
}

// UsingCommandHistoryPanel instructs the Tcpli to use the error panel as a command history.  When
// this is set, any command entered in the command panel is copied here after the user hits <enter>.
// Any text that the caller attempts to write to the error panel is redirected to the
// general output panel
func (ui *Tpcli) UsingCommandHistoryPanel() *Tpcli {
	ui.useErrorPanelAsCommandHistory = true
	return ui
}

// Start instructs Tpcli to draw the UI and start handling keyboard events.  This should
// be invoked as a goroutine.
func (ui *Tpcli) Start() {
	ui.createTviewApplication().
		createPanelForErrorOrCommandHistory().
		createCommandInputPanel().
		createGeneralOutputPanel().
		composeIntoUIGridUsingStackOrder(ui.panelTypesInOrder).
		addGlobalKeybindings()

	go ui.tviewApplication.Run()
}

func (ui *Tpcli) exit() {
	ui.Stop()
	ui.functionToExecuteAfterUIExits()
}

// Stop instructs Tpcli to stop the UI, clearing it.  This is not the same as an exit
// triggered by ^q or <esc>.  This only stops the UI.  It does not exit the function provided
// by OnUIExit.
func (ui *Tpcli) Stop() {
	ui.tviewApplication.Stop()
}

// ChannelOfEnteredCommands is a channel that emits the commands that the user enters in the
// command panel. A command is a string of UTF-8 text that ends with a newline (signaled by
// the <enter> key).  The commands strings sent on this channel omit the trailing newline.
func (ui *Tpcli) ChannelOfEnteredCommands() <-chan string {
	return ui.userInputStringChannel
}

// ReplaceCommandStringWith writes the newString to the command panel, replacing whatever
// is currently there.
func (ui *Tpcli) ReplaceCommandStringWith(newString string) {
	ui.commandInputPanel.ChangeCommandStringTo(newString)
}

// AddStringToGeneralOutput appends additionalContent to whatever text is currently in the
// general output panel.  A newline (\n) is appended to the text that is already there first,
// then the new text is appended.
func (ui *Tpcli) AddStringToGeneralOutput(additionalContent string) {
	ui.generalOutputPanel.AppendText(additionalContent)
}

// FmtToGeneralOutput is the same as AddStringToGeneralOutput, but it takes fmt.Sprintf
// parameters and expands them using that mechanism
func (ui *Tpcli) FmtToGeneralOutput(format string, a ...interface{}) {
	ui.AddStringToGeneralOutput(fmt.Sprintf(format, a...))
}

// AddStringToErrorOutput appends additionalContent to the error panel in the same way that
// AddStringToGeneralOutput does. However, if UsingCommandHistoryPanel is invoked, then
// any additionalContent submitted here is instead written to the general output panel.
func (ui *Tpcli) AddStringToErrorOutput(additionalContent string) {
	if ui.useErrorPanelAsCommandHistory {
		ui.generalOutputPanel.AppendText(additionalContent)
	} else {
		ui.errorOrHistoryPanel.AppendText(additionalContent)
	}
}

// FmtToErrorOutput is the same as AddStringToErrorOutput, but it takes fmt.Sprintf
// parameters and expands them using that mechanism
func (ui *Tpcli) FmtToErrorOutput(format string, a ...interface{}) {
	ui.AddStringToErrorOutput(fmt.Sprintf(format, a...))
}

func (ui *Tpcli) createTviewApplication() *Tpcli {
	ui.tviewApplication = tview.NewApplication()
	return ui
}

func (ui *Tpcli) createCommandInputPanel() *Tpcli {
	ui.commandInputPanel = newCommandInputPanel(ui.tviewApplication)
	if ui.useErrorPanelAsCommandHistory {
		ui.commandInputPanel.WhenACommandIsEntered(func(command string) {
			go func() { ui.userInputStringChannel <- command }()
			ui.errorOrHistoryPanel.AppendText(command)
			switch command {
			case "quit":
				fallthrough
			case "exit":
				ui.exit()
			}
		})
	} else {
		ui.commandInputPanel.WhenACommandIsEntered(func(command string) {
			go func() { ui.userInputStringChannel <- command }()
			switch command {
			case "quit":
				fallthrough
			case "exit":
				ui.exit()
			}
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
			ui.indexInOrderOfPanelWithFocus++
			if ui.indexInOrderOfPanelWithFocus >= len(ui.panelTypesInOrder) {
				ui.indexInOrderOfPanelWithFocus = 0
			}
			switch ui.panelTypesInOrder[ui.indexInOrderOfPanelWithFocus] {
			case commandPanel:
				ui.tviewApplication.SetFocus(ui.commandInputPanel.BackingTviewObject())
			case generalOutputPanel:
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

type commandInputPanel struct {
	promptTextWithTrailingSpace string
	parentTviewApplication      *tview.Application
	tviewInputField             *tview.InputField
	userCommandReadlineHistory  *ReadlineHistory
	callbackOnEnteredCommand    func(string)
}

func newCommandInputPanel(parentTviewApplication *tview.Application) *commandInputPanel {
	panel := &commandInputPanel{
		parentTviewApplication:      parentTviewApplication,
		promptTextWithTrailingSpace: "Enter command> ",
		userCommandReadlineHistory:  NewReadlineHistory(200),
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
			panel.tviewInputField.SetText(panel.userCommandReadlineHistory.Up())
			return nil
		} else if event.Key() == tcell.KeyDown {
			panel.tviewInputField.SetText(panel.userCommandReadlineHistory.Down())
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
		fmt.Fprint(panel.textView, s)
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
