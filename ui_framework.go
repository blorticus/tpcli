package tpcli

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// UIOutputPanel is
type UIOutputPanel interface {
	ClearPanel()
	AppendTextToPanel(additionalText string)
	Write(p []byte) (n int, err error)
}

// UICommandEntryPanel is
type UICommandEntryPanel interface {
	SetPromptTo(newPromptText string)
	SetCommandOutputChannelTo(newChannelOfEnteredCommandStrings chan<- string)
	ClearCommandText()
	SetCommandTextTo(commandText string)
}

// UIFramework is
type UIFramework interface {
	InitializeUIContainer() error
	AddTheGeneralOutputPanel() UIOutputPanel
	AddTheErrorPanel() UIOutputPanel
	AddTheCommandHistoryPanel()
	AddTheCommandInputPanel(channelOfEnteredCommandStrings chan<- string) UIOutputPanel
	Start() error
}

type defaultOutputPanel struct {
	uiTextView *tview.TextView
	parentUI   *defaultUIFramework
}

func (panel *defaultOutputPanel) ClearPanel() {
	panel.uiTextView.SetText("")
}

func (panel *defaultOutputPanel) AppendTextToPanel(additionalText string) {
	if panel.uiTextView.GetText(false) == "" {
		fmt.Fprintf(panel.uiTextView, additionalText)
	} else {
		fmt.Fprintf(panel.uiTextView, "\n%s", additionalText)
	}
}

func (panel *defaultOutputPanel) Write(p []byte) (n int, err error) {
	panel.AppendTextToPanel(string(p))
	return len(p), nil
}

type defaultCommandInputPanel struct {
	uiInputField                   *tview.InputField
	channelOfEnteredCommandStrings chan<- string
	promptTextWithTrailingSpace    string
	commandReadlineHistory         *readlineHistory
}

func newDefaultCommandInputPanel() *defaultCommandInputPanel {
	panel := &defaultCommandInputPanel{
		promptTextWithTrailingSpace: "Enter command> ",
		commandReadlineHistory:      newReadlineHistory(200),
	}

	panel.uiInputField = tview.NewInputField().
		SetLabel(panel.promptTextWithTrailingSpace).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldWidth(100).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				userProvidedCommandText := panel.uiInputField.GetText()

				panel.commandReadlineHistory.AddItem(userProvidedCommandText)
				panel.commandReadlineHistory.ResetIteration()
				go func() { panel.channelOfEnteredCommandStrings <- strings.Trim(userProvidedCommandText, " \n") }()
				panel.uiInputField.SetText("")
			}
		})

	panel.uiInputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp {
			if commandFromHistory, thereWereMoreCommandsInHistory := panel.commandReadlineHistory.Up(); thereWereMoreCommandsInHistory {
				panel.uiInputField.SetText(commandFromHistory)
			}
			return nil
		} else if event.Key() == tcell.KeyDown {
			if commandFromHistory, wasNotYetAtFirstCommand := panel.commandReadlineHistory.Down(); wasNotYetAtFirstCommand {
				panel.uiInputField.SetText(commandFromHistory)
			}
			return nil
		} else {
			return event
		}
	})

	return panel
}

func (panel *defaultCommandInputPanel) SetPromptTo(newPromptText string) {
	panel.promptTextWithTrailingSpace = newPromptText + " "
}

func (panel *defaultCommandInputPanel) SetCommandOutputChannelTo(newChannelOfEnteredCommandStrings chan<- string) {
	panel.channelOfEnteredCommandStrings = newChannelOfEnteredCommandStrings
}

func (panel *defaultCommandInputPanel) ClearCommandText() {
	panel.uiInputField.SetLabel(panel.promptTextWithTrailingSpace)
}
func (panel *defaultCommandInputPanel) SetCommandTextTo(commandText string) {
	panel.uiInputField.SetText(commandText)
}

type defaultUIFramework struct {
	tviewApplication  *tview.Application
	commandInputPanel *defaultCommandInputPanel
	// one of commandHistoryTextView or errorOuputTextView should be nil
	// depending on which type of panel was selected
	commandHistoryTextView *tview.TextView
	errorOutputTextView    *tview.TextView
	generalOutputTextView  *tview.TextView
	debugLogger            *log.Logger
}

func newDefaultUIFramework() *defaultUIFramework {
	return &defaultUIFramework{
		debugLogger: log.New(ioutil.Discard, "", 0),
	}
}

// child UI...Panels may need to compel a redraw, and are provided the
// UIFramework enclosing parent
func (ui *defaultUIFramework) draw() {
	ui.tviewApplication.Draw()
}

func (ui *defaultUIFramework) InitializeUIContainer() error {
	ui.tviewApplication = tview.NewApplication()
	return nil
}

func (ui *defaultUIFramework) AddTheGeneralOutputPanel() UIOutputPanel {
	if ui.generalOutputTextView != nil {
		panic("There is already a general output panel placed")
	}

	ui.generalOutputTextView = tview.NewTextView()

	ui.generalOutputTextView.SetBorder(true)

	ui.generalOutputTextView.SetChangedFunc(func() {
		ui.draw()
	})

	return &defaultOutputPanel{
		uiTextView: ui.generalOutputTextView,
		parentUI:   ui,
	}
}

func (ui *defaultUIFramework) AddTheErrorPanel() UIOutputPanel {
	if ui.errorOutputTextView != nil {
		panic("There is already an error panel placed")
	}

	if ui.commandHistoryTextView != nil {
		panic("A command history panel was already placed; there can be only an error panel or a command history panel, but not both")
	}

	ui.errorOutputTextView = tview.NewTextView()

	ui.errorOutputTextView.SetBorder(true)

	ui.errorOutputTextView.SetChangedFunc(func() {
		ui.draw()
	})

	return &defaultOutputPanel{
		uiTextView: ui.errorOutputTextView,
		parentUI:   ui,
	}
}

func (ui *defaultUIFramework) AddTheCommandHistoryPanel() UIOutputPanel {
	if ui.commandHistoryTextView != nil {
		panic("There is already a command history panel placed")
	}

	if ui.errorOutputTextView != nil {
		panic("An error panel was already placed; there can be only an error panel or a command history panel, but not both")
	}

	ui.commandHistoryTextView = tview.NewTextView()

	ui.commandHistoryTextView.SetBorder(true)

	ui.commandHistoryTextView.SetChangedFunc(func() {
		ui.draw()
	})

	return &defaultOutputPanel{
		uiTextView: ui.commandHistoryTextView,
		parentUI:   ui,
	}
}

func (ui *defaultUIFramework) AddTheCommandInputPanel(channelOfEnteredCommandStrings chan<- string) UICommandEntryPanel {
	ui.commandInputPanel = newDefaultCommandInputPanel()
	ui.commandInputPanel.SetCommandOutputChannelTo(channelOfEnteredCommandStrings)
	return ui.commandInputPanel
}

func (ui *defaultUIFramework) Start() (UIOutputPanel, error) {
	return nil, nil
}
