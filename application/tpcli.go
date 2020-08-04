package main

import (
	"github.com/blorticus/tpcli"
)

func main() {
	ui := tpcli.NewUI()

	ui.Start()

	commandInputTextChannel := ui.ChannelOfEnteredCommands()

	for {
		command := <-commandInputTextChannel
		ui.AddStringToGeneralOutput("Command: " + command)
	}
}
