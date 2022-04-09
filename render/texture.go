package render

type WrapMode int

const (
	WrapModeClamp WrapMode = iota
	WrapModeRepeat
	WrapModeMirroredRepeat
)

type FilterMode int

const (
	FilterModeNearest FilterMode = iota
	FilterModeLinear
	FilterModeAnisotropic
)

type DataFormat int

const (
	DataFormatRGBA8 DataFormat = iota
	DataFormatRGBA32F
)

type TextureObject interface {
	_isTextureObject() bool // ensures interface uniqueness
}

type Texture interface {
	TextureObject
	Release()
}

type ColorTexture2DInfo struct {
	Width           int
	Height          int
	Wrapping        WrapMode
	Filtering       FilterMode
	Mipmapping      bool
	GammaCorrection bool
	Format          DataFormat
	Data            []byte
}

type ColorTextureCubeInfo struct {
	Dimension       int
	Filtering       FilterMode
	Mipmapping      bool
	GammaCorrection bool
	Format          DataFormat
	FrontSideData   []byte
	BackSideData    []byte
	LeftSideData    []byte
	RightSideData   []byte
	TopSideData     []byte
	BottomSideData  []byte
}

type DepthTexture2DInfo struct {
	Width  int
	Height int
}

type StencilTexture2DInfo struct {
	Width  int
	Height int
}

type DepthStencilTexture2DInfo struct {
	Width  int
	Height int
}
