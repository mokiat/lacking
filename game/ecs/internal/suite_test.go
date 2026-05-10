package internal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestECSInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ECS Internal Suite")
}
