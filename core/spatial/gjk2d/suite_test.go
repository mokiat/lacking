package gjk2d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGJK2D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GJK 2D Suite")
}
