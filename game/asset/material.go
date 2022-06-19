package asset

type Material struct {
	Name            string
	Type            MaterialType
	BackfaceCulling bool
	AlphaTesting    bool
	AlphaThreshold  bool
	Translucent     bool
	Scalars         [16]*float32
	Textures        [16]string
}

const (
	MaterialTypePBR MaterialType = iota
	MaterialTypeAlbedo
)

type MaterialType uint8
