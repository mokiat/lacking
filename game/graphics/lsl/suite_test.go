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

func openTestFile(folder, name string) string {
	defer GinkgoRecover()
	data, err := os.ReadFile(filepath.Join("testdata", folder, name))
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
