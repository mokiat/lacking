package ecs

// NewEngine creates a new ECS engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Engine is the entrypoint to working with an
// Entity-Component System framework.
type Engine struct{}

// CreateScene creates a new Scene instance.
// Entities within a scene are isolated from
// entities in other scenes.
func (e *Engine) CreateScene() *Scene {
	return newScene()
}
