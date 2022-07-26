package physics

// NewEngine creates a new physics engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Engine is the entrypoint to working with a
// physics simulation.
type Engine struct{}

// CreateScene creates a new Scene and configures
// the simulation for it to run at maximum stepSeconds
// intervals.
func (e *Engine) CreateScene(stepSeconds float64) *Scene {
	return newScene(stepSeconds)
}
