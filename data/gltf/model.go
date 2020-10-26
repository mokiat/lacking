package gltf

const (
	AttributePosition  = string("POSITION")
	AttributeNormal    = string("NORMAL")
	AttributeTangent   = string("TANGENT")
	AttributeTexCoord0 = string("TEXCOORD_0")
	AttributeColor0    = string("COLOR_0")
)

const (
	ModePoints        = int(0)
	ModeLines         = int(1)
	ModeLineLoop      = int(2)
	ModeLineStrip     = int(3)
	ModeTriangles     = int(4)
	ModeTriangleStrip = int(5)
	ModeTriangleFan   = int(6)
)

const (
	ComponentTypeByte          = int(5120)
	ComponentTypeUnsignedByte  = int(5121)
	ComponentTypeShort         = int(5122)
	ComponentTypeUnsignedShort = int(5123)
	ComponentTypeUnsignedInt   = int(5125)
	ComponentTypeFloat         = int(5126)
)

const (
	TypeScalar = string("SCALAR")
	TypeVec2   = string("VEC2")
	TypeVec3   = string("VEC3")
	TypeVec4   = string("VEC4")
	TypeMat2   = string("MAT2")
	TypeMat3   = string("MAT3")
	TypeMat4   = string("MAT4")
)

type Document struct {
	Asset       Asset        `json:"asset"`
	Scene       *int         `json:"scene,omitempty"`
	Scenes      []Scene      `json:"scenes"`
	Nodes       []Node       `json:"nodes"`
	Materials   []Material   `json:"materials"`
	Textures    []Texture    `json:"textures"`
	Images      []Image      `json:"images"`
	Meshes      []Mesh       `json:"meshes"`
	Accessors   []Accessor   `json:"accessors"`
	Buffers     []Buffer     `json:"buffers"`
	BufferViews []BufferView `json:"bufferViews"`
}

type Asset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
	Copyright string `json:"copyright"`
}

type Scene struct {
	Name  string `json:"name"`
	Nodes []int  `json:"nodes"`
}

type Node struct {
	Name        string       `json:"name"`
	Children    []int        `json:"children"`
	Translation *Translation `json:"translation"`
	Scale       *Scale       `json:"scale"`
	Rotation    *Rotation    `json:"rotation"`
	Matrix      *Matrix      `json:"matrix"`
	Mesh        *int         `json:"mesh"`
}

type Translation [3]float32

type Scale [3]float32

type Rotation [4]float32

type Matrix [16]float32

type Material struct {
	Name                 string                `json:"name"`
	DoubleSided          bool                  `json:"doubleSided"`
	NormalTexture        *TextureInfo          `json:"normalTexture"`
	PBRMetallicRoughness *PBRMetallicRoughness `json:"pbrMetallicRoughness"`
}

type PBRMetallicRoughness struct {
	BaseColorFactor          *Color       `json:"baseColorFactor"`
	BaseColorTexture         *TextureInfo `json:"baseColorTexture"`
	MetallicFactor           *float32     `json:"metallicFactor"`
	RoughnessFactor          *float32     `json:"roughnessFactor"`
	MetallicRoughnessTexture *TextureInfo `json:"metallicRoughnessTexture"`
}

type Color [4]float32

type Mesh struct {
	Name       string      `json:"name"`
	Primitives []Primitive `json:"primitives"`
}

func (m Mesh) HasAttribute(name string) bool {
	for _, primitive := range m.Primitives {
		if primitive.HasAttribute(name) {
			return true
		}
	}
	return false
}

type Primitive struct {
	Material   *int           `json:"material"`
	Mode       *int           `json:"mode"`
	Indices    *int           `json:"indices"`
	Attributes map[string]int `json:"attributes"`
}

func (p Primitive) HasAttribute(name string) bool {
	_, ok := p.Attributes[name]
	return ok
}

type Accessor struct {
	BufferView    *int   `json:"bufferView"`
	ComponentType int    `json:"componentType"`
	Count         int    `json:"count"`
	Type          string `json:"type"`
}

type Buffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri"`
	Data       []byte `json:"-"`
}

type BufferView struct {
	Buffer     int  `json:"buffer"`
	ByteLength int  `json:"byteLength"`
	ByteOffset int  `json:"byteOffset"`
	ByteStride *int `json:"byteStride"`
}

type TextureInfo struct {
	Index    int      `json:"index"`
	TexCoord int      `json:"texCoord"`
	Scale    *float32 `json:"scale"`
}

type Texture struct {
	Source *int `json:"source"`
}

type Image struct {
	Name string `json:"name"`
}
