package ui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/ui"
)

var _ = Describe("Position", func() {
	var position ui.Position

	BeforeEach(func() {
		position = ui.NewPosition(3, 4)
	})

	Describe("NewPosition", func() {
		It("assigns the coordinates in order", func() {
			Expect(position.X).To(Equal(3))
			Expect(position.Y).To(Equal(4))
		})
	})

	Describe("Inverse", func() {
		It("negates both coordinates", func() {
			Expect(position.Inverse()).To(Equal(ui.NewPosition(-3, -4)))
		})

		It("returns the original position when applied twice", func() {
			Expect(position.Inverse().Inverse()).To(Equal(position))
		})

		It("leaves the zero position unchanged", func() {
			Expect(ui.NewPosition(0, 0).Inverse()).To(Equal(ui.NewPosition(0, 0)))
		})

		It("does not modify the original position", func() {
			position.Inverse()
			Expect(position).To(Equal(ui.NewPosition(3, 4)))
		})
	})

	Describe("Translate", func() {
		It("offsets both coordinates", func() {
			Expect(position.Translate(ui.NewPosition(10, 20))).To(Equal(ui.NewPosition(13, 24)))
		})

		It("moves in the opposite direction for a negative delta", func() {
			Expect(position.Translate(ui.NewPosition(-5, -6))).To(Equal(ui.NewPosition(-2, -2)))
		})

		It("is undone by a translation with the inverse delta", func() {
			delta := ui.NewPosition(10, 20)
			Expect(position.Translate(delta).Translate(delta.Inverse())).To(Equal(position))
		})

		It("does not modify the original position", func() {
			position.Translate(ui.NewPosition(10, 20))
			Expect(position).To(Equal(ui.NewPosition(3, 4)))
		})
	})

	Describe("String", func() {
		It("lists the coordinates", func() {
			Expect(position.String()).To(Equal("(3, 4)"))
		})

		It("includes the sign of negative coordinates", func() {
			Expect(ui.NewPosition(-3, -4).String()).To(Equal("(-3, -4)"))
		})
	})
})

