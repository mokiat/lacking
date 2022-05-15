package asset

import (
	"fmt"
	"io"
)

const (
	WrapModeUnspecified WrapMode = iota
	WrapModeRepeat
	WrapModeMirroredRepeat
	WrapModeClampToEdge
	WrapModeMirroredClampToEdge
)

type WrapMode uint8

const (
	FilterModeUnspecified FilterMode = iota
	FilterModeNearest
	FilterModeLinear
	FilterModeNearestMipmapNearest
	FilterModeNearestMipmapLinear
	FilterModeLinearMipmapNearest
	FilterModeLinearMipmapLinear
)

type FilterMode uint8

const (
	TexelFormatUnspecified TexelFormat = iota
	TexelFormatR8
	TexelFormatR16
	TexelFormatR32F
	TexelFormatRG8
	TexelFormatRG16
	TexelFormatRG32F
	TexelFormatRGB8
	TexelFormatRGB16
	TexelFormatRGB32F
	TexelFormatRGBA8
	TexelFormatRGBA16
	TexelFormatRGBA16F
	TexelFormatRGBA32F
	TexelFormatDepth32F
)

type TexelFormat uint32

type TwoDTexture struct {
	Width     uint16
	Height    uint16
	WrapModeS WrapMode
	WrapModeT WrapMode
	MagFilter FilterMode
	MinFilter FilterMode
	Format    TexelFormat
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
	MagFilter  FilterMode
	MinFilter  FilterMode
	Format     TexelFormat
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
