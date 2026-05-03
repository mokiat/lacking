package ecs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV5(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V5 Suite")
}
