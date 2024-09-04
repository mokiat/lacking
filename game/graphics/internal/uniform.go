package internal

const (
	UniformBufferBindingCamera      = 0
	UniformBufferBindingModel       = 1
	UniformBufferBindingMaterial    = 2
	UniformBufferBindingArmature    = 3
	UniformBufferBindingLight       = 4
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
