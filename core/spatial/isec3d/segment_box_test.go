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
		It("returns true for a segment passing through the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(-3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
		})

		It("returns true for a segment lying entirely inside the box", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentBox(seg, box)).To(BeTrue())
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
			// The line x+y=2.1 passes just outside the (1,1) corner; only the
			// edge cross-product axes can separate this case.
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
		It("yields a minimum-translation contact for a shallow penetration", func() {
			// Dips into the top (+Y) face: B is 0.5 below the surface.
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
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("produces the same contact regardless of endpoint order", func() {
			forward := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, 0.5, 0.0),
			}
			reversed := forward.Flipped()

			var sinkForward, sinkReversed shape3d.LastContact
			isec3d.ResolveSegmentBox(forward, box, sinkForward.AddContact)
			isec3d.ResolveSegmentBox(reversed, box, sinkReversed.AddContact)

			cf, okF := sinkForward.Contact()
			cr, okR := sinkReversed.Contact()
			Expect(okF).To(BeTrue())
			Expect(okR).To(BeTrue())
			Expect(cr.TargetNormal).To(dprectest.HaveVec3Coords(cf.TargetNormal.X, cf.TargetNormal.Y, cf.TargetNormal.Z))
			Expect(cr.Depth).To(BeNumerically("~", cf.Depth, 1e-6))
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

		It("reports the contact point on the box surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.3, 2.0, 0.2),
				B: dprec.NewVec3(0.3, 0.5, 0.2),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)
			contact, ok := sink.Contact()

			Expect(ok).To(BeTrue())
			// On the top face: y is at the half-height and the normal is unit.
			Expect(contact.TargetPoint.Y).To(BeNumerically("~", box.HalfHeight, 1e-6))
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("no longer penetrates after the segment is moved by the contact", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, 0.5, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentBox(seg, box, sink.AddContact)
			contact, _ := sink.Contact()

			lift := dprec.Vec3Prod(contact.TargetNormal, contact.Depth)
			moved := shape3d.Segment{
				A: dprec.Vec3Sum(seg.A, lift),
				B: dprec.Vec3Sum(seg.B, lift),
			}
			// Re-resolving yields at most a touching (zero-depth) contact.
			var resink shape3d.LastContact
			isec3d.ResolveSegmentBox(moved, box, resink.AddContact)
			if reContact, ok := resink.Contact(); ok {
				Expect(reContact.Depth).To(BeNumerically("~", 0.0, 1e-6))
			}
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
			// Dips into the world +X face: B is 0.5 inside (world x extent is 1).
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
