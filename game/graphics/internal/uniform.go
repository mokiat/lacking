package internal

const (
	UniformBufferBindingCamera = 0

	UniformBufferBindingModel    = 1
	UniformBufferBindingMaterial = 2
	UniformBufferBindingArmature = 3

	UniformBufferBindingLight           = 4
	UniformBufferBindingLightProperties = 5

	UniformBufferBindingSkybox = 1

	UniformBufferBindingPostprocess = 0
)

const (
	TextureBindingGeometryAlbedoTexture = 0

	TextureBindingLightingFramebufferColor0 = 0
	TextureBindingLightingFramebufferColor1 = 1
	TextureBindingLightingFramebufferColor2 = 2
	TextureBindingLightingFramebufferDepth  = 3
	TextureBindingShadowFramebufferDepth    = 4
	TextureBindingLightingReflectionTexture = 4
	TextureBindingLightingRefractionTexture = 5

	TextureBindingPostprocessFramebufferColor0 = 0

	TextureBindingSkyboxAlbedoTexture = 0
)
