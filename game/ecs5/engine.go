package ecs5

// Option is a configuration function that can be used to customize the
// behavior of the ECS engine.
type Option func(*config)

type config struct{}

// NewEngine creates a new ECS engine.
func NewEngine(opts ...Option) *Engine {
	return &Engine{}
}

// Engine is the entrypoint to working with an
// Entity-Component System framework.
type Engine struct{}

// CreateScene creates a new Scene instance.
// Entities within a scene are isolated from
// entities in other scenes.
func (e *Engine) CreateScene() *Scene {
	return newScene(defaultMaxEntityCount)
}
