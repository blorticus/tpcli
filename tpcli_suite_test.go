package tpcli_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTpcli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tpcli Suite")
}
