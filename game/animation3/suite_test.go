package animation3_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAnimation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Animation Suite")
}
