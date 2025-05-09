package chunked_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/chunked"
)

var _ = Describe("Object", Ordered, func() {
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

	BeforeAll(func() {
		storage = chunked.NewMemoryStorage()
		asset = chunked.NewAsset(storage, "example.laf")
	})

	It("is possible to encode asset", func() {
		model := EncodeModel{
			IDChunk:       &IDChunk{Name: "test"},
			LocationChunk: &LocationChunk{X: 1, Y: 2},
		}
		Expect(asset.Write(model)).To(Succeed())
	})

	It("is possible to decode asset", func() {
		var model DecodeModel
		Expect(asset.Read(&model)).To(Succeed())
		Expect(model.LocationChunk).To(Equal(&LocationChunk{X: 1, Y: 2}))
		Expect(model.PriorityChunk).To(BeNil())
	})
})
