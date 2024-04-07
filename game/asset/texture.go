package asset

import (
	"fmt"
	"io"

	newasset "github.com/mokiat/lacking/game/newasset"
)

const UnspecifiedIndex = int32(-1)

type TwoDTexture struct {
	Width  uint32
	Height uint32
	Format newasset.TexelFormat
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
	Dimension  uint32
	Format     newasset.TexelFormat
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
}

func (r TextureRef) Valid() bool {
	return r.TextureIndex >= 0
}
