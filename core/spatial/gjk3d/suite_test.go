package gjk3d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGJK3D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GJK 3D Suite")
}
