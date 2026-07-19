package physics_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPhysics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Physics Suite")
}
