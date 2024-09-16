package internal

const (
	// NOTE: The following binding points should not be zero, even if
	// this is not well documented.

	UniformBufferBindingCamera      = 1
	UniformBufferBindingModel       = 2
	UniformBufferBindingMaterial    = 3
	UniformBufferBindingArmature    = 4
	UniformBufferBindingLight       = 5
	UniformBufferBindingPostprocess = 6
	UniformBufferBindingBloom       = 7
)

const (
	TextureBindingGeometryAlbedoTexture = 0

	TextureBindingLightingFramebufferColor0 = 0
	TextureBindingLightingFramebufferColor1 = 1
	TextureBindingLightingFramebufferDepth  = 3
	TextureBindingLightingReflectionTexture = 4
	TextureBindingLightingRefractionTexture = 5

	TextureBindingLightingShadowMap = 4

	TextureBindingPostprocessFramebufferColor0 = 0
	TextureBindingPostprocessBloom             = 1
)
