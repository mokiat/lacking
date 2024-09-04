package graphics

// Option is a configuration function that can be used to customize the
// behavior of the graphics engine.
type Option func(*config)

// WithStageBuilder configures the graphics engine to use the specified stage
// builder function.
func WithStageBuilder(builder StageBuilderFunc) Option {
	return func(c *config) {
		c.StageBuilder = builder
	}
}

// WithDirectionalShadowMapCount configures the graphics engine to use the
// specified number of directional shadow maps.
//
// This value controls the number of directional lights that can have shadows
// at the same time.
func WithDirectionalShadowMapCount(count int) Option {
	return func(c *config) {
		c.DirectionalShadowMapCount = count
	}
}

// WithDirectionalShadowMapSize configures the graphics engine to use the
// specified size for the directional shadow maps. The size needs to be a power
// of two.
func WithDirectionalShadowMapSize(size int) Option {
	return func(c *config) {
		c.DirectionalShadowMapSize = size
	}
}

// WithDirectionalShadowMapCascadeCount configures the maximum number of
// cascades that the directional shadow maps will have.
//
// This value cannot be smaller than 1 and larger than 8 and will be clamped.
func WithDirectionalShadowMapCascadeCount(count int) Option {
	return func(c *config) {
		c.DirectionalShadowMapCascadeCount = max(1, min(count, 8))
	}
}

// WithSpotShadowMapCount configures the graphics engine to use the specified
// number of spot light shadow maps.
//
// This value controls the number of spot lights that can have shadows at the
// same time.
func WithSpotShadowMapCount(count int) Option {
	return func(c *config) {
		c.SpotShadowMapCount = count
	}
}

// WithSpotShadowMapSize configures the graphics engine to use the specified
// size for the spot light shadow maps. The size needs to be a power of two.
func WithSpotShadowMapSize(size int) Option {
	return func(c *config) {
		c.SpotShadowMapSize = size
	}
}

// WithPointShadowMapCount configures the graphics engine to use the specified
// number of point light shadow maps.
//
// This value controls the number of point lights that can have shadows at the
// same time.
func WithPointShadowMapCount(count int) Option {
	return func(c *config) {
		c.PointShadowMapCount = count
	}
}

// WithPointShadowMapSize configures the graphics engine to use the specified
// size for the point light shadow maps. The size needs to be a power of two.
func WithPointShadowMapSize(size int) Option {
	return func(c *config) {
		c.PointShadowMapSize = size
	}
}

type config struct {
	StageBuilder StageBuilderFunc

	DirectionalShadowMapCount        int
	DirectionalShadowMapSize         int
	DirectionalShadowMapCascadeCount int

	SpotShadowMapCount int
	SpotShadowMapSize  int

	PointShadowMapCount int
	PointShadowMapSize  int
}
