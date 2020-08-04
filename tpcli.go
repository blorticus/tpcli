package tpcli

import (
	"fmt"
	"log"

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
	errorPanel
	commandHistoryPanel
	generalOutputPanel
)

// ControlMessageType is
type ControlMessageType int

// ControlMessage types
const (
	Error ControlMessageType = iota
	UserSuppliedCommandString
	ReplacementForCommandStringValue
	AdditionalGeneralOutput
	AdditionalErrorOutput
)

// ControlMessage is
type ControlMessage struct {
	Type ControlMessageType
	Body string
}

// Tpcli is
type Tpcli struct {
	stackingOrder          StackingOrder
	tviewApplication       *tview.Application
	userCommandInputPanel  *commandInputPanel
	errorOutputPanel       *outputPanel
	generalOutputPanel     *outputPanel
	userInputStringChannel chan string
	debugLogger            *log.Logger
	//userCommandHistoryTextView *tview.TextView
}

// NewTpcli does
func NewTpcli() *Tpcli {
	return &Tpcli{
		stackingOrder: GeneralErrorCommand,
	}
}

// UsingStackingOrder is
func (ui *Tpcli) UsingStackingOrder(order StackingOrder) *Tpcli {
	ui.stackingOrder = order
	return ui
}

// UsingCommandHistoryPanel is
func (ui *Tpcli) UsingCommandHistoryPanel() *Tpcli {
	return ui
}

// Start is
func (ui *Tpcli) Start() {

}

// Stop is
func (ui *Tpcli) Stop() {
}

// ChannelOfControlMessagesFromTheUI is
func (ui *Tpcli) ChannelOfControlMessagesFromTheUI() <-chan *ControlMessage {
	return nil
}

// ReplaceCommandStringWith is
func (ui *Tpcli) ReplaceCommandStringWith(newString string) {

}

// AddStringToGeneralOutput is
func (ui *Tpcli) AddStringToGeneralOutput(additionalContent string) {

}

// AddStringToErrorOutput is
func (ui *Tpcli) AddStringToErrorOutput(additionalContent string) {

}

func (ui *Tpcli) buildTviewUIBasedOnStackingOrder() *Tpcli {
	ui.tviewApplication = tview.NewApplication()

	switch ui.stackingOrder {
	case CommandErrorGeneral:
		ui.addCommandInputPanel().
			addErrorPanel().
			addGeneralOutputPanel().
			composeIntoUIGrid([]panelTypes{commandPanel, errorPanel, generalOutputPanel})
	}

	ui.addGlobalKeybindings()

	return ui
}

func (ui *Tpcli) addCommandInputPanel() *Tpcli {
	ui.userCommandInputPanel = newCommandInputPanel(ui.tviewApplication)
	return ui
}

func (ui *Tpcli) addErrorPanel() *Tpcli {
	ui.errorOutputPanel = newOutputPanel(ui.tviewApplication)
	return ui
}

func (ui *Tpcli) addGeneralOutputPanel() *Tpcli {
	ui.generalOutputPanel = newOutputPanel(ui.tviewApplication)
	return ui
}

func (ui *Tpcli) addCommandHistoryPanel() *Tpcli {
	return ui
}

func (ui *Tpcli) composeIntoUIGrid(orderedListOfPanelsForLayout []panelTypes) *Tpcli {
	grid := tview.NewGrid()

	gridRowSizes := make([]int, 3)

	for i, panelType := range orderedListOfPanelsForLayout {
		switch panelType {
		case commandPanel:
			gridRowSizes[i] = 1
			grid.AddItem(ui.userCommandInputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, true)

		case errorPanel:
			gridRowSizes[i] = 10
			grid.AddItem(ui.errorOutputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, false)

		case commandHistoryPanel:
			gridRowSizes[i] = 10
			grid.AddItem(ui.errorOutputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, false)

		case generalOutputPanel:
			gridRowSizes[i] = 0
			grid.AddItem(ui.errorOutputPanel.BackingTviewObject(), i, 0, 1, 1, 0, 0, false)
		}
	}

	grid.
		SetRows(gridRowSizes...).
		SetColumns(0)

	ui.tviewApplication.SetRoot(grid, true)

	return ui
}

func (ui *Tpcli) addGlobalKeybindings() *Tpcli {
	ui.tviewApplication.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// case tcell.KeyTab:
		// 	switch ui.tviewApplication.GetFocus() {
		// 	case ui.userCommandHistoryTextView:
		// 		ui.tviewApplication.SetFocus(ui.userCommandInputField)
		// 	case ui.userCommandInputField:
		// 		ui.tviewApplication.SetFocus(ui.eventOutputTextView)
		// 	default:
		// 		ui.tviewApplication.SetFocus(ui.userCommandHistoryTextView)
		// 	}
		// 	return nil
		case tcell.KeyESC:
			ui.Exit()
		case tcell.KeyCtrlQ:
			ui.Exit()
		}

		return event
	})

	return ui
}

// func (ui *TestHarnessTextUI) sendNextInputCommandToChannelWithoutBlocking(commandText string) {
// 	go func() { ui.userInputStringChannel <- commandText }()
// }

// // UserInputStringCommandChannel retrieves a string channel that will contain user input
// // provided in the command input box
// func (ui *TestHarnessTextUI) UserInputStringCommandChannel() <-chan string {
// 	return ui.userInputStringChannel
// }

// // StartRunning launches the UI after its construction
// func (ui *TestHarnessTextUI) StartRunning() error {
// 	if err := ui.tviewApplication.Run(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Exit stops the application and exits with a status of zero
// func (ui *TestHarnessTextUI) Exit() {
// 	ui.tviewApplication.Stop()
// 	os.Exit(0)
// }

type commandInputPanel struct {
	promptTextWithTrailingSpace string
	parentTviewApplication      *tview.Application
	tviewInputField             *tview.InputField
	userCommandReadlineHistory  *readlineHistory
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

func (panel *commandInputPanel) BackingTviewObject() *tview.InputField {
	return panel.tviewInputField
}

func (panel *commandInputPanel) createPanelTviewInputField() {
	panel.tviewInputField = tview.NewInputField().
		SetLabel(panel.promptTextWithTrailingSpace).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldWidth(100).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				userProvidedCommandText := panel.tviewInputField.GetText()

				panel.userCommandReadlineHistory.AddItem(userProvidedCommandText)
				panel.userCommandReadlineHistory.ResetIteration()
				// if ui.userCommandHistoryTextView.GetText(false) == "" {
				// 	fmt.Fprintf(ui.userCommandHistoryTextView, userProvidedCommandText)
				// } else {
				// 	fmt.Fprintf(ui.userCommandHistoryTextView, "\n%s", userProvidedCommandText)
				// }
				//ui.userCommandHistoryTextView.SetText(userProvidedCommandText)
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

	textView.SetBorder(true)

	textView.SetChangedFunc(func() {
		parentTviewApplication.Draw()
	})

	return &outputPanel{
		textView: textView,
	}
}

func (panel *outputPanel) BackingTviewObject() *tview.TextView {
	return panel.textView
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
