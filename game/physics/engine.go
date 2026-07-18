package physics

// NewEngine creates a new physics engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Engine is the entrypoint to working with a
// physics simulation.
type Engine struct{}

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
		collisionRejectGroup:   info.CollisionRejectGroup,
		collisionSpheres:       info.CollisionSpheres,
		collisionBoxes:         info.CollisionBoxes,
		collisionMeshes:        info.CollisionMeshes,
		aerodynamicShapes:      info.AerodynamicShapes,
	}
}

// CreateScene creates a new Scene and configures
// the simulation for it to run at maximum stepSeconds
// intervals.
func (e *Engine) CreateScene() *Scene {
	return newScene(e)
}
