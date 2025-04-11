package lsl_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLSL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LSL Suite")
}

func openTestFile(segments ...string) string {
	defer GinkgoRecover()
	segments = append([]string{"testdata"}, segments...)
	data, err := os.ReadFile(filepath.Join(segments...))
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
