package physics

// NewEngine creates a new physics engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Engine is the entrypoint to working with a
// physics simulation.
type Engine struct{}

// CreateBodyDefinition creates a new BodyDefinition that can be used
// to create Body instances.
func (e *Engine) CreateBodyDefinition(info BodyDefinitionInfo) *BodyDefinition {
	return &BodyDefinition{
		mass:                   info.Mass,
		momentOfInertia:        info.MomentOfInertia,
		restitutionCoefficient: info.RestitutionCoefficient,
		dragFactor:             info.DragFactor,
		angularDragFactor:      info.AngularDragFactor,
		collisionShapes:        info.CollisionShapes,
		aerodynamicShapes:      info.AerodynamicShapes,
	}
}

// CreateScene creates a new Scene and configures
// the simulation for it to run at maximum stepSeconds
// intervals.
func (e *Engine) CreateScene(stepSeconds float64) *Scene {
	return newScene(e, stepSeconds)
}
