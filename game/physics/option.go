package physics

// Option is a configuration function that can be used to customize the
// behavior of a physics engine.
type Option func(c *config)

type config struct{}
