package asset

import "github.com/mokiat/gomath/sprec"

type Material struct {
	Name            string
	Type            MaterialType
	BackfaceCulling bool
	AlphaTesting    bool
	AlphaThreshold  float32
	Blending        bool
	Scalars         [16]*float32
	Textures        [16]string
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

// pbr material view getter for the base color
func (v *PBRMaterialView) BaseColor() sprec.Vec4 {
	return sprec.NewVec4(
		*v.delegate.Scalars[0],
		*v.delegate.Scalars[1],
		*v.delegate.Scalars[2],
		*v.delegate.Scalars[3],
	)
}
