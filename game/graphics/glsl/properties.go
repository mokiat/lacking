package glsl

type VersionProperties struct {
	Version        string
	NeedsPrecision bool
}

type AttributeProperties struct {
	HasAttributeCoord    bool
	HasAttributeNormal   bool
	HasAttributeTangent  bool
	HasAttributeTexCoord bool
	HasAttributeColor    bool
	HasAttributeArmature bool
}

type OutputProperties struct {
	HasFramebufferOutput0 bool
	HasFramebufferOutput1 bool
	HasFramebufferOutput2 bool
	HasFramebufferOutput3 bool
}

type TextureProperties struct {
	Textures []TextureProperty
}

type TextureProperty struct {
	Name string
	Type string
}

type UniformProperties struct {
	Uniforms []UniformProperty
}

type UniformProperty struct {
	Name string
	Type string
}

type VaryingProperties struct {
	Varyings []VaryingProperty
}

type VaryingProperty struct {
	Name      string
	Type      string
	Direction string
}

type MainProperties struct {
	MainStatements []string
}

type BaseProperties struct {
	VersionProperties
	AttributeProperties
	OutputProperties
	TextureProperties
	UniformProperties
	VaryingProperties
}

type ShadowProperties struct {
	BaseProperties
	MainProperties
}

type GeometryProperties struct {
	BaseProperties
	MainProperties
}

type SkyProperties struct {
	BaseProperties
	MainProperties
}

type ForwardProperties struct {
	BaseProperties
	MainProperties
}
