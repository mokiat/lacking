package asset

import (
	"fmt"
	"io"
)

const (
	WrapModeRepeat WrapMode = iota
	WrapModeMirroredRepeat
	WrapModeClampToEdge
)

type WrapMode uint8

const (
	FilterModeNearest FilterMode = iota
	FilterModeLinear
	FilterModeAnisotropic
)

type FilterMode uint8

const (
	TexelFormatR8 TexelFormat = iota
	TexelFormatR16
	TexelFormatR16F
	TexelFormatR32F
	TexelFormatRG8
	TexelFormatRG16
	TexelFormatRG16F
	TexelFormatRG32F
	TexelFormatRGB8
	TexelFormatRGB16
	TexelFormatRGB16F
	TexelFormatRGB32F
	TexelFormatRGBA8
	TexelFormatRGBA16
	TexelFormatRGBA16F
	TexelFormatRGBA32F
	TexelFormatDepth16F
	TexelFormatDepth32F
)

type TexelFormat uint8

const (
	TextureFlagMipmapping TextureFlag = 1 << iota
	TextureFlagLinear

	TextureFlagNone TextureFlag = 0
)

type TextureFlag uint8

func (f TextureFlag) Has(flag TextureFlag) bool {
	return f&flag == flag
}

type TwoDTexture struct {
	Width     uint16
	Height    uint16
	Wrapping  WrapMode
	Filtering FilterMode
	Format    TexelFormat
	Flags     TextureFlag
	Data      []byte
}

func (t *TwoDTexture) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, t)
}

func (t *TwoDTexture) DecodeFrom(in io.Reader) error {
	return decodeResource(in, t)
}

func (t *TwoDTexture) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return t.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (t *TwoDTexture) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return t.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

type CubeTextureSide struct {
	Data []byte
}

type CubeTexture struct {
	Dimension  uint16
	Filtering  FilterMode
	Format     TexelFormat
	Flags      TextureFlag
	FrontSide  CubeTextureSide
	BackSide   CubeTextureSide
	LeftSide   CubeTextureSide
	RightSide  CubeTextureSide
	TopSide    CubeTextureSide
	BottomSide CubeTextureSide
}

func (t *CubeTexture) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, t)
}

func (t *CubeTexture) DecodeFrom(in io.Reader) error {
	return decodeResource(in, t)
}

func (t *CubeTexture) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return t.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (t *CubeTexture) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return t.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}
