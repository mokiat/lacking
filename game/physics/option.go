package physics

import "time"

// Option is a configuration function that can be used to customize the
// behavior of a physics engine.
type Option func(c *config)

// WithTimestep configures the physics engine to use the provided timestep.
func WithTimestep(timestep time.Duration) Option {
	return func(c *config) {
		c.Timestep = timestep
	}
}

type config struct {
	Timestep time.Duration
}
