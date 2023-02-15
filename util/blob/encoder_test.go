package blob_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/util/blob"
)

func BenchmarkReflectEncoder(b *testing.B) {
	buffer := new(bytes.Buffer)
	encoder := blob.NewReflectEncoder(buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(uint8(0x12))
		encoder.Encode(uint16(0x24))
		encoder.Encode(uint32(0x48))
	}
}

func BenchmarkGobEncoder(b *testing.B) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(uint8(0x12))
		encoder.Encode(uint16(0x24))
		encoder.Encode(uint32(0x48))
	}
}

var _ = FDescribe("Encoder", func() {
	var (
		encoder *blob.ReflectEncoder
		buffer  *bytes.Buffer
	)

	stream := func(values ...uint8) []uint8 {
		return values
	}

	BeforeEach(func() {
		buffer = new(bytes.Buffer)
		encoder = blob.NewReflectEncoder(buffer)
	})

	DescribeTable("types",
		func(data any, expected []byte) {
			Expect(encoder.Encode(data)).To(Succeed())
			Expect(buffer.Bytes()).To(Equal(expected))
		},
		Entry("uint8", uint8(0x13), stream(0x13)),
		Entry("*uint8", gog.PtrOf(uint8(0x13)), stream(0x13)),
		Entry("int8", int8(0x13), stream(0x13)),
		Entry("*int8", gog.PtrOf(int8(0x13)), stream(0x13)),
		Entry("int16", int16(0x31CA), stream(0xCA, 0x31)),
		Entry("*int16", gog.PtrOf(int16(0x31CA)), stream(0xCA, 0x31)),
		Entry("uint16", uint16(0xF1CA), stream(0xCA, 0xF1)),
		Entry("*uint16", gog.PtrOf(uint16(0xF1CA)), stream(0xCA, 0xF1)),
		Entry("int32", int32(0x31CA7632), stream(0x32, 0x76, 0xCA, 0x31)),
		Entry("*int32", gog.PtrOf(int32(0x31CA7632)), stream(0x32, 0x76, 0xCA, 0x31)),
		Entry("uint32", uint32(0xF1CA7632), stream(0x32, 0x76, 0xCA, 0xF1)),
		Entry("*uint32", gog.PtrOf(uint32(0xF1CA7632)), stream(0x32, 0x76, 0xCA, 0xF1)),
	)
})
