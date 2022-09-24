package pack

import (
	"fmt"
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/data"
	"github.com/x448/float16"
)

type Color struct {
	R float64
	G float64
	B float64
	A float64
}

type ImageProvider interface {
	Image() *Image
}

type Image struct {
	Width  int
	Height int
	Texels [][]Color
}

func (i *Image) IsSquare() bool {
	return i.Width == i.Height
}

func (i *Image) Texel(x, y int) Color {
	return i.Texels[y][x]
}

func (i *Image) TexelUV(uv dprec.Vec2) Color {
	return i.Texel(
		int(uv.X*float64(i.Width-1)),
		int((1.0-uv.Y)*float64(i.Height-1)),
	)
}

func (i *Image) BilinearTexel(x, y float64) Color {
	floorX := math.Floor(x)
	ceilX := math.Ceil(x)
	floorY := math.Floor(y)
	ceilY := math.Ceil(y)
	fractX := x - floorX
	fractY := y - floorY

	topLeft := i.Texels[int(floorY)][int(floorX)]
	topRight := i.Texels[int(floorY)][int(ceilX)]
	bottomLeft := i.Texels[int(ceilY)][int(floorX)]
	bottomRight := i.Texels[int(ceilY)][int(ceilX)]

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
	if newWidth < i.Width/2 {
		image := i.Scale(i.Width/2, i.Height)
		return image.Scale(newWidth, newHeight)
	}
	if newHeight < i.Height/2 {
		image := i.Scale(i.Width, i.Height/2)
		return image.Scale(newWidth, newHeight)
	}

	newTexels := make([][]Color, newHeight)
	for y := 0; y < newHeight; y++ {
		newTexels[y] = make([]Color, newWidth)
		oldY := float64(y) * (float64(i.Height-1) / float64(newHeight-1))
		for x := 0; x < newWidth; x++ {
			oldX := float64(x) * (float64(i.Width-1) / float64(newWidth-1))
			newTexels[y][x] = i.BilinearTexel(oldX, oldY)
		}
	}
	return &Image{
		Width:  newWidth,
		Height: newHeight,
		Texels: newTexels,
	}
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
	data := data.Buffer(make([]byte, 4*4*i.Width*i.Height))
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

type CubeSide int

const (
	CubeSideFront CubeSide = iota
	CubeSideRear
	CubeSideLeft
	CubeSideRight
	CubeSideTop
	CubeSideBottom
)

type CubeImageProvider interface {
	CubeImage() *CubeImage
}

type CubeImage struct {
	Dimension int
	Sides     [6]CubeImageSide
}

func (s CubeImage) TexelUVW(uvw dprec.Vec3) Color {
	side, uv := UVWToCubeUV(uvw)
	image := s.SideToImage(side)
	return image.TexelUV(uv)
}

type CubeImageSide struct {
	Texels [][]Color
}

func (s CubeImageSide) Texel(x, y int) Color {
	return s.Texels[y][x]
}

func (i *CubeImage) SideToImage(side CubeSide) *Image {
	return &Image{
		Width:  i.Dimension,
		Height: i.Dimension,
		Texels: i.Sides[side].Texels,
	}
}

func (t *CubeImage) Scale(newDimension int) *CubeImage {
	result := &CubeImage{
		Dimension: newDimension,
	}
	for i := range t.Sides {
		tmpImage := t.SideToImage(CubeSide(i))
		scaledImage := tmpImage.Scale(newDimension, newDimension)
		result.Sides[i] = CubeImageSide{
			Texels: scaledImage.Texels,
		}
	}
	return result
}

func (t *CubeImage) RGBA8Data(side CubeSide) []byte {
	data := make([]byte, 4*t.Dimension*t.Dimension)
	offset := 0
	texSide := t.Sides[side]
	for y := 0; y < t.Dimension; y++ {
		for x := 0; x < t.Dimension; x++ {
			texel := texSide.Texel(x, t.Dimension-y-1)
			data[offset+0] = byte(255.0 * dprec.Clamp(texel.R, 0.0, 1.0))
			data[offset+1] = byte(255.0 * dprec.Clamp(texel.G, 0.0, 1.0))
			data[offset+2] = byte(255.0 * dprec.Clamp(texel.B, 0.0, 1.0))
			data[offset+3] = byte(255.0 * dprec.Clamp(texel.A, 0.0, 1.0))
			offset += 4
		}
	}
	return data
}

func (t *CubeImage) RGBA16FData(side CubeSide) []byte {
	data := data.Buffer(make([]byte, 2*4*t.Dimension*t.Dimension))
	offset := 0
	texSide := t.Sides[side]
	for y := 0; y < t.Dimension; y++ {
		for x := 0; x < t.Dimension; x++ {
			texel := texSide.Texel(x, t.Dimension-y-1)
			data.SetUint16(offset+0, uint16(float16.Fromfloat32(float32(texel.R))))
			data.SetUint16(offset+2, uint16(float16.Fromfloat32(float32(texel.G))))
			data.SetUint16(offset+4, uint16(float16.Fromfloat32(float32(texel.B))))
			data.SetUint16(offset+6, uint16(float16.Fromfloat32(float32(texel.A))))
			offset += 8
		}
	}
	return data
}

func (t *CubeImage) RGBA32FData(side CubeSide) []byte {
	data := data.Buffer(make([]byte, 4*4*t.Dimension*t.Dimension))
	offset := 0
	texSide := t.Sides[side]
	for y := 0; y < t.Dimension; y++ {
		for x := 0; x < t.Dimension; x++ {
			texel := texSide.Texel(x, t.Dimension-y-1)
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
