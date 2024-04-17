package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/render"
)

// GeometryRenderPassInfo contains the information representing the rendering
// behavior of a material during the geometry pass.
type GeometryRenderPassInfo struct {

	// Layer controls the render ordering of this pass. Lower values will be
	// rendered first. Having too many layers can affect performance.
	Layer int32

	// Culling specifies the culling mode.
	Culling opt.T[render.CullMode]

	// FrontFace specifies the front face orientation.
	FrontFace opt.T[render.FaceOrientation]

	// DepthTest specifies whether depth testing should be enabled.
	DepthTest opt.T[bool]

	// DepthWrite specifies whether depth writing should be enabled.
	DepthWrite opt.T[bool]

	// DepthComparison specifies the depth comparison function.
	DepthComparison opt.T[render.Comparison]

	// Shader is the geometry shader that will be used to render the material.
	Shader *GeometryShader
}

// ShadowRenderPassInfo contains the information representing the rendering
// behavior of a material during the shadow pass.
type ShadowRenderPassInfo struct {

	// Culling specifies the culling mode.
	Culling opt.T[render.CullMode]

	// FrontFace specifies the front face orientation.
	FrontFace opt.T[render.FaceOrientation]

	// Shader is the shadow shader that will be used to render the material.
	Shader *ShadowShader
}

// ForwardRenderPassInfo contains the information representing the rendering
// behavior of a material during the forward pass.
type ForwardRenderPassInfo struct {

	// Layer controls the render ordering of this pass. Lower values will be
	// rendered first. Having too many layers can affect performance.
	Layer int32

	// Culling specifies the culling mode.
	Culling opt.T[render.CullMode]

	// FrontFace specifies the front face orientation.
	FrontFace opt.T[render.FaceOrientation]

	// DepthTest specifies whether depth testing should be enabled.
	DepthTest opt.T[bool]

	// DepthWrite specifies whether depth writing should be enabled.
	DepthWrite opt.T[bool]

	// DepthComparison specifies the depth comparison function.
	DepthComparison opt.T[render.Comparison]

	// Blending specifies whether the output will be mixed with the
	// background. Useful for unlit/emissive special effects.
	Blending opt.T[bool]

	// Shader is the forward shader that will be used to render the material.
	Shader *ForwardShader
}

// TextureBindingInfo contains the information needed to bind a texture to a
// material.
type TextureBindingInfo struct {

	// Texture specifies the texture to be used.
	Texture render.Texture

	// Wrapping specifies the texture wrapping mode.
	Wrapping render.WrapMode

	// Filtering specifies the texture filtering mode.
	Filtering render.FilterMode

	// Mipmapping specifies whether mipmapping should be enabled and whether
	// mipmaps should be generated.
	Mipmapping bool
}
