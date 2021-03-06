// Package tpcli is a golang package (and application) that provides a simple three panel terminal-based UI.
//
// Overview
//
// tpcli is the "Three Panel Command-line Interface".  It is both a golang package and an
// application.  This page documents the package.
//
// The UI consists of three "panels": a command
// entry panel, a general output panel and third panel that is either for error
// output or which records the history of entered commands.
//
// The command entry panel starts with focus.  It is a single row panel which supports
// basic shell-emacs bindings (e.g., ^a to go to the start of the line, ^e to the end
// of the line) and arrow key readline-style history navigation.
//
// The UI runs is started as a goroutine.  When the user enters a string in the command entry panel and hits
// <enter>, the command string is delivered on a channel.  Text may be written to the
// general output panel.  The third panel serves one of two functions.  By default, it is an
// error output panel.  It is just like the general output panel, and differs only semantically.
// It may alternatively be set to command history panel.  In this case, every time the user
// enters a command string, it is appended to this panel, providing a command history.
//
// The user may use <tab> to switch between the panels.  Only the command input panel will
// accept input.  If either of the other two panels has focus, the arrow keys may be used to
// scroll up or down through the text output.
//
// The panels may be stacked in any order desired.  The default order places the output panel
// first, then the error output panel, then the command entry panel.
//
// If the user hits <esc> or <ctrl>-q, the UI exits.  This mean it Stop()s, and an additional
// function is called.  By default, that function is os.Exit(0).  However, this may be overridden
// via OnUIExit().
//
// For convenience, there is also a CommandProcessor that allows you to define patterns for possible
// commands, and associate those will callback methods when the user enters those commands.
//
// Example
//
//  ui := tpcli.NewUI().ChangeStackingOrderTo(tpcli.CommandGeneralError)
//  go ui.Start()
//  for {
//      nextCommand := <-ui.ChannelOfEnteredCommands()
//          switch nextCommand {
//              "quit":
//                  ui.Stop()
//              "time":
//                  hour, min, sec := time.Now().Clock()
//                  ui.FmtToGeneralOutput("The time of day is: %02d:%02d:%02d", hour, min, sec)
//              default:
//                  ui.AddStringToErrorOutput("You can only ask for the time, I'm afraid.")
//      }
//  }
//
package tpcli
