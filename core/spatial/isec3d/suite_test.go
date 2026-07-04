package isec3d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestISec3D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Intersection 3D Suite")
}
