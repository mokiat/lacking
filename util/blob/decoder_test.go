package blob_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/util/blob"
)

var _ = Describe("Decoder", func() {
	var (
		decoder *blob.ReflectDecoder
		buffer  *bytes.Buffer

		inTarget any
	)

	BeforeEach(func() {
		buffer = new(bytes.Buffer)
		decoder = blob.NewReflectDecoder(buffer)
	})

	JustBeforeEach(func() {
		Expect(decoder.Decode(inTarget)).To(Succeed())
	})

	When("uint8", func() {
		var target uint8

		BeforeEach(func() {
			inTarget = &target
			buffer.WriteByte(123)
		})

		It("decodes the data", func() {
			Expect(target).To(Equal(uint8(123)))
		})
	})
})
