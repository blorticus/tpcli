package tpcli

import (
	"fmt"
	"io/ioutil"
	"log"

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
				// if panel.commandHistoryTextView.GetText(false) == "" {
				// 	fmt.Fprintf(panel.commandHistoryTextView, userProvidedCommandText)
				// } else {
				// 	fmt.Fprintf(ui.commandHistoryTextView, "\n%s", userProvidedCommandText)
				// }
				// goroutine so as not to block here and thus hold up UI rendering
				go func() { panel.channelOfEnteredCommandStrings <- userProvidedCommandText }()
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

// SimpleStringCircularQueue is a simple circular buffer of strings
type SimpleStringCircularQueue struct {
	stringSlice             []string
	capacity                uint
	headIndex               uint
	indexOfNextInsert       uint
	indexOfLastSliceElement uint
	countOfItemsInQueue     uint
}

// NewSimpleStringCircularBuffer creates a simple circular buffer of strings, with
// the queue holding up to 'capacity' number of items.
func NewSimpleStringCircularBuffer(capacity uint) *SimpleStringCircularQueue {
	return &SimpleStringCircularQueue{
		stringSlice:             make([]string, capacity),
		headIndex:               0,
		indexOfNextInsert:       0,
		indexOfLastSliceElement: capacity - 1,
		countOfItemsInQueue:     0,
	}
}

// PutItemAtEnd places an item at the end of the circular queue
func (queue *SimpleStringCircularQueue) PutItemAtEnd(item string) {
	if queue.countOfItemsInQueue > 0 && queue.indexOfNextInsert == queue.headIndex {
		if queue.headIndex == queue.indexOfLastSliceElement {
			queue.headIndex = 0
		} else {
			queue.headIndex++
		}
	}

	queue.stringSlice[queue.indexOfNextInsert] = item

	if queue.indexOfNextInsert == queue.indexOfLastSliceElement {
		queue.indexOfNextInsert = 0
	} else {
		queue.indexOfNextInsert++
	}

	if queue.countOfItemsInQueue < uint(len(queue.stringSlice)) {
		queue.countOfItemsInQueue++
	}
}

// IsEmpty returns true if the queue has no items in it; false otherwise
func (queue *SimpleStringCircularQueue) IsEmpty() bool {
	return queue.countOfItemsInQueue == 0
}

// IsNotEmpty returns true if the queue has at least one item in it; false otherwise
func (queue *SimpleStringCircularQueue) IsNotEmpty() bool {
	return queue.countOfItemsInQueue != 0
}

// NumberOfItemsInTheQueue returns a count of the number of items in the queue
func (queue *SimpleStringCircularQueue) NumberOfItemsInTheQueue() uint {
	return queue.countOfItemsInQueue
}

// GetItemAtIndex retrieves the string at the specified index (0 is the first item)
func (queue *SimpleStringCircularQueue) GetItemAtIndex(index uint) (item string, thereIsAnItemAtThatIndex bool) {
	if queue.countOfItemsInQueue == 0 || index > queue.countOfItemsInQueue-1 {
		return "", false
	}

	sliceIndexOfItem := queue.headIndex + index
	if sliceIndexOfItem > queue.indexOfLastSliceElement {
		sliceIndexOfItem -= (queue.indexOfLastSliceElement + 1)
	}

	return queue.stringSlice[sliceIndexOfItem], true
}

type readlineHistory struct {
	attachedQueue           *SimpleStringCircularQueue
	indexOfLastItemReturned uint
	iterationHasStarted     bool
}

func newReadlineHistory(maximumHistoryEntries uint) *readlineHistory {
	return &readlineHistory{
		attachedQueue:           NewSimpleStringCircularBuffer(maximumHistoryEntries),
		indexOfLastItemReturned: 0,
		iterationHasStarted:     false,
	}
}

func (history *readlineHistory) Up() (historyItem string, wasNotYetAtTopOfList bool) {
	if history.attachedQueue.IsNotEmpty() {
		if history.iterationHasStarted {
			if history.iteratorIsNotAtStartOfHistoryList() {
				v, _ := history.attachedQueue.GetItemAtIndex(history.indexOfLastItemReturned - 1)
				history.indexOfLastItemReturned--
				return v, true
			}
		} else {
			history.iterationHasStarted = true
			v, _ := history.attachedQueue.GetItemAtIndex(history.attachedQueue.NumberOfItemsInTheQueue() - 1)
			history.indexOfLastItemReturned = history.attachedQueue.NumberOfItemsInTheQueue() - 1
			return v, true
		}
	}

	return "", false
}

func (history *readlineHistory) Down() (historyItem string, wasNotYetAtBottomOfList bool) {
	if history.attachedQueue.IsNotEmpty() {
		if history.iterationHasStarted {
			if history.iteratorIsNotAtEndOfHistoryList() {
				v, _ := history.attachedQueue.GetItemAtIndex(history.indexOfLastItemReturned + 1)
				history.indexOfLastItemReturned++
				return v, true
			}
		}
	}

	return "", false
}

func (history *readlineHistory) iteratorIsNotAtEndOfHistoryList() bool {
	return history.attachedQueue.NumberOfItemsInTheQueue() > history.indexOfLastItemReturned+1
}

func (history *readlineHistory) iteratorIsNotAtStartOfHistoryList() bool {
	return history.indexOfLastItemReturned != 0
}

func (history *readlineHistory) ResetIteration() {
	history.indexOfLastItemReturned = 0
	history.iterationHasStarted = false
}

func (history *readlineHistory) AddItem(item string) {
	history.attachedQueue.PutItemAtEnd(item)
}
