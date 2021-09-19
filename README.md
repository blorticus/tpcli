# Brief

tpcli is the Three-Panel Command-Line UI.  It is both a golang module and an application using that module that can communicate with external applications.

## The UI

The UI is terminal-based, and presents three panels stacked one atop the other.  The three panels include: general output, error output/command-history and command input.  The command input panel is a single row, and supports both bash-like keybinding (e.g., ^a to go to the beginning of a line, ^e to go to the end, ^k to remove from the cursor and beyond) and command-history scrolling with up- and down-arrow keys.  The error output panel can either be used to display the recent command history (that is, as commands are entered into the command input panel, they appear in a scrolling list in this panel), or it can be used for error output (actually, since the UI itself has no notion of what an "error" is, it really is just another output display).  The general output is used for general messages.  The panels can be arranged in any order desired, and the error panel is optional.  The only selectable panel is the command input panel.  ^Q or escape will cause the UI to exit (presumably returning to a shell).

## As a golang Module

First, the tpcli is constructed, then started in a goroutine.  The UI goroutine will send both errors and user-inputed command strings over a channel.  The contents of the command input panel can be changed from the connecting application, and additional text can be added to either the ouptut panel and the error panel.  If the error panel is set to a command history, any output sent to the error panel is redirected to the output panel instead.

```golang
package main

import (
	"time"

	"github.com/blorticus/tpcli"
)

func main() {
	ui := tpcli.NewUI().ChangeStackingOrderTo(tpcli.CommandGeneralError)
	go ui.Start()

	for {
		switch <-ui.ChannelOfEnteredCommands() {
		case "quit":
			ui.Stop()
		case "time":
			hour, min, sec := time.Now().Clock()
			ui.FmtToGeneralOutput("The time of day is: %02d:%02d:%02d", hour, min, sec)
		default:
			ui.AddStringToErrorOutput("You can only ask for the time, I'm afraid.")
		}

	}
}
```

Note that `^q` and the escape key will both cause the UI to exit.

## As an Application

If the three-panel CLI is run as an application, it will bind to and listen on either a Unix (SOCK_STREAM) socket or a TCP socket.  Messages are delivered over this socket.  Messages sent from the application are commands that have been fully input (that is, some text was entered in the command input panel, and the user hit enter).  Messages to the application are output to general output or the error ouput box (if the box isn't a command-history).  A message is JSON encoded, as follows:

```json
{
    "type": "$type",
    "message": "$message"
}
```

$type must be one of the following:

```html
 protocol_error
 input_command_received
 input_command_replacement
 general_output
 error_output
 user_exited
```

The application will emit "protocol_error" and "input_command_received" messages, and will receive "input_command_replacement", "general_output" and "error_output".  It will silently ignore any non-supported message type value and any message received that is intended only for output (i.e., "protocol_error" and "input_command_received").

A protocol_error is a general error message for the application's peer.  The $message contains a text string for the error.

An input_command_received is a command that the user entered (including, possibly, the empty command).  The $message is the command value.  It excludes the trailing newline.

An input_command_replacement is text that should be placed in the command panel.  The $message is the command, which will be interpretted as UTF-8.  Any non-printable characters are ignored.  A protocol_error is raised if it contains a newline and that newline is not the last character.

A general_output is UTF-8 text that is appended to the general output panel.  A newline is appended to any existing text, then the new $message text is added.  Newlines are permitted.  Other non-printable characters are ignored.

An error_output is text that is appended to the error box.  If the application is configured to use command history in that box, the message is delivered to the general output panel instead.

The application is invoked thusly:

```bash
tpcli <bind> [-order <panel_order>] [-debug <debug_file_path>]
```

where `<bind>` is either `-unix <path/to/socket>` or `-tcp <ip>:<port>`; `<panel_order>` is the order in which the panels are stacked.  The default bind is `-tcp localhost:6000`.  The `<panel_order>` is a three letter sequence, with `c` representing the command entry panel, `h` representing the command-history panel, `e` representing the error panel, and `o` representing the output panel.  Thus, if one wishes to place the output panel first, then the history panel, then the command entry panel, one would provide `-order ohc`.  `ohc` is the default.  Only one of `h` or `e` can be provided, and each of the three letters must be unique (that is, a single panel type cannot be applied twice).

Messages as described above flow on the specified bound socket.
