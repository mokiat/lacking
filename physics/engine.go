package physics

import (
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/shape"
)

const (
	gravity     = 9.8
	windDensity = 1.2

	impulseIterations = 100
	nudgeIterations   = 100
)

func NewEngine(step time.Duration) *Engine {
	return &Engine{
		step:            step,
		accumulatedTime: 0,

		gravity:      sprec.NewVec3(0.0, -gravity, 0.0),
		windVelocity: sprec.NewVec3(0.0, 0.0, 0.0),
		windDensity:  windDensity,

		intersectionSet: shape.NewIntersectionResultSet(128),
	}
}

type Engine struct {
	step            time.Duration
	accumulatedTime time.Duration

	gravity      sprec.Vec3
	windVelocity sprec.Vec3
	windDensity  float32

	bodies               []*Body
	constraints          []Constraint
	collisionConstraints []Constraint

	intersectionSet *shape.IntersectionResultSet
}

func (e *Engine) Bodies() []*Body {
	return e.bodies
}

func (e *Engine) Update(elapsedTime time.Duration) {
	e.accumulatedTime += elapsedTime
	for e.accumulatedTime > e.step {
		e.accumulatedTime -= e.step
		e.runSimulation(Context{
			ElapsedSeconds:    float32(e.step.Seconds()),
			ImpulseIterations: impulseIterations,
			NudgeIterations:   nudgeIterations,
		})
	}
}

func (e *Engine) Add(aspect interface{}) {
	if body, ok := aspect.(*Body); ok {
		e.AddBody(body)
	}
	if constraint, ok := aspect.(Constraint); ok {
		e.AddConstraint(constraint)
	}
}

func (e *Engine) AddBody(body *Body) {
	e.bodies = append(e.bodies, body)
}

func (e *Engine) AddConstraint(constraint Constraint) {
	e.constraints = append(e.constraints, constraint)
}

func (e *Engine) runSimulation(ctx Context) {
	e.detectCollisions()
	e.resetConstraints()
	e.applyForces()
	e.integrate(ctx)
	e.applyImpulses(ctx)
	e.applyMotion(ctx)
	e.applyNudges(ctx)
}

func (e *Engine) resetConstraints() {
	for _, constraint := range e.constraints {
		constraint.Reset()
	}
	for _, constraint := range e.collisionConstraints {
		constraint.Reset()
	}
}

func (e *Engine) applyForces() {
	for _, body := range e.bodies {
		if body.IsStatic {
			continue
		}
		body.ResetAcceleration()
		body.ResetAngularAcceleration()

		body.AddAcceleration(e.gravity)
		deltaWindVelocity := sprec.Vec3Diff(e.windVelocity, body.Velocity)
		body.ApplyForce(sprec.Vec3Prod(deltaWindVelocity, e.windDensity*body.DragFactor*deltaWindVelocity.Length()))
		body.ApplyTorque(sprec.Vec3Prod(body.AngularVelocity, -e.windDensity*body.AngularDragFactor*body.AngularVelocity.Length()))
	}

	// TODO: Restrict max linear + angular accelerations
}

func (e *Engine) integrate(ctx Context) {
	for _, body := range e.bodies {
		if body.IsStatic {
			continue
		}
		deltaVelocity := sprec.Vec3Prod(body.Acceleration, ctx.ElapsedSeconds)
		body.AddVelocity(deltaVelocity)
		deltaAngularVelocity := sprec.Vec3Prod(body.AngularAcceleration, ctx.ElapsedSeconds)
		body.AddAngularVelocity(deltaAngularVelocity)

		// TODO: Restrict max linear + angular velocities
	}
}

func (e *Engine) applyImpulses(ctx Context) {
	for i := 0; i < ctx.ImpulseIterations; i++ {
		for _, constraint := range e.constraints {
			constraint.ApplyImpulse(ctx)
		}
		for _, constraint := range e.collisionConstraints {
			constraint.ApplyImpulse(ctx)
		}
	}
}

func (e *Engine) applyMotion(ctx Context) {
	for _, body := range e.bodies {
		deltaPosition := sprec.Vec3Prod(body.Velocity, ctx.ElapsedSeconds)
		body.Translate(deltaPosition)
		deltaRotation := sprec.Vec3Prod(body.AngularVelocity, ctx.ElapsedSeconds)
		body.Rotate(deltaRotation)
	}
}

func (e *Engine) applyNudges(ctx Context) {
	for i := 0; i < ctx.NudgeIterations; i++ {
		for _, constraint := range e.constraints {
			constraint.ApplyNudge(ctx)
		}
		for _, constraint := range e.collisionConstraints {
			constraint.ApplyNudge(ctx)
		}
	}
}

func (e *Engine) detectCollisions() {
	for _, body := range e.bodies {
		body.InCollision = false
	}

	e.collisionConstraints = e.collisionConstraints[:0]
	for i := 0; i < len(e.bodies); i++ {
		for j := i + 1; j < len(e.bodies); j++ {
			first := e.bodies[i]
			second := e.bodies[j]
			e.checkCollisionTwoBodies(first, second)
		}
	}
}

func (e *Engine) checkCollisionTwoBodies(first, second *Body) {
	if first.IsStatic && second.IsStatic {
		return
	}

	// FIXME: Temporary, to prevent non-static entities from colliding for now
	// Currently, only static to non-static is supported
	if !first.IsStatic && !second.IsStatic {
		return
	}

	for _, firstPlacement := range first.CollisionShapes {
		firstPlacementWS := firstPlacement.Transformed(first.Position, first.Orientation)

		for _, secondPlacement := range second.CollisionShapes {
			secondPlacementWS := secondPlacement.Transformed(second.Position, second.Orientation)

			e.intersectionSet.Reset()
			shape.CheckIntersection(firstPlacementWS, secondPlacementWS, e.intersectionSet)

			if e.intersectionSet.Found() {
				first.InCollision = true
				second.InCollision = true
			}

			for _, intersection := range e.intersectionSet.Intersections() {
				// TODO: Once both non-static are supported, a dual-body collision constraint
				// should be used instead of individual uni-body constraints

				if !first.IsStatic {
					e.collisionConstraints = append(e.collisionConstraints, GroundCollisionConstraint{
						Body:         first,
						Normal:       intersection.FirstDisplaceNormal,
						ContactPoint: intersection.FirstContact,
						Depth:        intersection.Depth,
					})
				}

				if !second.IsStatic {
					e.collisionConstraints = append(e.collisionConstraints, GroundCollisionConstraint{
						Body:         second,
						Normal:       intersection.SecondDisplaceNormal,
						ContactPoint: intersection.SecondContact,
						Depth:        intersection.Depth,
					})
				}
			}
		}
	}
}
