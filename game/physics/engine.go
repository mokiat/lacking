package physics

import (
	"time"

	"github.com/mokiat/lacking/game/physics/collision"
)

// NewEngine creates a new physics engine.
func NewEngine(opts ...Option) *Engine {
	cfg := config{
		Timestep: 16 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Engine{
		timestep: cfg.Timestep,
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
