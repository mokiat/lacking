package lsl_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLSL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LSL Suite")
}
