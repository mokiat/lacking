package asset

import (
	"fmt"
	"io"

	newasset "github.com/mokiat/lacking/game/newasset"
)

const UnspecifiedIndex = int32(-1)

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

type TwoDTexture struct {
	Width  uint16
	Height uint16
	Format TexelFormat
	Flags  newasset.TextureFlag
	Data   []byte
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
	Format     TexelFormat
	Flags      newasset.TextureFlag
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

type TextureRef struct {
	TextureIndex int32
	TextureID    string
}

func (r TextureRef) Valid() bool {
	return r.TextureID != "" || r.TextureIndex >= 0
}
