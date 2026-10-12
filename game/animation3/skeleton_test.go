package animation3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("Skeleton", func() {

	var skeleton *animation3.Skeleton

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]string{"hip", "spine", "head"})
	})

	Describe("BoneCount", func() {

		It("returns the number of bones", func() {
			Expect(skeleton.BoneCount()).To(Equal(3))
		})
	})

	Describe("BoneName", func() {

		DescribeTable("returns the name of the bone at a valid index",
			func(index int, expected string) {
				Expect(skeleton.BoneName(index)).To(Equal(expected))
			},
			Entry("first bone", 0, "hip"),
			Entry("middle bone", 1, "spine"),
			Entry("last bone", 2, "head"),
		)

		DescribeTable("returns an empty string for an out-of-range index",
			func(index int) {
				Expect(skeleton.BoneName(index)).To(BeEmpty())
			},
			Entry("negative index", -1),
			Entry("index equal to the bone count", 3),
			Entry("index far past the end", 100),
		)
	})

	Describe("BoneIndex", func() {

		DescribeTable("returns the index of a known bone",
			func(name string, expected int) {
				index, ok := skeleton.BoneIndex(name)
				Expect(ok).To(BeTrue())
				Expect(index).To(Equal(expected))
			},
			Entry("first bone", "hip", 0),
			Entry("middle bone", "spine", 1),
			Entry("last bone", "head", 2),
		)

		It("reports an unknown bone", func() {
			index, ok := skeleton.BoneIndex("tail")
			Expect(ok).To(BeFalse())
			Expect(index).To(Equal(0))
		})

		It("is case sensitive", func() {
			_, ok := skeleton.BoneIndex("Hip")
			Expect(ok).To(BeFalse())
		})
	})

	It("maps names and indices consistently", func() {
		for index := range skeleton.BoneCount() {
			name := skeleton.BoneName(index)
			actualIndex, ok := skeleton.BoneIndex(name)
			Expect(ok).To(BeTrue())
			Expect(actualIndex).To(Equal(index))
		}
	})

	When("created with duplicate names", func() {

		BeforeEach(func() {
			skeleton = animation3.NewSkeleton([]string{"hip", "spine", "hip", "head", "spine"})
		})

		It("counts each name once", func() {
			Expect(skeleton.BoneCount()).To(Equal(3))
		})

		It("assigns consecutive indices in order of first occurrence", func() {
			Expect(skeleton.BoneName(0)).To(Equal("hip"))
			Expect(skeleton.BoneName(1)).To(Equal("spine"))
			Expect(skeleton.BoneName(2)).To(Equal("head"))

			index, ok := skeleton.BoneIndex("head")
			Expect(ok).To(BeTrue())
			Expect(index).To(Equal(2))
		})
	})

	When("created with no names", func() {

		DescribeTable("is empty",
			func(names []string) {
				empty := animation3.NewSkeleton(names)
				Expect(empty.BoneCount()).To(Equal(0))
				Expect(empty.BoneName(0)).To(BeEmpty())
				_, ok := empty.BoneIndex("hip")
				Expect(ok).To(BeFalse())
			},
			Entry("nil slice", nil),
			Entry("empty slice", []string{}),
		)
	})

	It("is not affected by later changes to the input slice", func() {
		names := []string{"hip", "spine"}
		s := animation3.NewSkeleton(names)
		names[0] = "changed"

		Expect(s.BoneName(0)).To(Equal("hip"))
		index, ok := s.BoneIndex("hip")
		Expect(ok).To(BeTrue())
		Expect(index).To(Equal(0))
		_, ok = s.BoneIndex("changed")
		Expect(ok).To(BeFalse())
	})

})
