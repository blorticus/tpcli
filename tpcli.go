package tpcli

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
	stackingOrder StackingOrder
}

// NewTpcli does
func NewTpcli() *Tpcli {
	return &Tpcli{
		stackingOrder: GeneralErrorCommand,
	}
}

// UsingStackingOrder is
func (ui *Tpcli) UsingStackingOrder(order StackingOrder) *Tpcli {
	return ui
}

// UsingCommandHistoryPanel is
func (ui *Tpcli) UsingCommandHistoryPanel() *Tpcli {
	return ui
}

// Start is
func (ui *Tpcli) Start() {
	// var generalOutputPanel UIOutputPanel
	// var errorOutputPanel UIOutputPanel
	// var inputPanel UICommandEntryPanel

	uiFramework := newDefaultUIFramework()

	channelOfUserEnteredCommands := make(chan string, 10)

	switch ui.stackingOrder {
	case CommandErrorGeneral:
		_ = uiFramework.AddTheCommandHistoryPanel()
		_ = uiFramework.AddTheErrorPanel()
		_ = uiFramework.AddTheCommandInputPanel(channelOfUserEnteredCommands)
	}
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
