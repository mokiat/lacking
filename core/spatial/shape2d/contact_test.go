package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Contact", func() {
	var contact shape2d.Contact

	BeforeEach(func() {
		contact = shape2d.Contact{
			TargetPoint:  dprec.NewVec2(10.0, 0.0),
			TargetNormal: dprec.NewVec2(1.0, 0.0),
			Depth:        2.0,
		}
	})

	Describe("EvalSourcePoint", func() {
		It("lies Depth away from TargetPoint along the inverse normal", func() {
			Expect(contact.EvalSourcePoint()).To(dprectest.HaveVec2Coords(8.0, 0.0))
		})
	})

	Describe("EvalSourceNormal", func() {
		It("is the inverse of TargetNormal", func() {
			Expect(contact.EvalSourceNormal()).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})
	})

	Describe("Flipped", func() {
		It("promotes the source point and normal to the target", func() {
			flipped := contact.Flipped()
			Expect(flipped.TargetPoint).To(dprectest.HaveVec2Coords(8.0, 0.0))
			Expect(flipped.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(flipped.Depth).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("round-trips back to the original when applied twice", func() {
			result := contact.Flipped().Flipped()
			Expect(result.TargetPoint).To(dprectest.HaveVec2Coords(contact.TargetPoint.X, contact.TargetPoint.Y))
			Expect(result.TargetNormal).To(dprectest.HaveVec2Coords(contact.TargetNormal.X, contact.TargetNormal.Y))
			Expect(result.Depth).To(BeNumerically("~", contact.Depth, 1e-6))
		})

		It("swaps the roles of the source and target points", func() {
			flipped := contact.Flipped()
			Expect(flipped.EvalSourcePoint()).To(dprectest.HaveVec2Coords(contact.TargetPoint.X, contact.TargetPoint.Y))
		})
	})
})

func newContact(depth float64) shape2d.Contact {
	return shape2d.Contact{
		TargetPoint:  dprec.NewVec2(depth, 0.0),
		TargetNormal: dprec.NewVec2(1.0, 0.0),
		Depth:        depth,
	}
}

var _ = Describe("LastContact", func() {
	var sink shape2d.LastContact

	BeforeEach(func() {
		sink = shape2d.LastContact{}
	})

	It("reports no contact when none were added", func() {
		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})

	It("retains the most recently added contact", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(3.0))
		sink.AddContact(newContact(2.0))

		contact, ok := sink.Contact()
		Expect(ok).To(BeTrue())
		Expect(contact.Depth).To(BeNumerically("~", 2.0, 1e-6))
	})

	It("forgets contacts after Reset", func() {
		sink.AddContact(newContact(1.0))
		sink.Reset()

		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("ContactList", func() {
	var sink shape2d.ContactList

	BeforeEach(func() {
		sink = shape2d.ContactList{}
	})

	It("is empty when no contacts were added", func() {
		Expect(sink.Contacts()).To(BeEmpty())
	})

	It("retains every contact in the order it was added", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(3.0))
		sink.AddContact(newContact(2.0))

		contacts := sink.Contacts()
		Expect(contacts).To(HaveLen(3))
		Expect(contacts[0].Depth).To(BeNumerically("~", 1.0, 1e-6))
		Expect(contacts[1].Depth).To(BeNumerically("~", 3.0, 1e-6))
		Expect(contacts[2].Depth).To(BeNumerically("~", 2.0, 1e-6))
	})

	It("preserves duplicate contacts", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(1.0))

		Expect(sink.Contacts()).To(HaveLen(2))
	})

	It("clears the contacts after Reset", func() {
		sink.AddContact(newContact(1.0))
		sink.Reset()

		Expect(sink.Contacts()).To(BeEmpty())
	})

	It("reuses its capacity across Reset", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(2.0))
		capBefore := cap(sink.Contacts())

		sink.Reset()
		sink.AddContact(newContact(3.0))

		Expect(cap(sink.Contacts())).To(Equal(capBefore))
	})

	It("can be ranged over directly as a slice", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(2.0))

		var total float64
		for _, contact := range sink {
			total += contact.Depth
		}
		Expect(total).To(BeNumerically("~", 3.0, 1e-6))
	})

	It("does not grow its backing array when pre-sized with make", func() {
		list := make(shape2d.ContactList, 0, 4)
		capBefore := cap(list)
		for i := range 4 {
			list.AddContact(newContact(float64(i)))
		}
		Expect(cap(list)).To(Equal(capBefore))
		Expect(list).To(HaveLen(4))
	})
})

var _ = Describe("DeepestContact", func() {
	var sink shape2d.DeepestContact

	BeforeEach(func() {
		sink = shape2d.DeepestContact{}
	})

	It("reports no contact when none were added", func() {
		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})

	It("retains the contact with the greatest depth", func() {
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(3.0))
		sink.AddContact(newContact(2.0))

		contact, ok := sink.Contact()
		Expect(ok).To(BeTrue())
		Expect(contact.Depth).To(BeNumerically("~", 3.0, 1e-6))
	})

	It("forgets contacts after Reset", func() {
		sink.AddContact(newContact(3.0))
		sink.Reset()

		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("ShallowestContact", func() {
	var sink shape2d.ShallowestContact

	BeforeEach(func() {
		sink = shape2d.ShallowestContact{}
	})

	It("reports no contact when none were added", func() {
		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})

	It("retains the contact with the smallest depth", func() {
		sink.AddContact(newContact(3.0))
		sink.AddContact(newContact(1.0))
		sink.AddContact(newContact(2.0))

		contact, ok := sink.Contact()
		Expect(ok).To(BeTrue())
		Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
	})

	It("forgets contacts after Reset", func() {
		sink.AddContact(newContact(1.0))
		sink.Reset()

		_, ok := sink.Contact()
		Expect(ok).To(BeFalse())
	})
})
