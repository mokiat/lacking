package asset

import "github.com/mokiat/gomath/sprec"

type Material struct {
	Name            string
	Type            MaterialType
	BackfaceCulling bool
	AlphaTesting    bool
	AlphaThreshold  float32
	Blending        bool
	ScalarMask      uint32
	Scalars         [16]float32
	Textures        [16]TextureRef
}

const (
	MaterialTypePBR MaterialType = iota
	MaterialTypeAlbedo
)

type MaterialType uint8

func NewPBRMaterialView(delegate *Material) *PBRMaterialView {
	return &PBRMaterialView{
		delegate: delegate,
	}
}

type PBRMaterialView struct {
	delegate *Material
}

func (v *PBRMaterialView) BaseColor() sprec.Vec4 {
	return sprec.NewVec4(
		v.delegate.Scalars[0],
		v.delegate.Scalars[1],
		v.delegate.Scalars[2],
		v.delegate.Scalars[3],
	)
}

func (v *PBRMaterialView) SetBaseColor(color sprec.Vec4) {
	v.delegate.Scalars[0] = color.X
	v.delegate.Scalars[1] = color.Y
	v.delegate.Scalars[2] = color.Z
	v.delegate.Scalars[3] = color.W
}

func (v *PBRMaterialView) Metallic() float32 {
	return v.delegate.Scalars[4]
}

func (v *PBRMaterialView) SetMetallic(metallic float32) {
	v.delegate.Scalars[4] = metallic
}

func (v *PBRMaterialView) Roughness() float32 {
	return v.delegate.Scalars[5]
}

func (v *PBRMaterialView) SetRoughness(roughness float32) {
	v.delegate.Scalars[5] = roughness
}

func (v *PBRMaterialView) NormalScale() float32 {
	return v.delegate.Scalars[6]
}

func (v *PBRMaterialView) SetNormalScale(scale float32) {
	v.delegate.Scalars[6] = scale
}

func (v *PBRMaterialView) BaseColorTexture() TextureRef {
	return v.delegate.Textures[0]
}

func (v *PBRMaterialView) SetBaseColorTexture(texture TextureRef) {
	v.delegate.Textures[0] = texture
}

func (v *PBRMaterialView) MetallicRoughnessTexture() TextureRef {
	return v.delegate.Textures[1]
}

func (v *PBRMaterialView) SetMetallicRoughnessTexture(texture TextureRef) {
	v.delegate.Textures[1] = texture
}

func (v *PBRMaterialView) NormalTexture() TextureRef {
	return v.delegate.Textures[2]
}

func (v *PBRMaterialView) SetNormalTexture(texture TextureRef) {
	v.delegate.Textures[2] = texture
}
