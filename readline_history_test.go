package tpcli_test

import (
	//"fmt"
	//"regexp"

	"github.com/blorticus/tpcli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("readline_history", func() {
	var readlineHistory *tpcli.ReadlineHistory

	BeforeEach(func() {
		readlineHistory = tpcli.NewReadlineHistory(10)
	})

	Context("Empty readline", func() {
		It("should return empty string and be at top of list on Up()", func() {
			item := readlineHistory.Up()
			Expect(item).To(Equal(""))
		})

		It("should return empty string and be at top of list on Down()", func() {
			item := readlineHistory.Down()
			Expect(item).To(Equal(""))
		})
	})

	Context("readline with three entries", func() {
		JustBeforeEach(func() {
			readlineHistory.AddItem("first item")
			readlineHistory.AddItem("second item")
			readlineHistory.AddItem("third item")
			readlineHistory.ResetIteration()
		})

		It("should work properly moving only Up", func() {
			By("moving up first time")
			item := readlineHistory.Up()
			Expect(item).To(Equal("third item"))

			By("moving up second time")
			item = readlineHistory.Up()
			Expect(item).To(Equal("second item"))

			By("moving up third time")
			item = readlineHistory.Up()
			Expect(item).To(Equal("first item"))

			By("moving up fourth time")
			item = readlineHistory.Up()
			Expect(item).To(Equal("first item"))

			By("moving up fifth time")
			item = readlineHistory.Up()
			Expect(item).To(Equal("first item"))

		})

		It("should work properly moving all the way Up then all the way down", func() {
			By("moving up three times")
			readlineHistory.Up()
			readlineHistory.Up()
			item := readlineHistory.Up()
			Expect(item).To(Equal("first item"))

			By("Moving down the first time from the top")
			item = readlineHistory.Down()
			Expect(item).To(Equal("second item"))

			By("Moving down the second time")
			item = readlineHistory.Down()
			Expect(item).To(Equal("third item"))

			By("Moving down the third time")
			item = readlineHistory.Down()
			Expect(item).To(Equal(""))

			By("Moving down the fourth time")
			item = readlineHistory.Down()
			Expect(item).To(Equal(""))
		})
	})
})
