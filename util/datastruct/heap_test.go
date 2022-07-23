package datastruct_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/util/datastruct"
)

var _ = Describe("Heap", func() {
	var heap *datastruct.Heap[int]

	lessInt := func(a, b int) bool {
		return a < b
	}

	BeforeEach(func() {
		heap = datastruct.NewHeap(lessInt)
	})

	It("is empty by default", func() {
		Expect(heap.IsEmpty()).To(BeTrue())
	})

	When("items are added", func() {
		BeforeEach(func() {
			heap.Push(21)
			heap.Push(10)
			heap.Push(2)
			heap.Push(15)
			heap.Push(6)
			heap.Push(15)
		})

		It("is not empty", func() {
			Expect(heap.IsEmpty()).To(BeFalse())
		})

		It("is possible to peek smallest item", func() {
			Expect(heap.Peek()).To(Equal(2))
		})

		It("is possible to fetch all items in order", func() {
			Expect(heap.Pop()).To(Equal(2))
			Expect(heap.Pop()).To(Equal(6))
			Expect(heap.Pop()).To(Equal(10))
			Expect(heap.Pop()).To(Equal(15))
			Expect(heap.Pop()).To(Equal(15))
			Expect(heap.Pop()).To(Equal(21))
			Expect(heap.IsEmpty()).To(BeTrue())
		})

		When("some items are removed and more are added", func() {
			BeforeEach(func() {
				heap.Pop() // 2
				heap.Pop() // 6
				heap.Pop() // 10
				heap.Pop() // 15
				Expect(heap.IsEmpty()).To(BeFalse())

				heap.Push(1)
				heap.Push(31)
				heap.Push(1)
				heap.Push(12)
			})

			It("is possible to fetch new and remaining items in order", func() {
				Expect(heap.Pop()).To(Equal(1))
				Expect(heap.Pop()).To(Equal(1))
				Expect(heap.Pop()).To(Equal(12))
				Expect(heap.Pop()).To(Equal(15))
				Expect(heap.Pop()).To(Equal(21))
				Expect(heap.Pop()).To(Equal(31))
				Expect(heap.IsEmpty()).To(BeTrue())
			})
		})
	})
})
