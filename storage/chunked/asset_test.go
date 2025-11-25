package chunked_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/storage/chunked"
)

var _ = Describe("Asset", func() {
	type PrimaryModel struct {
		ID       *IDChunk       `chunk:"id"`
		Location *LocationChunk `chunk:"location"`
	}

	type SecondaryModel struct {
		Loc  *LocationChunk `chunk:"location"`
		Prio *PriorityChunk `chunk:"priority"`
	}

	type NestedModel struct {
		PrimaryModel
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
		var output PrimaryModel
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a struct with nil chunks", func() {
		var output PrimaryModel
		Expect(asset.Write(output)).To(Succeed())

		var input PrimaryModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.ID).To(BeNil())
		Expect(input.Location).To(BeNil())
	})

	It("is possible to encode a struct with chunks", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a struct with chunks", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input PrimaryModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.ID).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.Location).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to decode a struct with chunks into a mismatching struct", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input SecondaryModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.Loc).To(Equal(&LocationChunk{X: 1, Y: 2}))
		Expect(input.Prio).To(BeNil())
	})

	It("is possible to encode a nested struct with chunks", func() {
		output := NestedModel{
			PrimaryModel{
				ID:       &IDChunk{Name: "test"},
				Location: &LocationChunk{X: 1, Y: 2},
			},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input PrimaryModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.ID).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.Location).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to decode a struct with chunks into a nested struct", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input NestedModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.ID).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.Location).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to encode a chunk provider", func() {
		output := chunked.ChunkList{
			chunked.FromValue("id", IDChunk{Name: "test"}),
			chunked.FromValue("location", LocationChunk{X: 1, Y: 2}),
		}
		Expect(asset.Write(output)).To(Succeed())
	})

	It("is possible to decode a chunk provider into a struct model", func() {
		output := chunked.ChunkList{
			chunked.FromValue("id", IDChunk{Name: "test"}),
			chunked.FromValue("location", LocationChunk{X: 1, Y: 2}),
		}
		Expect(asset.Write(output)).To(Succeed())

		var input PrimaryModel
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.ID).To(Equal(&IDChunk{Name: "test"}))
		Expect(input.Location).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})

	It("is possible to decode a struct with chunks into a chunk consumer", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		var input chunked.ChunkHolder
		Expect(asset.Read(&input)).To(Succeed())
		Expect(input.Items).To(HaveLen(2))
		for _, item := range input.Items {
			Expect(item).To(BeAssignableToTypeOf(chunked.RawChunk{}))
		}
	})

	It("is possible to go through raw chunks", func() {
		output := PrimaryModel{
			ID:       &IDChunk{Name: "test"},
			Location: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(output)).To(Succeed())

		type Wrapper struct {
			chunked.ChunkHolder // nested
		}
		var input struct {
			Wrapper // nested further
		}
		Expect(asset.Read(&input)).To(Succeed())
		Expect(asset.Write(input)).To(Succeed())

		var output2 PrimaryModel
		Expect(asset.Read(&output2)).To(Succeed())
		Expect(output.ID).To(Equal(&IDChunk{Name: "test"}))
		Expect(output.Location).To(Equal(&LocationChunk{X: 1, Y: 2}))
	})
})
