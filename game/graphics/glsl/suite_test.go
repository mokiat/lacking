package glsl_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGLSL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GLSL Suite")
}

func openTestFile(segments ...string) string {
	defer GinkgoRecover()
	segments = append([]string{"testdata"}, segments...)
	data, err := os.ReadFile(filepath.Join(segments...))
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
