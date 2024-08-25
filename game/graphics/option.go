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

// WithCascadeShadowMapSize configures the graphics engine to use the specified
// size for each cascade shadow map. The size needs to be a power of two.
func WithCascadeShadowMapSize(size int) Option {
	return func(c *config) {
		c.CascadeShadowMapSize = size
	}
}

// WithCascadeShadowMapCount configures the graphics engine to use the specified
// number of cascade shadow maps.
//
// Note: Since cascade shadow maps are used by directional lights that currently
// have three predefined cascades, the count should be a multiple of three in
// order to be able to distribute the cascades evenly. This may change in the
// future.
func WithCascadeShadowMapCount(count int) Option {
	return func(c *config) {
		c.CascadeShadowMapCount = count
	}
}

// WithAtlasShadowMapSize configures the graphics engine to use the specified
// size for the atlas shadow map. The size needs to be a power of two.
func WithAtlasShadowMapSize(size int) Option {
	return func(c *config) {
		c.AtlasShadowMapSize = size
	}
}

// WithAtlasShadowMapSectors configures the graphics engine to use the specified
// number of sectors in the atlas shadow map.
//
// The sectors need to be a power of four (e.g. 1, 4, 16, 64).
//
// Making more sectors will result in more lights being able to cast shadows
// at the same time at the cost of quality.
func WithAtlasShadowMapSectors(count int) Option {
	return func(c *config) {
		c.AtlasShadowMapSectors = count
	}
}

type config struct {
	StageBuilder StageBuilderFunc

	CascadeShadowMapSize  int
	CascadeShadowMapCount int
	AtlasShadowMapSize    int
	AtlasShadowMapSectors int
}
