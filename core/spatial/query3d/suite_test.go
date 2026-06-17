package query3d_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQuery3D(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query3D Suite")
}
