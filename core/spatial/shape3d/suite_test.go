package shape3d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShape3D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shape3D Suite")
}
