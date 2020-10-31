package asset

import "io"

type DataFormat uint32

const (
	DataFormatRGBA8 DataFormat = iota
	DataFormatRGBA32F
)

type TwoDTexture struct {
	Width  uint16
	Height uint16
	Format DataFormat
	Data   []byte
}

type TextureSide int

const (
	TextureSideFront TextureSide = iota
	TextureSideBack
	TextureSideLeft
	TextureSideRight
	TextureSideTop
	TextureSideBottom
)

type CubeTexture struct {
	Dimension uint16
	Format    DataFormat
	Sides     [6]CubeTextureSide
}

type CubeTextureSide struct {
	Data []byte
}

func EncodeTwoDTexture(out io.Writer, texture *TwoDTexture) error {
	return Encode(out, texture)
}

func DecodeTwoDTexture(in io.Reader, texture *TwoDTexture) error {
	return Decode(in, texture)
}

func EncodeCubeTexture(out io.Writer, texture *CubeTexture) error {
	return Encode(out, texture)
}

func DecodeCubeTexture(in io.Reader, texture *CubeTexture) error {
	return Decode(in, texture)
}
