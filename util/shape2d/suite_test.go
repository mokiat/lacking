package shape2d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShape2D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shape 2D Suite")
}
