package physics

import (
	"time"

	"github.com/mokiat/lacking/game/physics/collision"
)

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

// CreateMaterial creates a new Material that can be used to describe an
// object's behavior.
func (e *Engine) CreateMaterial(info MaterialInfo) *Material {
	return &Material{
		frictionCoefficient:    info.FrictionCoefficient,
		restitutionCoefficient: info.RestitutionCoefficient,
	}
}

// CreateBodyDefinition creates a new BodyDefinition that can be used
// to create Body instances.
func (e *Engine) CreateBodyDefinition(info BodyDefinitionInfo) *BodyDefinition {
	return &BodyDefinition{
		mass:                   info.Mass,
		momentOfInertia:        info.MomentOfInertia,
		frictionCoefficient:    info.FrictionCoefficient,
		restitutionCoefficient: info.RestitutionCoefficient,
		dragFactor:             info.DragFactor,
		angularDragFactor:      info.AngularDragFactor,
		collisionGroup:         info.CollisionGroup,
		collisionSet: collision.NewSet(
			collision.WithSpheres(info.CollisionSpheres),
			collision.WithBoxes(info.CollisionBoxes),
			collision.WithMeshes(info.CollisionMeshes),
		),
		aerodynamicShapes: info.AerodynamicShapes,
	}
}

// CreateScene creates a new Scene and configures
// the simulation for it to run at maximum stepSeconds
// intervals.
func (e *Engine) CreateScene() *Scene {
	return newScene(e, e.timestep)
}
