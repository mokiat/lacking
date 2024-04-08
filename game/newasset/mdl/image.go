package mdl

import (
	"fmt"
	"math"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/dprec"
	"github.com/x448/float16"
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

func (i *Image) CopyFrom(src *Image) {
	if src.width != i.width || src.height != i.height {
		src = src.Scale(i.width, i.height)
	}
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			i.texels[y][x] = src.texels[y][x]
		}
	}
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

func NewCubeImage(dimension int) *CubeImage {
	frontImage := NewImage(dimension, dimension)
	backImage := NewImage(dimension, dimension)
	leftImage := NewImage(dimension, dimension)
	rightImage := NewImage(dimension, dimension)
	topImage := NewImage(dimension, dimension)
	bottomImage := NewImage(dimension, dimension)
	return &CubeImage{
		dimension: dimension,
		sides: [6]CubeImageSide{
			{texels: frontImage.texels},
			{texels: backImage.texels},
			{texels: leftImage.texels},
			{texels: rightImage.texels},
			{texels: topImage.texels},
			{texels: bottomImage.texels},
		},
	}
}

type CubeImage struct {
	dimension int
	sides     [6]CubeImageSide
}

func (i CubeImage) TexelUVW(uvw dprec.Vec3) Color {
	side, uv := UVWToCubeUV(uvw)
	image := i.SideToImage(side)
	return image.TexelUV(uv)
}

func (i *CubeImage) SideToImage(side CubeSide) *Image {
	return &Image{
		width:  i.dimension,
		height: i.dimension,
		texels: i.sides[side].texels,
	}
}

func (i *CubeImage) SetSide(side CubeSide, image *Image) {
	dstImage := i.SideToImage(side)
	dstImage.CopyFrom(image)
}

func (t *CubeImage) Scale(newDimension int) *CubeImage {
	result := &CubeImage{
		dimension: newDimension,
	}
	for i := range t.sides {
		tmpImage := t.SideToImage(CubeSide(i))
		scaledImage := tmpImage.Scale(newDimension, newDimension)
		result.sides[i] = CubeImageSide{
			texels: scaledImage.texels,
		}
	}
	return result
}

func (t *CubeImage) RGBA8Data(side CubeSide) []byte {
	data := make([]byte, 4*t.dimension*t.dimension)
	offset := 0
	texSide := t.sides[side]
	for y := 0; y < t.dimension; y++ {
		for x := 0; x < t.dimension; x++ {
			texel := texSide.Texel(x, t.dimension-y-1)
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
	data := gblob.LittleEndianBlock(make([]byte, 2*4*t.dimension*t.dimension))
	offset := 0
	texSide := t.sides[side]
	for y := 0; y < t.dimension; y++ {
		for x := 0; x < t.dimension; x++ {
			texel := texSide.Texel(x, t.dimension-y-1)
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
	data := gblob.LittleEndianBlock(make([]byte, 4*4*t.dimension*t.dimension))
	offset := 0
	texSide := t.sides[side]
	for y := 0; y < t.dimension; y++ {
		for x := 0; x < t.dimension; x++ {
			texel := texSide.Texel(x, t.dimension-y-1)
			data.SetFloat32(offset+0, float32(texel.R))
			data.SetFloat32(offset+4, float32(texel.G))
			data.SetFloat32(offset+8, float32(texel.B))
			data.SetFloat32(offset+12, float32(texel.A))
			offset += 16
		}
	}
	return data
}

type CubeImageSide struct {
	texels [][]Color
}

func (s CubeImageSide) Texel(x, y int) Color {
	return s.texels[y][x]
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

func BuildCubeSideFromEquirectangular(side CubeSide, srcImage *Image) *Image {
	dimension := srcImage.Height() / 2
	dstImage := NewImage(dimension, dimension)

	uv := dprec.ZeroVec2()
	startU := 0.0
	deltaU := 1.0 / float64(dimension-1)
	startV := 1.0
	deltaV := -1.0 / float64(dimension-1)
	uv.Y = startV
	for y := 0; y < dimension; y++ {
		uv.X = startU
		for x := 0; x < dimension; x++ {
			dstImage.SetTexel(x, y, srcImage.TexelUV(UVWToEquirectangularUV(CubeUVToUVW(side, uv))))
			uv.X += deltaU
		}
		uv.Y += deltaV
	}
	return dstImage
}

func BuildIrradianceCubeImage(srcImage *CubeImage, sampleCount int) *CubeImage {
	dimension := srcImage.dimension

	dstImage := &CubeImage{
		dimension: dimension,
	}
	for i := range srcImage.sides {
		texels := make([][]Color, dimension)
		for y := range texels {
			texels[y] = make([]Color, dimension)
		}

		uv := dprec.ZeroVec2()
		startU := 0.0
		deltaU := 1.0 / float64(dimension-1)
		startV := 1.0
		deltaV := -1.0 / float64(dimension-1)

		uv.Y = startV
		for y := 0; y < dimension; y++ {
			uv.X = startU

			for x := 0; x < dimension; x++ {
				uvw := CubeUVToUVW(CubeSide(i), uv)
				startLat := dprec.Radians(-dprec.Pi / 2.0)
				endLat := dprec.Radians(dprec.Pi / 2.0)
				deltaLat := (endLat - startLat) / dprec.Radians(float64(sampleCount))

				color := Color{}
				positiveSamples := 0.0
				for lat := startLat; lat < endLat; lat += deltaLat {
					startLong := dprec.Radians(-dprec.Pi)
					endLong := dprec.Radians(dprec.Pi)
					deltaLong := (endLong - startLong) / (dprec.Radians(float64(sampleCount) * (dprec.Cos(lat) + 0.01)))

					for long := startLong; long < endLong; long += deltaLong {
						flatX := dprec.Sin(long) * dprec.Cos(lat)
						flatY := dprec.Cos(long) * dprec.Cos(lat)
						targetUVW := dprec.NewVec3(flatX, flatY, dprec.Sin(lat))
						if dot := dprec.Vec3Dot(uvw, targetUVW); dot > 0.0 {
							positiveSamples++
							targetColor := srcImage.TexelUVW(targetUVW)
							color.R += targetColor.R * dot
							color.G += targetColor.G * dot
							color.B += targetColor.B * dot
						}
					}
				}

				texels[y][x] = Color{
					R: dprec.Pi * color.R / positiveSamples,
					G: dprec.Pi * color.G / positiveSamples,
					B: dprec.Pi * color.B / positiveSamples,
					A: 1.0,
				}

				uv.X += deltaU
			}
			uv.Y += deltaV
		}

		dstImage.sides[i] = CubeImageSide{
			texels: texels,
		}
	}
	return dstImage
}
