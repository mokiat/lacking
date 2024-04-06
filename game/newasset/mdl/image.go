package mdl

import (
	"fmt"
	"math"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/dprec"
)

type Color struct {
	R float64
	G float64
	B float64
	A float64
}

const (
	CubeSideFront CubeSide = iota
	CubeSideRear
	CubeSideLeft
	CubeSideRight
	CubeSideTop
	CubeSideBottom
)

type CubeSide uint8

func NewImage(width, height int) *Image {
	texels := make([][]Color, height)
	for y := 0; y < height; y++ {
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
	texels [][]Color
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

func (i *Image) Texel(x, y int) Color {
	return i.texels[y][x]
}

func (i *Image) SetTexel(x, y int, texel Color) {
	i.texels[y][x] = texel
}

func (i *Image) TexelUV(uv dprec.Vec2) Color {
	return i.Texel(
		int(uv.X*float64(i.width-1)),
		int((1.0-uv.Y)*float64(i.height-1)),
	)
}

func (i *Image) BilinearTexel(x, y float64) Color {
	floorX := math.Floor(x)
	ceilX := math.Ceil(x)
	floorY := math.Floor(y)
	ceilY := math.Ceil(y)
	fractX := x - floorX
	fractY := y - floorY

	topLeft := i.texels[int(floorY)][int(floorX)]
	topRight := i.texels[int(floorY)][int(ceilX)]
	bottomLeft := i.texels[int(ceilY)][int(floorX)]
	bottomRight := i.texels[int(ceilY)][int(ceilX)]

	return Color{
		R: topLeft.R*(1.0-fractX)*(1.0-fractY) +
			topRight.R*fractX*(1.0-fractY) +
			bottomLeft.R*(1.0-fractX)*fractY +
			bottomRight.R*fractX*fractY,

		G: topLeft.G*(1.0-fractX)*(1.0-fractY) +
			topRight.G*fractX*(1.0-fractY) +
			bottomLeft.G*(1.0-fractX)*fractY +
			bottomRight.G*fractX*fractY,

		B: topLeft.B*(1.0-fractX)*(1.0-fractY) +
			topRight.B*fractX*(1.0-fractY) +
			bottomLeft.B*(1.0-fractX)*fractY +
			bottomRight.B*fractX*fractY,

		A: topLeft.A*(1.0-fractX)*(1.0-fractY) +
			topRight.A*fractX*(1.0-fractY) +
			bottomLeft.A*(1.0-fractX)*fractY +
			bottomRight.A*fractX*fractY,
	}
}

func (i *Image) Scale(newWidth, newHeight int) *Image {
	// FIXME: Do proper bilinear scaling
	if newWidth < i.width/2 {
		image := i.Scale(i.width/2, i.height)
		return image.Scale(newWidth, newHeight)
	}
	if newHeight < i.height/2 {
		image := i.Scale(i.width, i.height/2)
		return image.Scale(newWidth, newHeight)
	}

	newTexels := make([][]Color, newHeight)
	for y := 0; y < newHeight; y++ {
		newTexels[y] = make([]Color, newWidth)
		oldY := float64(y) * (float64(i.height-1) / float64(newHeight-1))
		for x := 0; x < newWidth; x++ {
			oldX := float64(x) * (float64(i.width-1) / float64(newWidth-1))
			newTexels[y][x] = i.BilinearTexel(oldX, oldY)
		}
	}
	return &Image{
		width:  newWidth,
		height: newHeight,
		texels: newTexels,
	}
}

func (i *Image) RGBA8Data() []byte {
	data := make([]byte, 4*i.width*i.height)
	offset := 0
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			texel := i.Texel(x, i.height-y-1)
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
	data := gblob.LittleEndianBlock(make([]byte, 4*4*i.width*i.height))
	offset := 0
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			texel := i.Texel(x, i.height-y-1)
			data.SetFloat32(offset+0, float32(texel.R))
			data.SetFloat32(offset+4, float32(texel.G))
			data.SetFloat32(offset+8, float32(texel.B))
			data.SetFloat32(offset+12, float32(texel.A))
			offset += 16
		}
	}
	return data
}

func CubeUVToUVW(side CubeSide, uv dprec.Vec2) dprec.Vec3 {
	switch side {
	case CubeSideFront:
		return dprec.UnitVec3(dprec.NewVec3(uv.X*2.0-1.0, uv.Y*2.0-1.0, 1.0))
	case CubeSideRear:
		return dprec.UnitVec3(dprec.NewVec3(1.0-uv.X*2.0, uv.Y*2.0-1.0, -1.0))
	case CubeSideLeft:
		return dprec.UnitVec3(dprec.NewVec3(-1.0, uv.Y*2.0-1.0, uv.X*2.0-1.0))
	case CubeSideRight:
		return dprec.UnitVec3(dprec.NewVec3(1.0, uv.Y*2.0-1.0, 1.0-uv.X*2.0))
	case CubeSideTop:
		return dprec.UnitVec3(dprec.NewVec3(uv.X*2.0-1.0, 1.0, 1.0-uv.Y*2.0))
	case CubeSideBottom:
		return dprec.UnitVec3(dprec.NewVec3(uv.X*2.0-1.0, -1.0, uv.Y*2.0-1.0))
	default:
		panic(fmt.Errorf("unknown cube side: %d", side))
	}
}

func UVWToCubeUV(uvw dprec.Vec3) (CubeSide, dprec.Vec2) {
	if dprec.Abs(uvw.X) >= dprec.Abs(uvw.Y) && dprec.Abs(uvw.X) >= dprec.Abs(uvw.Z) {
		uv := dprec.Vec2Quot(dprec.NewVec2(-uvw.Z, uvw.Y), dprec.Abs(uvw.X))
		if uvw.X > 0 {
			return CubeSideRight, dprec.NewVec2(uv.X/2.0+0.5, uv.Y/2.0+0.5)
		} else {
			return CubeSideLeft, dprec.NewVec2(0.5-uv.X/2.0, uv.Y/2.0+0.5)
		}
	}
	if dprec.Abs(uvw.Z) >= dprec.Abs(uvw.X) && dprec.Abs(uvw.Z) >= dprec.Abs(uvw.Y) {
		uv := dprec.Vec2Quot(dprec.NewVec2(uvw.X, uvw.Y), dprec.Abs(uvw.Z))
		if uvw.Z > 0 {
			return CubeSideFront, dprec.NewVec2(uv.X/2.0+0.5, uv.Y/2.0+0.5)
		} else {
			return CubeSideRear, dprec.NewVec2(0.5-uv.X/2.0, uv.Y/2.0+0.5)
		}
	}
	uv := dprec.Vec2Quot(dprec.NewVec2(uvw.X, uvw.Z), dprec.Abs(uvw.Y))
	if uvw.Y > 0 {
		return CubeSideTop, dprec.NewVec2(uv.X/2.0+0.5, 0.5-uv.Y/2.0)
	} else {
		return CubeSideBottom, dprec.NewVec2(uv.X/2.0+0.5, uv.Y/2.0+0.5)
	}
}

func UVWToEquirectangularUV(uvw dprec.Vec3) dprec.Vec2 {
	return dprec.NewVec2(
		0.5+(0.5/dprec.Pi)*math.Atan2(uvw.Z, uvw.X),
		0.5+(1.0/dprec.Pi)*math.Asin(uvw.Y),
	)
}
