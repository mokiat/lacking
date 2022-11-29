package physics

import "time"

// NewEngine creates a new physics engine.
func NewEngine(timestep time.Duration) *Engine {
	return &Engine{
		timestep: timestep,
	}
}

// Engine is the entrypoint to working with a
// physics simulation.
type Engine struct {
	timestep time.Duration
}

// CreateBodyDefinition creates a new BodyDefinition that can be used
// to create Body instances.
func (e *Engine) CreateBodyDefinition(info BodyDefinitionInfo) *BodyDefinition {
	return &BodyDefinition{
		mass:                   info.Mass,
		momentOfInertia:        info.MomentOfInertia,
		restitutionCoefficient: info.RestitutionCoefficient,
		dragFactor:             info.DragFactor,
		angularDragFactor:      info.AngularDragFactor,
		collisionGroup:         info.CollisionGroup,
		collisionShapes:        info.CollisionShapes,
		aerodynamicShapes:      info.AerodynamicShapes,
	}
}

// CreateScene creates a new Scene and configures
// the simulation for it to run at maximum stepSeconds
// intervals.
func (e *Engine) CreateScene() *Scene {
	return newScene(e, e.timestep)
}
