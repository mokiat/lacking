package physics_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	"github.com/mokiat/lacking/game/physics"
)

var _ = Describe("Scene", func() {
	const step = 16 * time.Millisecond

	var scene *physics.Scene

	BeforeEach(func() {
		scene = physics.NewScene()
	})

	Describe("collision notifications while paused", func() {
		type pairEvent struct {
			first, second physics.BodyID
			active        bool
		}
		type soloEvent struct {
			body    physics.BodyID
			terrain physics.TerrainID
			active  bool
		}

		var (
			pairEvents []pairEvent
			soloEvents []soloEvent
		)

		BeforeEach(func() {
			pairEvents = nil
			soloEvents = nil
			scene.SubscribePairCollision(func(first, second physics.BodyID, active bool) {
				pairEvents = append(pairEvents, pairEvent{first, second, active})
			})
			scene.SubscribeSoloCollision(func(body physics.BodyID, terrain physics.TerrainID, active bool) {
				soloEvents = append(soloEvents, soloEvent{body, terrain, active})
			})
		})

		// attachSphere makes the body immovable, so that the collision
		// persists across steps instead of being resolved, and gives it a
		// unit sphere collision shape.
		attachSphere := func(body physics.BodyHandle) {
			body.SetMass(math.Inf(1))
			body.AttachCollisionSphere(physics.CollisionSphere{
				Shape: shape3d.NewSphere(dprec.ZeroVec3(), 1.0),
			})
		}

		When("two bodies are colliding", func() {
			BeforeEach(func() {
				bodies := scene.Bodies()
				attachSphere(bodies.CreateHandle(dprec.NewVec3(0.0, 0.0, 0.0), dprec.IdentityQuat()))
				attachSphere(bodies.CreateHandle(dprec.NewVec3(0.5, 0.0, 0.0), dprec.IdentityQuat()))

				scene.Update(step)
				Expect(pairEvents).To(HaveLen(1))
				Expect(pairEvents[0].active).To(BeTrue())
			})

			It("does not report the collision as ended when paused", func() {
				scene.SetTimeScale(0.0)
				scene.Update(step)
				scene.Update(step)
				Expect(pairEvents).To(HaveLen(1))
			})

			It("does not report the collision again when resumed", func() {
				scene.SetTimeScale(0.0)
				scene.Update(step)
				scene.SetTimeScale(1.0)
				scene.Update(step)
				Expect(pairEvents).To(HaveLen(1))
			})
		})

		When("a body is colliding with a terrain", func() {
			BeforeEach(func() {
				terrain := scene.Terrains().CreateHandle()
				terrain.AttachCollisionMesh(physics.CollisionMesh{
					Shape: shape3d.NewMesh([]shape3d.Triangle{
						shape3d.NewTriangle(
							dprec.NewVec3(-10.0, 0.0, 10.0),
							dprec.NewVec3(10.0, 0.0, 10.0),
							dprec.NewVec3(0.0, 0.0, -10.0),
						),
					}),
				})
				attachSphere(scene.Bodies().CreateHandle(dprec.NewVec3(0.0, 0.5, 0.0), dprec.IdentityQuat()))

				scene.Update(step)
				Expect(soloEvents).To(HaveLen(1))
				Expect(soloEvents[0].active).To(BeTrue())
			})

			It("does not report the collision as ended when paused", func() {
				scene.SetTimeScale(0.0)
				scene.Update(step)
				scene.Update(step)
				Expect(soloEvents).To(HaveLen(1))
			})

			It("does not report the collision again when resumed", func() {
				scene.SetTimeScale(0.0)
				scene.Update(step)
				scene.SetTimeScale(1.0)
				scene.Update(step)
				Expect(soloEvents).To(HaveLen(1))
			})
		})
	})
})
