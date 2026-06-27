package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SegmentBox", func() {
	// A cube of half-extent 1 centered at the origin, axis aligned.
	var box shape3d.Box

	BeforeEach(func() {
		box = shape3d.Box{
			Center:     dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:   shape3d.IdentityRotation(),
			HalfWidth:  1.0,
			HalfHeight: 1.0,
			HalfLength: 1.0,
		}
	})

	Describe("CheckSegmentBox", func() {
		It("returns true for a segment entering the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
		})

		It("returns true entering from the opposite direction", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(-3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
		})

		It("returns false for a segment lying entirely inside the box", func() {
			// Face-culled: there is no crossing into a front face from outside.
			seg := shape3d.Segment{
				A: dprec.NewVec3(-0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeFalse())
		})

		It("returns false for a segment that starts inside the box", func() {
			// The start is inside, so the only crossing is an exit through a
			// back-facing face, which is culled.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeFalse())
		})

		It("returns false for a segment that misses the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 2.0, 0.0),
				B: dprec.NewVec3(3.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeFalse())
		})

		It("returns false when the box lies beyond the segment's extent", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeFalse())
		})

		It("returns false for a diagonal segment that slips past a corner", func() {
			// The line x+y=2.1 passes just outside the (1,1) corner.
			seg := shape3d.Segment{
				A: dprec.NewVec3(2.1, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 2.1, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeFalse())
		})

		It("returns true for a diagonal segment that clips a corner", func() {
			// The line x+y=1.9 passes just inside the (1,1) corner.
			seg := shape3d.Segment{
				A: dprec.NewVec3(1.9, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 1.9, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
		})

		Context("with a rotated, non-cube box", func() {
			var rotated shape3d.Box

			BeforeEach(func() {
				// A box long along its local X, rotated 90 degrees about Z, so in
				// world space it extends +/-2 along Y and +/-0.5 along X and Z.
				rotated = shape3d.Box{
					Center:     dprec.NewVec3(0.0, 0.0, 0.0),
					Rotation:   shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
					HalfWidth:  2.0,
					HalfHeight: 0.5,
					HalfLength: 0.5,
				}
			})

			It("returns true for a segment along the rotated long axis", func() {
				seg := shape3d.Segment{
					A: dprec.NewVec3(0.0, -3.0, 0.0),
					B: dprec.NewVec3(0.0, 3.0, 0.0),
				}
				Expect(isec3d.CheckSegmentBox(seg, rotated)).To(BeTrue())
			})

			It("returns false for a segment beyond the rotated short axis", func() {
				seg := shape3d.Segment{
					A: dprec.NewVec3(1.0, -3.0, 0.0),
					B: dprec.NewVec3(1.0, 3.0, 0.0),
				}
				Expect(isec3d.CheckSegmentBox(seg, rotated)).To(BeFalse())
			})
		})
	})

	Describe("ResolveSegmentBox", func() {
		It("yields a contact at the entry face", func() {
			// Enters the top (+Y) face at (0, 1, 0); B sits 0.5 below the surface.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, 0.5, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			// B at y=0.5 lies 0.5 below the entry plane y=1.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("places the contact point where the segment crosses the surface", func() {
			// A diagonal segment that enters the top (+Y) face at (0, 1, 0).
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 2.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
		})

		It("reports the entry face for the side the segment comes from", func() {
			// Reversing the direction makes the segment enter the +X face instead.
			seg := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(-3.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
		})

		It("does not yield a contact when the segment misses the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 2.0, 0.0),
				B: dprec.NewVec3(3.0, 2.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the segment starts inside the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 3.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("reports a contact point on the box surface with a unit normal", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.3, 2.0, 0.2),
				B: dprec.NewVec3(0.3, 0.5, 0.2),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)
			contact, ok := sink.Contact()

			Expect(ok).To(BeTrue())
			// Enters the top face: the entry keeps the segment's x and z.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.3, 1.0, 0.2))
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("brings the far endpoint onto the entry face when moved by Depth", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, 0.5, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)
			contact, _ := sink.Contact()

			// Moving B along the normal by Depth lands it on the entry plane.
			movedB := dprec.Vec3Sum(seg.B, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			Expect(movedB.Y).To(BeNumerically("~", box.HalfHeight, 1e-6))
		})

		It("resolves against a rotated box in world space", func() {
			// A box long along its local X, rotated 90 degrees about Z, so in
			// world space it extends +/-1 along X and +/-2 along Y.
			rotated := shape3d.Box{
				Center:     dprec.NewVec3(0.0, 0.0, 0.0),
				Rotation:   shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
				HalfWidth:  2.0,
				HalfHeight: 1.0,
				HalfLength: 1.0,
			}
			// Enters the world +X face at (1, 0, 0); B sits 0.5 inside.
			seg := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, rotated, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})
	})
})
