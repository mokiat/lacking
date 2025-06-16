package ecs5

// Option is a configuration function that can be used to customize the
// behavior of the ECS engine.
type Option func(*config)

// WithMaxEntityCount controls the maximum number of entities that a scene
// will need to manage.
//
// Keeping this value small could reduce memory usage and increase performance.
//
// By default it is equal to 1048576 (1024x1024), which is also the maximum that
// this can be set to.
func WithMaxEntityCount(count int) Option {
	return func(cfg *config) {
		cfg.maxEntityCount = max(0, min(count, defaultMaxEntityCount))
	}
}

type config struct {
	maxEntityCount int
}

// NewEngine creates a new ECS engine.
func NewEngine(opts ...Option) *Engine {
	cfg := config{
		maxEntityCount: defaultMaxEntityCount,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Engine{
		maxEntityCount: cfg.maxEntityCount,
	}
}

// Engine is the entrypoint to working with an
// Entity-Component System framework.
type Engine struct {
	maxEntityCount int
}

// CreateScene creates a new Scene instance.
// Entities within a scene are isolated from
// entities in other scenes.
func (e *Engine) CreateScene() *Scene {
	return newScene(e.maxEntityCount)
}
