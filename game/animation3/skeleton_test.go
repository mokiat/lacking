package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("Skeleton", func() {

	// restTransform returns a rest transform that is unique for the specified
	// seed, so that tests can tell the rest transforms of bones apart.
	restTransform := func(seed float64) animation3.Transform {
		return animation3.Transform{
			Translation: dprec.NewVec3(seed, 0.0, 0.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(seed), dprec.BasisYVec3()),
			Scale:       dprec.NewVec3(1.0, seed, 1.0),
		}
	}

	var (
		hipRest   animation3.Transform
		spineRest animation3.Transform
		headRest  animation3.Transform
		skeleton  *animation3.Skeleton
	)

	BeforeEach(func() {
		hipRest = restTransform(1.0)
		spineRest = restTransform(2.0)
		headRest = restTransform(3.0)
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "hip", RestTransform: hipRest},
			{Name: "spine", RestTransform: spineRest},
			{Name: "head", RestTransform: headRest},
		})
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

	Describe("BoneRestTransform", func() {

		It("returns the rest transform of the bone at a valid index", func() {
			Expect(skeleton.BoneRestTransform(0)).To(Equal(hipRest))
			Expect(skeleton.BoneRestTransform(1)).To(Equal(spineRest))
			Expect(skeleton.BoneRestTransform(2)).To(Equal(headRest))
		})

		It("returns the zero transform for a bone created without one", func() {
			s := animation3.NewSkeleton([]animation3.SkeletonBone{
				{Name: "hip"},
			})
			Expect(s.BoneRestTransform(0)).To(Equal(animation3.Transform{}))
		})
	})

	It("maps names, indices and rest transforms consistently", func() {
		restTransforms := map[string]animation3.Transform{
			"hip":   hipRest,
			"spine": spineRest,
			"head":  headRest,
		}
		for index := range skeleton.BoneCount() {
			name := skeleton.BoneName(index)
			actualIndex, ok := skeleton.BoneIndex(name)
			Expect(ok).To(BeTrue())
			Expect(actualIndex).To(Equal(index))
			Expect(skeleton.BoneRestTransform(index)).To(Equal(restTransforms[name]))
		}
	})

	When("created with duplicate names", func() {

		BeforeEach(func() {
			skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
				{Name: "hip", RestTransform: hipRest},
				{Name: "spine", RestTransform: spineRest},
				{Name: "hip", RestTransform: restTransform(4.0)},
				{Name: "head", RestTransform: headRest},
				{Name: "spine", RestTransform: restTransform(5.0)},
			})
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

		It("keeps the rest transform of the first occurrence", func() {
			Expect(skeleton.BoneRestTransform(0)).To(Equal(hipRest))
			Expect(skeleton.BoneRestTransform(1)).To(Equal(spineRest))
			Expect(skeleton.BoneRestTransform(2)).To(Equal(headRest))
		})
	})

	When("created with no bones", func() {

		DescribeTable("is empty",
			func(bones []animation3.SkeletonBone) {
				empty := animation3.NewSkeleton(bones)
				Expect(empty.BoneCount()).To(Equal(0))
				_, ok := empty.BoneIndex("hip")
				Expect(ok).To(BeFalse())
			},
			Entry("nil slice", nil),
			Entry("empty slice", []animation3.SkeletonBone{}),
		)
	})

	It("is not affected by later changes to the input slice", func() {
		bones := []animation3.SkeletonBone{
			{Name: "hip", RestTransform: hipRest},
			{Name: "spine", RestTransform: spineRest},
		}
		s := animation3.NewSkeleton(bones)
		bones[0].Name = "changed"
		bones[0].RestTransform = restTransform(4.0)

		Expect(s.BoneName(0)).To(Equal("hip"))
		Expect(s.BoneRestTransform(0)).To(Equal(hipRest))
		index, ok := s.BoneIndex("hip")
		Expect(ok).To(BeTrue())
		Expect(index).To(Equal(0))
		_, ok = s.BoneIndex("changed")
		Expect(ok).To(BeFalse())
	})

})
