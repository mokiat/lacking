package pack

import (
	"github.com/mokiat/gblob"
)

// TODO: Replace with mdl.Image

type Color struct {
	R float64
	G float64
	B float64
	A float64
}

type Image struct {
	Name   string
	Width  int
	Height int
	Texels [][]Color
}

func (i *Image) Texel(x, y int) Color {
	return i.Texels[y][x]
}

func (i *Image) RGBA8Data() []byte {
	data := make([]byte, 4*i.Width*i.Height)
	offset := 0
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			texel := i.Texel(x, i.Height-y-1)
			data[offset+0] = byte(255.0 * texel.R)
			data[offset+1] = byte(255.0 * texel.G)
			data[offset+2] = byte(255.0 * texel.B)
			data[offset+3] = byte(255.0 * texel.A)
			offset += 4
		}
	}
	return data
}

func (i *Image) RGBA32FData() []byte {
	data := gblob.LittleEndianBlock(make([]byte, 4*4*i.Width*i.Height))
	offset := 0
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			texel := i.Texel(x, i.Height-y-1)
			data.SetFloat32(offset+0, float32(texel.R))
			data.SetFloat32(offset+4, float32(texel.G))
			data.SetFloat32(offset+8, float32(texel.B))
			data.SetFloat32(offset+12, float32(texel.A))
			offset += 16
		}
	}
	return data
}
