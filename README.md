# Brief

tpcli is the Three-Panel Command-Line UI.  It is both a golang module and an application using that module that can communicate with external applications.

# The UI

The UI is terminal-based, and presents three panels stacked one atop the other.  The three panels include: general output, error output/command-history and command input.  The command input panel is a single row, and supports both bash-like keybinding (e.g., ^a to go to the beginning of a line, ^e to go to the end, ^k to remove from the cursor and beyond) and command-history scrolling with up- and down-arrow keys.  The error output panel can either be used to display the recent command history (that is, as commands are entered into the command input panel, they appear in a scrolling list in this panel), or it can be used for error output (actually, since the UI itself has no notion of what an "error" is, it really is just another output display).  The general output is used for general messages.  The panels can be arranged in any order desired, and the error panel is optional.  The only selectable panel is the command input panel.  ^Q or escape will cause the UI to exit (presumably returning to a shell).

# As a golang Module

# As an Application

If the three-panel CLI is run as an application, it will bind to and listen on either a Unix (SOCK_STREAM) socket or a TCP socket.  Messages are delivered over this socket.  Messages sent from the application are commands that have been fully input (that is, some text was entered in the command input panel, and the user hit enter).  Messages to the application are output to general output or the error ouput box (if the box isn't a command-history).  A message has the following format (network byte order):

```
    16 bits       8 bits         variable
  +------------+--------------+--------------+
  | MSG LENGTH | MSG TYPE     | BODY         |
  +------------+--------------+--------------+
```

The message length includes the header (i.e., the MSG LENGTH field and the MSG TYPE field).  It is a count of octets in the message.  It must, therefore, be at least 3.

The following MSG TYPEs are defined:

```
 1 - protocol_error
 2 - input_command_received
 3 - input_command_replacement
 4 - general_output
 5 - error_output
```

The application will emit type 1 and type 2 messages, and will receive type 3, 4 and 5.  It will silently ignore any non-supported message type value (e.g., a MSG TYPE value of 10), a message of invalid length (less than 3), and any message received that is intended only for output (i.e., type 1 and 2 messages).

A protocol_error is a general error message for the application's peer.  The BODY contains a text string for the error.

An input_command_received is a command that the user entered (including, possibly, the empty command).  The body is the command value.  It excludes the trailing newline.

An input_command_replacement is text that should be placed in the command panel.  The body is the command, which will be interpretted as UTF-8.  Any non-printable characters are ignored.  A protocol_error is raised if it contains a newline and that newline is not the last character.

A general_output is UTF-8 text that is appended to the general output panel.  A newline is appended to any existing text, then the new BODY text is added.  Newlines are permitted.  Other non-printable characters are ignored.

An error_output is text that is appended to the error box.  If the application is configured to use command history in that box, the message is delivered to the general output panel instead.

The application is invoked thusly:

```
tpcli <bind> [-o|-order <panel_order>] [-ehpanel error|history]
```

where `<bind>` is either `-unix <path/to/socket>` or `-tcp <ip>:<port>`; `<panel_order>` is the order in which the panels are stacked; and `-ehpanel` indicates whether the error/command-history panel should be used for error output for command-history output.

The `<panel_order>` is a three letter sequence, with `c` representing the command entry panel, `h` representing the error/command-history panel, and `o` representing the output panel.  Thus, if one wishes to place the output panel first, then the history panel, then the command entry panel, one would provide `-order ohc`.  `ohc` is the default.  For `-ehpanel`, `error` is the default.
