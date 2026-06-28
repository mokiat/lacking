package placement2d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlacement2D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Placement2D Suite")
}