var _ = Describe("Size", func() {
	var size ui.Size

	BeforeEach(func() {
		size = ui.NewSize(30, 40)
	})

	Describe("NewSize", func() {
		It("assigns the dimensions in order", func() {
			Expect(size.Width).To(Equal(30))
			Expect(size.Height).To(Equal(40))
		})
	})

	Describe("Inverse", func() {
		It("negates both dimensions", func() {
			Expect(size.Inverse()).To(Equal(ui.NewSize(-30, -40)))
		})

		It("returns the original size when applied twice", func() {
			Expect(size.Inverse().Inverse()).To(Equal(size))
		})

		It("produces an empty size out of a non-empty one", func() {
			Expect(size.Inverse().Empty()).To(BeTrue())
		})
	})

	Describe("Grow", func() {
		It("increases both dimensions", func() {
			Expect(size.Grow(ui.NewSize(5, 6))).To(Equal(ui.NewSize(35, 46)))
		})

		It("decreases both dimensions for a negative delta", func() {
			Expect(size.Grow(ui.NewSize(-5, -6))).To(Equal(ui.NewSize(25, 34)))
		})

		It("leaves the size unchanged for a zero delta", func() {
			Expect(size.Grow(ui.NewSize(0, 0))).To(Equal(size))
		})

		It("does not modify the original size", func() {
			size.Grow(ui.NewSize(5, 6))
			Expect(size).To(Equal(ui.NewSize(30, 40)))
		})
	})

	Describe("Shrink", func() {
		It("decreases both dimensions", func() {
			Expect(size.Shrink(ui.NewSize(5, 6))).To(Equal(ui.NewSize(25, 34)))
		})

		It("increases both dimensions for a negative delta", func() {
			Expect(size.Shrink(ui.NewSize(-5, -6))).To(Equal(ui.NewSize(35, 46)))
		})

		It("is the exact opposite of Grow", func() {
			delta := ui.NewSize(5, 6)
			Expect(size.Grow(delta).Shrink(delta)).To(Equal(size))
		})

		It("is not clamped and can produce negative dimensions", func() {
			result := size.Shrink(ui.NewSize(50, 60))
			Expect(result).To(Equal(ui.NewSize(-20, -20)))
			Expect(result.Empty()).To(BeTrue())
		})
	})

	Describe("Empty", func() {
		It("returns false when both dimensions are positive", func() {
			Expect(size.Empty()).To(BeFalse())
		})

		It("returns true when the width is zero", func() {
			Expect(ui.NewSize(0, 40).Empty()).To(BeTrue())
		})

		It("returns true when the height is zero", func() {
			Expect(ui.NewSize(30, 0).Empty()).To(BeTrue())
		})

		It("returns true when the width is negative", func() {
			Expect(ui.NewSize(-30, 40).Empty()).To(BeTrue())
		})

		It("returns true when the height is negative", func() {
			Expect(ui.NewSize(30, -40).Empty()).To(BeTrue())
		})

		It("returns true for the zero size", func() {
			Expect(ui.Size{}.Empty()).To(BeTrue())
		})

		It("returns false for the smallest non-empty size", func() {
			Expect(ui.NewSize(1, 1).Empty()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("lists the dimensions", func() {
			Expect(size.String()).To(Equal("(30, 40)"))
		})
	})
})

var _ = Describe("Bounds", func() {
	var bounds ui.Bounds

	BeforeEach(func() {
		bounds = ui.NewBounds(3, 4, 30, 40)
	})

	Describe("NewBounds", func() {
		It("assigns the position and the dimensions in order", func() {
			Expect(bounds.Position).To(Equal(ui.NewPosition(3, 4)))
			Expect(bounds.Size).To(Equal(ui.NewSize(30, 40)))
		})
	})

	Describe("Empty", func() {
		It("is false when the size is not empty", func() {
			Expect(bounds.Empty()).To(BeFalse())
		})

		It("is true when the size is empty, regardless of the position", func() {
			Expect(ui.NewBounds(3, 4, 0, 40).Empty()).To(BeTrue())
			Expect(ui.NewBounds(3, 4, 30, 0).Empty()).To(BeTrue())
		})
	})

	Describe("Contains", func() {
		It("contains a position strictly inside", func() {
			Expect(bounds.Contains(ui.NewPosition(10, 20))).To(BeTrue())
		})

		It("contains the top-left corner", func() {
			Expect(bounds.Contains(ui.NewPosition(3, 4))).To(BeTrue())
		})

		It("contains the last position along each axis", func() {
			Expect(bounds.Contains(ui.NewPosition(32, 43))).To(BeTrue())
		})

		It("excludes the right edge", func() {
			Expect(bounds.Contains(ui.NewPosition(33, 20))).To(BeFalse())
		})

		It("excludes the bottom edge", func() {
			Expect(bounds.Contains(ui.NewPosition(10, 44))).To(BeFalse())
		})

		It("excludes the bottom-right corner", func() {
			Expect(bounds.Contains(ui.NewPosition(33, 44))).To(BeFalse())
		})

		It("excludes a position above the top edge", func() {
			Expect(bounds.Contains(ui.NewPosition(10, 3))).To(BeFalse())
		})

		It("excludes a position left of the left edge", func() {
			Expect(bounds.Contains(ui.NewPosition(2, 20))).To(BeFalse())
		})

		It("contains nothing when the size is empty", func() {
			empty := ui.NewBounds(3, 4, 0, 0)
			Expect(empty.Contains(ui.NewPosition(3, 4))).To(BeFalse())
			Expect(empty.Contains(ui.NewPosition(2, 3))).To(BeFalse())
		})

		It("contains nothing when the size is negative", func() {
			negative := ui.NewBounds(3, 4, -30, -40)
			Expect(negative.Contains(ui.NewPosition(3, 4))).To(BeFalse())
			Expect(negative.Contains(ui.NewPosition(-10, -20))).To(BeFalse())
		})

		It("works with negative positions", func() {
			shifted := ui.NewBounds(-10, -20, 5, 5)
			Expect(shifted.Contains(ui.NewPosition(-10, -20))).To(BeTrue())
			Expect(shifted.Contains(ui.NewPosition(-6, -16))).To(BeTrue())
			Expect(shifted.Contains(ui.NewPosition(-5, -20))).To(BeFalse())
		})
	})

	Describe("Translate", func() {
		It("offsets the position and preserves the size", func() {
			Expect(bounds.Translate(ui.NewPosition(10, 20))).To(Equal(ui.NewBounds(13, 24, 30, 40)))
		})

		It("is undone by a translation with the inverse delta", func() {
			delta := ui.NewPosition(10, 20)
			Expect(bounds.Translate(delta).Translate(delta.Inverse())).To(Equal(bounds))
		})

		It("does not modify the original bounds", func() {
			bounds.Translate(ui.NewPosition(10, 20))
			Expect(bounds).To(Equal(ui.NewBounds(3, 4, 30, 40)))
		})
	})

	Describe("Grow", func() {
		It("enlarges the size and anchors the top-left corner", func() {
			Expect(bounds.Grow(ui.NewSize(5, 6))).To(Equal(ui.NewBounds(3, 4, 35, 46)))
		})

		It("expands only to the right and downwards", func() {
			result := bounds.Grow(ui.NewSize(5, 6))
			Expect(result.Contains(ui.NewPosition(37, 49))).To(BeTrue())
			Expect(result.Contains(ui.NewPosition(2, 3))).To(BeFalse())
		})
	})

	Describe("Shrink", func() {
		It("reduces the size and anchors the top-left corner", func() {
			Expect(bounds.Shrink(ui.NewSize(5, 6))).To(Equal(ui.NewBounds(3, 4, 25, 34)))
		})

		It("contracts only from the right and the bottom", func() {
			result := bounds.Shrink(ui.NewSize(5, 6))
			Expect(result.Contains(ui.NewPosition(3, 4))).To(BeTrue())
			Expect(result.Contains(ui.NewPosition(28, 20))).To(BeFalse())
		})

		It("is not clamped and can produce empty bounds", func() {
			result := bounds.Shrink(ui.NewSize(50, 60))
			Expect(result).To(Equal(ui.NewBounds(3, 4, -20, -20)))
			Expect(result.Empty()).To(BeTrue())
		})

		It("is the exact opposite of Grow", func() {
			delta := ui.NewSize(5, 6)
			Expect(bounds.Grow(delta).Shrink(delta)).To(Equal(bounds))
		})
	})

	Describe("Resize", func() {
		It("replaces the size and preserves the position", func() {
			Expect(bounds.Resize(50, 60)).To(Equal(ui.NewBounds(3, 4, 50, 60)))
		})

		It("can produce empty bounds", func() {
			Expect(bounds.Resize(0, 0).Empty()).To(BeTrue())
		})
	})

	Describe("Intersect", func() {
		It("returns the overlapping area of two partially overlapping bounds", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(5, 5, 10, 10)
			Expect(first.Intersect(second)).To(Equal(ui.NewBounds(5, 5, 5, 5)))
		})

		It("is commutative", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(5, 5, 10, 10)
			Expect(first.Intersect(second)).To(Equal(second.Intersect(first)))
		})

		It("returns the inner bounds when one contains the other", func() {
			outer := ui.NewBounds(0, 0, 100, 100)
			inner := ui.NewBounds(10, 20, 30, 40)
			Expect(outer.Intersect(inner)).To(Equal(inner))
			Expect(inner.Intersect(outer)).To(Equal(inner))
		})

		It("returns the same bounds when intersected with itself", func() {
			Expect(bounds.Intersect(bounds)).To(Equal(bounds))
		})

		It("is idempotent", func() {
			other := ui.NewBounds(5, 5, 10, 10)
			once := bounds.Intersect(other)
			Expect(once.Intersect(other)).To(Equal(once))
		})

		It("returns empty bounds for bounds that only touch along an edge", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(10, 0, 10, 10)
			Expect(first.Intersect(second).Empty()).To(BeTrue())
		})

		It("returns empty bounds when the two do not overlap at all", func() {
			// The exact position and dimensions of the result are unspecified
			// in such a case, hence only emptiness is asserted.
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(20, 20, 5, 5)
			Expect(first.Intersect(second).Empty()).To(BeTrue())
			Expect(second.Intersect(first).Empty()).To(BeTrue())
		})

		It("returns empty bounds when the two overlap in X but not in Y", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(5, 20, 10, 10)
			Expect(first.Intersect(second).Empty()).To(BeTrue())
		})

		It("returns empty bounds when the two overlap in Y but not in X", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(20, 5, 10, 10)
			Expect(first.Intersect(second).Empty()).To(BeTrue())
		})

		It("returns empty bounds when either of the two is empty", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(5, 5, 0, 0)
			Expect(first.Intersect(second).Empty()).To(BeTrue())
			Expect(second.Intersect(first).Empty()).To(BeTrue())
		})

		It("handles negative coordinates", func() {
			first := ui.NewBounds(-10, -10, 20, 20)
			second := ui.NewBounds(-5, -5, 20, 20)
			Expect(first.Intersect(second)).To(Equal(ui.NewBounds(-5, -5, 15, 15)))
		})

		It("contains exactly the positions that both bounds contain", func() {
			first := ui.NewBounds(0, 0, 10, 10)
			second := ui.NewBounds(5, 5, 10, 10)
			result := first.Intersect(second)
			for y := -1; y <= 16; y++ {
				for x := -1; x <= 16; x++ {
					position := ui.NewPosition(x, y)
					expected := first.Contains(position) && second.Contains(position)
					Expect(result.Contains(position)).To(Equal(expected), "at %s", position)
				}
			}
		})

		It("does not modify the original bounds", func() {
			bounds.Intersect(ui.NewBounds(5, 5, 10, 10))
			Expect(bounds).To(Equal(ui.NewBounds(3, 4, 30, 40)))
		})
	})

	Describe("String", func() {
		It("lists the position and the size", func() {
			Expect(bounds.String()).To(Equal("((3, 4), (30, 40))"))
		})
	})
})
