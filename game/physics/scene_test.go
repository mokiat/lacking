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
			var secondBody physics.BodyHandle

			BeforeEach(func() {
				bodies := scene.Bodies()
				attachSphere(bodies.CreateHandle(dprec.NewVec3(0.0, 0.0, 0.0), dprec.IdentityQuat()))
				secondBody = bodies.CreateHandle(dprec.NewVec3(0.5, 0.0, 0.0), dprec.IdentityQuat())
				attachSphere(secondBody)

				scene.Update(step)
				Expect(pairEvents).To(HaveLen(1))
				Expect(pairEvents[0].active).To(BeTrue())
			})

			It("can continue simulating after one of the bodies is deleted", func() {
				scene.Bodies().Delete(secondBody.ID())
				Expect(func() { scene.Update(step) }).ToNot(Panic())
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
			var body physics.BodyHandle

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
				body = scene.Bodies().CreateHandle(dprec.NewVec3(0.0, 0.5, 0.0), dprec.IdentityQuat())
				attachSphere(body)

				scene.Update(step)
				Expect(soloEvents).To(HaveLen(1))
				Expect(soloEvents[0].active).To(BeTrue())
			})

			It("can continue simulating after the body is deleted", func() {
				scene.Bodies().Delete(body.ID())
				Expect(func() { scene.Update(step) }).ToNot(Panic())
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

	Describe("contact resolution", func() {
		const (
			// groundLevel is the height of the surface that bodies land on.
			groundLevel = 0.0

			settleSteps  = 300
			observeSteps = 120
		)

		// boxSize describes a thin slab. Collision detection reports a
		// single contact point per step for such a shape lying on its face,
		// which on its own makes the slab endlessly pivot around that point.
		boxSize := dprec.NewVec3(0.1, 0.04, 0.25)

		createGroundTerrain := func(restitution float64) {
			terrain := scene.Terrains().CreateHandle()
			terrain.AttachCollisionMesh(physics.CollisionMesh{
				Shape: shape3d.NewMesh([]shape3d.Triangle{
					shape3d.NewTriangle(
						dprec.NewVec3(-10.0, groundLevel, 10.0),
						dprec.NewVec3(10.0, groundLevel, 10.0),
						dprec.NewVec3(0.0, groundLevel, -10.0),
					),
				}),
				FrictionCoefficient:    1.0,
				RestitutionCoefficient: restitution,
			})
		}

		// createGroundBody creates a body that is held in place and has a
		// wide and long box collision shape, the top of which is at ground
		// level.
		createGroundBody := func() physics.BodyHandle {
			const mass = 10.0
			size := dprec.NewVec3(1.2, 0.2, 5.0)
			position := dprec.NewVec3(0.0, groundLevel-size.Y/2.0, 0.0)

			body := scene.Bodies().CreateHandle(position, dprec.IdentityQuat())
			body.SetMass(mass)
			body.SetMomentOfInertia(physics.SolidBoxMomentOfInertia(mass, size.X, size.Y, size.Z))
			body.AttachCollisionBox(physics.CollisionBox{
				Shape:                  shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.Vec3Quot(size, 2.0)),
				FrictionCoefficient:    1.0,
				RestitutionCoefficient: 0.1,
			})
			scene.SoloConstraints().Create(body.ID(), physics.NewFixedPositionSolver(physics.FixedPositionSolverConfig{
				Position: position,
			}))
			scene.SoloConstraints().Create(body.ID(), physics.NewFixedRotationSolver(physics.FixedRotationSolverConfig{
				Rotation: dprec.IdentityQuat(),
			}))
			return body
		}

		// createBox creates a slab-shaped body that lies flat, its center
		// being at the specified position.
		createBox := func(position dprec.Vec3) physics.BodyHandle {
			const mass = 1.0
			body := scene.Bodies().CreateHandle(position, dprec.IdentityQuat())
			body.SetMass(mass)
			body.SetMomentOfInertia(physics.SolidBoxMomentOfInertia(mass, boxSize.X, boxSize.Y, boxSize.Z))
			body.AttachCollisionBox(physics.CollisionBox{
				Shape:                  shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.Vec3Quot(boxSize, 2.0)),
				FrictionCoefficient:    0.8,
				RestitutionCoefficient: 0.4,
			})
			return body
		}

		// tilt returns the angle, in degrees, between the up direction of
		// the body and the up direction of the world.
		tilt := func(body physics.BodyHandle) float64 {
			up := dprec.QuatVec3Rotation(body.Rotation(), dprec.BasisYVec3())
			return dprec.Vec3Angle(up, dprec.BasisYVec3()).Degrees()
		}

		// expectAtRest lets the body settle and then asserts that it lies
		// still and flat, its center being at the specified height.
		expectAtRest := func(body physics.BodyHandle, height float64) {
			for range settleSteps {
				scene.Update(step)
			}
			for range observeSteps {
				scene.Update(step)
				Expect(body.AngularVelocity().Length()).To(BeNumerically("<", 0.5))
				Expect(body.Velocity().Length()).To(BeNumerically("<", 0.05))
				Expect(tilt(body)).To(BeNumerically("<", 2.0))
				Expect(body.Position().Y).To(BeNumerically("~", height, 0.005))
			}
		}

		BeforeEach(func() {
			scene.GlobalAccelerators().Create(physics.NewGravitySolver())
		})

		It("brings a box that is dropped flat onto a terrain to rest", func() {
			createGroundTerrain(0.1)
			box := createBox(dprec.NewVec3(0.1, groundLevel+boxSize.Y/2.0+0.05, 0.2))

			expectAtRest(box, groundLevel+boxSize.Y/2.0)
		})

		It("brings a box that is dropped flat onto another body to rest", func() {
			createGroundBody()
			box := createBox(dprec.NewVec3(0.1, groundLevel+boxSize.Y/2.0+0.05, 0.2))

			expectAtRest(box, groundLevel+boxSize.Y/2.0)
		})

		It("keeps a box that is dropped flat onto another box on top of it", func() {
			createGroundBody()
			lower := createBox(dprec.NewVec3(0.0, groundLevel+boxSize.Y/2.0+0.05, 0.0))
			upper := createBox(dprec.NewVec3(0.01, groundLevel+boxSize.Y*1.5+0.06, 0.02))

			for range settleSteps {
				scene.Update(step)
			}
			for range observeSteps {
				scene.Update(step)
				Expect(tilt(lower)).To(BeNumerically("<", 5.0))
				Expect(tilt(upper)).To(BeNumerically("<", 5.0))
				Expect(lower.Position().Y).To(BeNumerically("~", groundLevel+boxSize.Y/2.0, 0.005))
				Expect(upper.Position().Y).To(BeNumerically("~", groundLevel+boxSize.Y*1.5, 0.005))
			}
		})

		It("bounces a ball off of a terrain according to restitution", func() {
			const (
				radius      = 0.25
				dropHeight  = 2.0
				restitution = 0.5
			)
			createGroundTerrain(restitution)
			ball := scene.Bodies().CreateHandle(dprec.NewVec3(0.0, groundLevel+radius+dropHeight, 0.0), dprec.IdentityQuat())
			ball.SetMass(1.0)
			ball.SetMomentOfInertia(physics.SolidSphereMomentOfInertia(1.0, radius))
			ball.AttachCollisionSphere(physics.CollisionSphere{
				Shape:                  shape3d.NewSphere(dprec.ZeroVec3(), radius),
				FrictionCoefficient:    0.5,
				RestitutionCoefficient: restitution,
			})

			// Simulate past the first bounce and up to the peak that follows.
			hasBounced := false
			reboundHeight := 0.0
			for range 250 {
				scene.Update(step)
				if ball.Velocity().Y > 0.0 {
					hasBounced = true
				}
				if hasBounced {
					reboundHeight = max(reboundHeight, ball.Position().Y-radius-groundLevel)
				}
			}

			// A ball that leaves the surface at a fraction of the speed it
			// arrived with reaches that fraction squared of its drop height.
			Expect(hasBounced).To(BeTrue())
			Expect(reboundHeight).To(BeNumerically("~", restitution*restitution*dropHeight, 0.1))
		})
	})
})
