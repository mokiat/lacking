package mdl

import (
	"fmt"
	"math"

	"github.com/mokiat/gomath/dprec"
)

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
			dstImage.SetTexel(x, y, srcImage.TexelUVBilinear(UVWToEquirectangularUV(CubeUVToUVW(side, uv))))
			uv.X += deltaU
		}
		uv.Y += deltaV
	}
	return dstImage
}

func BuildIrradianceCubeImage(srcImage *CubeImage, sampleCount int) *CubeImage {
	dimension := srcImage.size

	dstImage := &CubeImage{
		size: dimension,
	}
	for i := range srcImage.sides {
		dstSide := NewImage(dimension, dimension)

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

				dstSide.SetTexel(x, y, Color{
					R: dprec.Pi * color.R / positiveSamples,
					G: dprec.Pi * color.G / positiveSamples,
					B: dprec.Pi * color.B / positiveSamples,
					A: 1.0,
				})

				uv.X += deltaU
			}
			uv.Y += deltaV
		}

		dstImage.sides[i] = dstSide
	}
	return dstImage
}
