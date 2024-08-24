package graphics

// Option is a configuration function that can be used to customize the
// behavior of the graphics engine.
type Option func(*config)

// WithStageBuilder configures the graphics engine to use the specified stage
// builder function.
func WithStageBuilder(builder StageBuilderFunc) Option {
	return func(c *config) {
		c.stageBuilder = builder
	}
}

type config struct {
	stageBuilder StageBuilderFunc
}
