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

	When("an asset is encoded", func() {
		BeforeEach(func() {
			model := EncodeModel{
				IDChunk:       &IDChunk{Name: "test"},
				LocationChunk: &LocationChunk{X: 1, Y: 2},
			}
			Expect(asset.Write(model)).To(Succeed())
		})

		It("is possible to decode it into a different model", func() {
			var model DecodeModel
			Expect(asset.Read(&model)).To(Succeed())
			Expect(model.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
			Expect(model.PriorityChunk).To(BeNil())
		})
	})

	When("an asset with nil chunks is encoded", func() {
		BeforeEach(func() {
			model := EncodeModel{}
			Expect(asset.Write(model)).To(Succeed())
		})

		It("produces an empty asset", func() {
			var model EncodeModel
			Expect(asset.Read(&model)).To(Succeed())
			Expect(model.IDChunk).To(BeNil())
			Expect(model.LocationChunk).To(BeNil())
		})
	})

})
