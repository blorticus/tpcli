package tpcli_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/blorticus/tpcli"
)

// A general example for CommandProcessor
func ExampleCommandProcessor() {
	var ui *tpcli.Tpcli

	cp := tpcli.NewCommandProcessor().
		WhenCommandMatches(`add (\d+) to (\d+)`, func(matchGroups []string) error {
			x, err := strconv.Atoi(matchGroups[1])
			if err != nil {
				return err
			}

			y, err := strconv.Atoi(matchGroups[2])
			if err != nil {
				return nil
			}

			ui.AddStringToGeneralOutput(fmt.Sprintf("Sum is: %d", x+y))
			return nil
		}).
		WhenCommandMatches(regexp.MustCompile(`^quit$`), func(matchGroups []string) error {
			os.Exit(0)
			return nil
		})

	ui.Start()

	for {
		nextCommand := <-ui.ChannelOfEnteredCommands()
		commandMatchedSomething, errorReturnedByCallback := cp.ProcessCommandString(nextCommand)

		if !commandMatchedSomething {
			ui.AddStringToErrorOutput("Command not understood")
		}

		if errorReturnedByCallback != nil {
			ui.AddStringToErrorOutput(errorReturnedByCallback.Error())
		}
	}
}
