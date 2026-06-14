package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("VisitorBucket", func() {
	Describe("zero value", func() {
		It("starts empty", func() {
			var bucket shape2d.VisitorBucket[int]
			Expect(bucket.Items()).To(BeEmpty())
		})

		It("accepts items via Add", func() {
			var bucket shape2d.VisitorBucket[int]
			bucket.Add(1)
			Expect(bucket.Items()).To(Equal([]int{1}))
		})
	})

	Describe("NewVisitorBucket", func() {
		It("starts empty regardless of initial capacity", func() {
			bucket := shape2d.NewVisitorBucket[int](64)
			Expect(bucket.Items()).To(BeEmpty())
		})
	})

	Describe("Add", func() {
		It("appends items in order", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(10)
			bucket.Add(20)
			bucket.Add(30)
			Expect(bucket.Items()).To(Equal([]int{10, 20, 30}))
		})

		It("always returns true to continue iteration", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			Expect(bucket.Add(1)).To(BeTrue())
		})
	})

	Describe("Reset", func() {
		It("clears all stored items", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Add(2)
			bucket.Reset()
			Expect(bucket.Items()).To(BeEmpty())
		})

		It("allows the bucket to be reused after reset", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Reset()
			bucket.Add(2)
			bucket.Add(3)
			Expect(bucket.Items()).To(Equal([]int{2, 3}))
		})

		It("releases references for pointer items", func() {
			bucket := shape2d.NewVisitorBucket[*int](4)
			v := 42
			bucket.Add(&v)
			bucket.Reset()
			Expect(bucket.Items()).To(BeEmpty())
		})
	})

	Describe("Items", func() {
		It("returns all added items as a slice", func() {
			bucket := shape2d.NewVisitorBucket[string](4)
			bucket.Add("a")
			bucket.Add("b")
			Expect(bucket.Items()).To(ConsistOf("a", "b"))
		})
	})

	Describe("Each", func() {
		It("iterates all items in order", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Add(2)
			bucket.Add(3)

			var seen []int
			bucket.Each(func(item int) bool {
				seen = append(seen, item)
				return true
			})
			Expect(seen).To(Equal([]int{1, 2, 3}))
		})

		It("stops early when yield returns false", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Add(2)
			bucket.Add(3)

			var seen []int
			bucket.Each(func(item int) bool {
				seen = append(seen, item)
				return item < 2
			})
			Expect(seen).To(Equal([]int{1, 2}))
		})

		It("does nothing on an empty bucket", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			called := false
			bucket.Each(func(int) bool {
				called = true
				return true
			})
			Expect(called).To(BeFalse())
		})
	})

	Describe("Iter", func() {
		It("iterates all items in order", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Add(2)
			bucket.Add(3)

			var seen []int
			for item := range bucket.Iter() {
				seen = append(seen, item)
			}
			Expect(seen).To(Equal([]int{1, 2, 3}))
		})

		It("supports early termination via break", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			bucket.Add(1)
			bucket.Add(2)
			bucket.Add(3)

			var seen []int
			for item := range bucket.Iter() {
				seen = append(seen, item)
				if item == 2 {
					break
				}
			}
			Expect(seen).To(Equal([]int{1, 2}))
		})

		It("does nothing on an empty bucket", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			called := false
			for range bucket.Iter() {
				called = true
			}
			Expect(called).To(BeFalse())
		})
	})

	Describe("VisitorFunc", func() {
		It("returns a function that adds items to the bucket", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			fn := bucket.VisitorFunc()
			fn(10)
			fn(20)
			Expect(bucket.Items()).To(Equal([]int{10, 20}))
		})

		It("returned function always returns true", func() {
			bucket := shape2d.NewVisitorBucket[int](4)
			fn := bucket.VisitorFunc()
			Expect(fn(1)).To(BeTrue())
		})
	})
})
