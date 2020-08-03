package main

import (
	"github.com/blorticus/tpcli"
)

func main() {
	ui := tpcli.NewTpcli()
	//channelOfCommandsFromUI := ui.ChannelOfControlMessagesFromTheUI()

	ui.Start()
}
