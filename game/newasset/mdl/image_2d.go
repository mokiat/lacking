package mdl

import (
	"math"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/dprec"
)

func NewImage(width, height int) *Image {
	texels := make([][]Color, height)
	for y := range texels {
		texels[y] = make([]Color, width)
	}
	return &Image{
		width:  width,
		height: height,
		texels: texels,
	}
}

type Image struct {
	width  int
	height int
	texels [][]Color // TODO: Use single slice
}

func (i *Image) Width() int {
	return i.width
}

func (i *Image) Height() int {
	return i.height
}

func (i *Image) IsSquare() bool {
	return i.width == i.height
}

func (i *Image) IsEqualSize(other *Image) bool {
	return i.width == other.width && i.height == other.height
}

func (i *Image) CopyFrom(src *Image) {
	if !src.IsEqualSize(i) {
		src = src.Scale(i.width, i.height)
	}
	for y := range i.height {
		copy(i.texels[y], src.texels[y])
	}
}

func (i *Image) Texel(x, y int) Color {
	x = max(0, min(x, i.width-1))
	y = max(0, min(y, i.height-1))
	return i.texels[y][x]
}

func (i *Image) SetTexel(x, y int, texel Color) {
	i.texels[y][x] = texel
}

func (i *Image) TexelUV(uv dprec.Vec2) Color {
	x := int(uv.X * float64(i.width))
	y := int((1.0 - uv.Y) * float64(i.height))
	return i.Texel(x, y)
}

func (i *Image) TexelUVBilinear(uv dprec.Vec2) Color {
	x := uv.X*float64(i.width) - 0.5
	y := (1.0-uv.Y)*float64(i.height) - 0.5

	leftX := math.Floor(x)
	topY := math.Floor(y)
	horizontalMix := x - leftX
	verticalMix := y - topY

	topLeftColor := i.Texel(int(leftX), int(topY))
	topRightColor := i.Texel(int(leftX)+1, int(topY))
	bottomLeftColor := i.Texel(int(leftX), int(topY)+1)
	bottomRightColor := i.Texel(int(leftX)+1, int(topY)+1)

	topR := dprec.Mix(topLeftColor.R, topRightColor.R, horizontalMix)
	topG := dprec.Mix(topLeftColor.G, topRightColor.G, horizontalMix)
	topB := dprec.Mix(topLeftColor.B, topRightColor.B, horizontalMix)
	topA := dprec.Mix(topLeftColor.A, topRightColor.A, horizontalMix)
	bottomR := dprec.Mix(bottomLeftColor.R, bottomRightColor.R, horizontalMix)
	bottomG := dprec.Mix(bottomLeftColor.G, bottomRightColor.G, horizontalMix)
	bottomB := dprec.Mix(bottomLeftColor.B, bottomRightColor.B, horizontalMix)
	bottomA := dprec.Mix(bottomLeftColor.A, bottomRightColor.A, horizontalMix)

	return Color{
		R: dprec.Mix(topR, bottomR, verticalMix),
		G: dprec.Mix(topG, bottomG, verticalMix),
		B: dprec.Mix(topB, bottomB, verticalMix),
		A: dprec.Mix(topA, bottomA, verticalMix),
	}
}

func (i *Image) Scale(newWidth, newHeight int) *Image {
	switch {

	case newWidth == i.width && newHeight == i.height:
		result := NewImage(i.width, i.height)
		result.CopyFrom(i)
		return result

	case newWidth < i.width/2 && newHeight < i.height/2:
		intermediate := i.Scale(i.width/2, i.height/2)
		return intermediate.Scale(newWidth, newHeight)

	case newWidth < i.width/2:
		intermediate := i.Scale(i.width/2, i.height)
		return intermediate.Scale(newWidth, newHeight)

	case newHeight < i.height/2:
		intermediate := i.Scale(i.width, i.height/2)
		return intermediate.Scale(newWidth, newHeight)

	default:
		result := NewImage(newWidth, newHeight)
		for y := range newHeight {
			currentV := 1.0 - (float64(y))/float64(newHeight-1)
			for x := range newWidth {
				currentU := (float64(x)) / float64(newWidth-1)
				result.SetTexel(x, y, i.TexelUVBilinear(dprec.NewVec2(currentU, currentV)))
			}
		}
		return result
	}
}

func (i *Image) DataRGBA8() []byte {
	const texelSize = 4
	data := make([]byte, texelSize*i.width*i.height)
	offset := 0
	for y := range i.height {
		for x := range i.width {
			texel := i.Texel(x, i.height-y-1)
			r, g, b, a := texel.RGBA8()
			data[offset+0] = r
			data[offset+1] = g
			data[offset+2] = b
			data[offset+3] = a
			offset += texelSize
		}
	}
	return data
}

func (i *Image) DataRGBA16F() []byte {
	const texelSize = 4 * 2
	data := gblob.LittleEndianBlock(make([]byte, texelSize*i.width*i.height))
	offset := 0
	for y := range i.height {
		for x := range i.width {
			texel := i.Texel(x, i.height-y-1)
			r, g, b, a := texel.RGBA16F()
			data.SetUint16(offset+0, r.Bits())
			data.SetUint16(offset+2, g.Bits())
			data.SetUint16(offset+4, b.Bits())
			data.SetUint16(offset+6, a.Bits())
			offset += texelSize
		}
	}
	return data
}

func (i *Image) DataRGBA32F() []byte {
	const texelSize = 4 * 4
	data := gblob.LittleEndianBlock(make([]byte, texelSize*i.width*i.height))
	offset := 0
	for y := range i.height {
		for x := range i.width {
			texel := i.Texel(x, i.height-y-1)
			r, g, b, a := texel.RGBA32F()
			data.SetFloat32(offset+0, r)
			data.SetFloat32(offset+4, g)
			data.SetFloat32(offset+8, b)
			data.SetFloat32(offset+12, a)
			offset += texelSize
		}
	}
	return data
}
