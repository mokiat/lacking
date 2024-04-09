package mdl

import (
	"github.com/mokiat/gomath/dprec"
)

const (
	CubeSideFront CubeSide = iota
	CubeSideRear
	CubeSideLeft
	CubeSideRight
	CubeSideTop
	CubeSideBottom
)

type CubeSide uint8

func NewCubeImage(size int) *CubeImage {
	return &CubeImage{
		size: size,
		sides: [6]*Image{
			NewImage(size, size),
			NewImage(size, size),
			NewImage(size, size),
			NewImage(size, size),
			NewImage(size, size),
			NewImage(size, size),
		},
	}
}

type CubeImage struct {
	size  int
	sides [6]*Image
}

func (i *CubeImage) Side(side CubeSide) *Image {
	return i.sides[side]
}

func (i *CubeImage) SetSide(side CubeSide, image *Image) {
	i.sides[side].CopyFrom(image)
}

func (i CubeImage) TexelUVW(uvw dprec.Vec3) Color {
	side, uv := UVWToCubeUV(uvw)
	sideImage := i.sides[side]
	return sideImage.TexelUV(uv)
}

func (i CubeImage) TexelUVWBilinear(uvw dprec.Vec3) Color {
	side, uv := UVWToCubeUV(uvw)
	sideImage := i.sides[side]
	return sideImage.TexelUVBilinear(uv)
}

func (i *CubeImage) Scale(newSize int) *CubeImage {
	dstImage := NewCubeImage(newSize)
	for side, dstSideImage := range dstImage.sides {
		dstSideImage.CopyFrom(i.sides[side])
	}
	return dstImage
}
