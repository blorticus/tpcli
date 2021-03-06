package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	mainApplication := &application{}

	cliArgumentsProcessor, err := ProcessCliArguments()
	mainApplication.dieIfError(err)

	if cliArgumentsProcessor.WantsToLogToDebugFile() {
		mainApplication.activateDebugLoggingUsingFile(cliArgumentsProcessor.DebugLogFileFullPath())
	} else {
		mainApplication.deactivateDebugLogging()
	}
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type application struct {
	debugLogger *log.Logger
}

func (app *application) die(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

func (app *application) dieIfError(err error) {
	if err != nil {
		app.die(err.Error())
	}
}

func (app *application) activateDebugLoggingUsingFile(fileName string) {
	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0640)
	panicIfError(err)

	app.debugLogger = log.New(fileHandle, "", 0)
}

func (app *application) deactivateDebugLogging() {
	app.debugLogger = log.New(ioutil.Discard, "", 0)
}

func (app *application) setPanelStackingOrderToUserSpecifiedValue(order string) {

}

func (app *application) setPanelStackingOrderToDefault() {

}

func (app *application) setEhPanelChoiceToUserSpecifiedValue(value string) {

}

func (app *application) setEhPanelChoiceToDefault() {

}

func (app *application) bindCommandSocket(bindType string, bindValue string) {

}
