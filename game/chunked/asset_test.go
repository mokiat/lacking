package chunked_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/chunked"
)

var _ = Describe("Asset", func() {
	type EncodeModel struct {
		*IDChunk
		*LocationChunk
	}

	type DecodeModel struct {
		*LocationChunk
		*PriorityChunk
	}

	var (
		storage chunked.Storage
		asset   *chunked.Asset
	)

	BeforeEach(func() {
		storage = chunked.NewMemoryStorage()
		asset = chunked.NewAsset(storage, "example.dat")
	})

	It("is possible to encode a struct with nil chunks", func() {
		output := EncodeModel{}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a struct with nil chunks", func() {
		output := EncodeModel{}
		Expect(asset.Write(output)).To(Succeed())

		var input EncodeModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.IDChunk).To(BeNil())
		Expect(input.LocationChunk).To(BeNil())
	})

	It("is possible to encode a struct with chunks", func() {
		output := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a struct with chunks", func() {
		output := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input EncodeModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.IDChunk).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to decode a struct with chunks into a mismatching struct", func() {
		output := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input DecodeModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
		Expect(input.PriorityChunk).To(BeNil())
	})

	It("is possible to decode a struct with chunks into a chunk consumer", func() {
		output := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input chunked.BaseChunkHolder
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.Items).To(HaveLen(2))
		for _, item := range input.Items {
			Expect(item).To(BeAssignableToTypeOf(&chunked.RawChunk{}))
		}
	})

	It("is possible to go through raw chunks", func() {
		output := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		type Wrapper struct {
			chunked.BaseChunkHolder // nested
		}
		var input struct {
			Wrapper // nested further
		}
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.Items).To(HaveLen(2))
		for _, item := range input.Items {
			Expect(item).To(BeAssignableToTypeOf(&chunked.RawChunk{}))
		}

		Expect(asset.Write(input)).To(Succeed())

		var output2 EncodeModel
		Expect(asset.Read(&output2)).To(Succeed())
		Expect(output.IDChunk).To(Equal(&IDChunk{Name: "test"}))
		Expect(output.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to encode a non-chunked value", func() {
		output := "example"
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to encode a chunk provider", func() {
		output := chunked.ChunkList{
			&IDChunk{Name: "test"},
			&LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a chunk provider into a struct model", func() {
		output := chunked.ChunkList{
			&IDChunk{Name: "test"},
			&LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input EncodeModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.IDChunk).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to encode a chunk directly", func() {
		output := &IDChunk{Name: "test"}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a chunk directly", func() {
		output := &IDChunk{Name: "test"}
		Expect(asset.Write(output)).To(Succeed())

		var input IDChunk
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input).To(Equal(IDChunk{Name: "test"}))
	})
})
