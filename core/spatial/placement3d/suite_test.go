package placement3d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlacement3D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Placement3D Suite")
}
