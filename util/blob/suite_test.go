package blob_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBlob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blob Suite")
}
