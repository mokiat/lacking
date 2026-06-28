package isec2d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestISec2D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Intersection 2D Suite")
}
