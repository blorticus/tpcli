package tpcli_test

import (
	"fmt"
	"regexp"

	"github.com/blorticus/tpcli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type commandProcessorCallbacks struct {
	nameOfLastCallback           string
	matchGroupsFromLastCallbacks []string
	throwThisErrorOnNextCall     error
}

func (c *commandProcessorCallbacks) Reset() *commandProcessorCallbacks {
	c.nameOfLastCallback = ""
	c.matchGroupsFromLastCallbacks = []string{}
	c.throwThisErrorOnNextCall = nil
	return c
}

func (c *commandProcessorCallbacks) induceErrorOnNextCallback(err error) {
	c.throwThisErrorOnNextCall = err
}

func (c *commandProcessorCallbacks) genericOn(nameOfCallback string, matchGroups []string) error {
	c.nameOfLastCallback = nameOfCallback
	c.matchGroupsFromLastCallbacks = matchGroups
	return c.throwThisErrorOnNextCall
}

func (c *commandProcessorCallbacks) OnQuit(matchGroups []string) error {
	return c.genericOn("OnQuit", matchGroups)
}

func (c *commandProcessorCallbacks) OnRead(matchGroups []string) error {
	return c.genericOn("OnRead", matchGroups)
}

func (c *commandProcessorCallbacks) OnCompile(matchGroups []string) error {
	return c.genericOn("OnCompile", matchGroups)
}

var _ = Describe("CommandProcessor", func() {
	var (
		processor *tpcli.CommandProcessor
		callback  *commandProcessorCallbacks
	)

	BeforeEach(func() {
		callback = &commandProcessorCallbacks{}
		processor = tpcli.NewCommandProcessor().
			WhenCommandMatches(`^quit$`, callback.OnQuit).
			WhenCommandMatches(`^read (\S+) from file (.+)$`, callback.OnRead).
			WhenCommandMatches(regexp.MustCompile(`^(\S+) must compile (\d+) times$`), callback.OnCompile)
	})

	JustBeforeEach(func() {
		callback.Reset()
	})

	Describe("Testing Three Callbacks", func() {
		Context("command: quit", func() {
			It("should match, generate no error, and have single element matchGroups", func() {
				matchesAnyMatcher, err := processor.ProcessCommandString("quit")
				Expect(matchesAnyMatcher).To(BeTrue())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(callback.nameOfLastCallback).To(Equal("OnQuit"))
				Expect(callback.matchGroupsFromLastCallbacks).To(Equal([]string{"quit"}))
			})
		})

		Context("command: read variant 1 without error", func() {
			It("should match, generate no error, and have correct matchGroups", func() {
				matchesAnyMatcher, err := processor.ProcessCommandString("read something from file /foo/bar/baz.txt")
				Expect(matchesAnyMatcher).To(BeTrue())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(callback.nameOfLastCallback).To(Equal("OnRead"))
				Expect(callback.matchGroupsFromLastCallbacks).To(Equal([]string{"read something from file /foo/bar/baz.txt", "something", "/foo/bar/baz.txt"}))
			})
		})

		Context("command: read variant 2 with error", func() {
			It("should match, generate an error, and have correct matchGroups", func() {
				callback.induceErrorOnNextCallback(fmt.Errorf("this is an error"))
				matchesAnyMatcher, err := processor.ProcessCommandString("read something from file /foo/bar bat/baz.txt")
				Expect(matchesAnyMatcher).To(BeTrue())
				Expect(err).Should(HaveOccurred())
				Expect(callback.nameOfLastCallback).To(Equal("OnRead"))
				Expect(callback.matchGroupsFromLastCallbacks).To(Equal([]string{"read something from file /foo/bar bat/baz.txt", "something", "/foo/bar bat/baz.txt"}))
			})
		})

		Context("command: complete", func() {
			It("should match, generate no error, and have correct matchGroups", func() {
				matchesAnyMatcher, err := processor.ProcessCommandString("foo must compile 200 times")
				Expect(matchesAnyMatcher).To(BeTrue())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(callback.nameOfLastCallback).To(Equal("OnCompile"))
				Expect(callback.matchGroupsFromLastCallbacks).To(Equal([]string{"foo must compile 200 times", "foo", "200"}))
			})
		})

		Context("non-matching command similar to quit", func() {
			It("should not match, generate the error 'Command not understood', and have called no callback", func() {
				matchesAnyMatcher, err := processor.ProcessCommandString("quit ")
				Expect(matchesAnyMatcher).NotTo(BeTrue())
				Expect(err).Should(MatchError("command not understood"))
				Expect(callback.nameOfLastCallback).To(Equal(""))
			})
		})

		Context("non-matching command similar to compile", func() {
			It("should not match, generate the error 'Command not understood', and have called no callback", func() {
				matchesAnyMatcher, err := processor.ProcessCommandString("foo must compile three times")
				Expect(matchesAnyMatcher).NotTo(BeTrue())
				Expect(err).Should(MatchError("command not understood"))
				Expect(callback.nameOfLastCallback).To(Equal(""))
			})
		})
	})
})
