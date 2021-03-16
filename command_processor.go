package tpcli

import (
	"fmt"
	"regexp"
)

type matcher struct {
	pattern  *regexp.Regexp
	callback func([]string) error
}

// CommandProcessor is a helper for processing user commands input in the Tpcli command entry panel.
// It is supplied regular expression matchers, and if any of them match a user string, an associated
// callback function is invoked.  If none of the supplied matchers match the command string, this
// is indicated.
type CommandProcessor struct {
	matchersInOrderProvided []*matcher
}

// NewCommandProcessor creates a new, empty command processor
func NewCommandProcessor() *CommandProcessor {
	return &CommandProcessor{
		matchersInOrderProvided: make([]*matcher, 0, 20),
	}
}

// WhenCommandMatches adds a matcher with its callback.  'pattern' may be either a string or a *regexp.Regexp.  If it is
// a string, then it is fed to regexp.MustCompile (and will panic if the compilation fails).  If it is neither type,
// this method panics
func (processor *CommandProcessor) WhenCommandMatches(pattern interface{}, doCallback func([]string) error) *CommandProcessor {
	if _, patternIsARegexp := pattern.(*regexp.Regexp); patternIsARegexp {
		processor.matchersInOrderProvided = append(processor.matchersInOrderProvided, &matcher{
			pattern:  pattern.(*regexp.Regexp),
			callback: doCallback,
		})
	} else if _, patternIsAString := pattern.(string); patternIsAString {
		processor.matchersInOrderProvided = append(processor.matchersInOrderProvided, &matcher{
			pattern:  regexp.MustCompile(pattern.(string)),
			callback: doCallback,
		})
	} else {
		panic("WhenCommandMatches invoked with a pattern that is neither a string nor a *regexp.Regexp")
	}

	return processor
}

// ProcessCommandString accepts a commandString and matches it against all matchers previously supplied to this
// ComamndProcessor, in the order that they were provided.  On the first match, the asssociated callback is
// invoked and this method returns true and the error from the callback.  If no pre-defined matchers match, then
// this method returns false and the string "Command not understood"
func (processor *CommandProcessor) ProcessCommandString(commandString string) (matchesAnyDefinedPattern bool, errorFromCallback error) {
	for _, matcher := range processor.matchersInOrderProvided {
		if matchGroups := matcher.pattern.FindStringSubmatch(commandString); len(matchGroups) > 0 {
			err := matcher.callback(matchGroups)
			return true, err
		}
	}

	return false, fmt.Errorf("command not understood")
}
