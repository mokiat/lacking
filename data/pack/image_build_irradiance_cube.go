package pack

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
)

type BuildIrradianceCubeImageAction struct {
	imageProvider CubeImageProvider
	sampleCount   int
	image         *CubeImage
}

type BuildIrradianceCubeImageOption func(a *BuildIrradianceCubeImageAction)

func WithSampleCount(count int) BuildIrradianceCubeImageOption {
	return func(a *BuildIrradianceCubeImageAction) {
		a.sampleCount = count
	}
}

func (a *BuildIrradianceCubeImageAction) Describe() string {
	return fmt.Sprintf("build_irradiance_cube_image(samples: %d)", a.sampleCount)
}

func (a *BuildIrradianceCubeImageAction) CubeImage() *CubeImage {
	if a.image == nil {
		panic("reading data from unprocessed action")
	}
	return a.image
}

func (a *BuildIrradianceCubeImageAction) Run() error {
	srcImage := a.imageProvider.CubeImage()
	dimension := srcImage.Dimension

	a.image = &CubeImage{
		Dimension: dimension,
	}
	for i := range srcImage.Sides {
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
				deltaLat := (endLat - startLat) / dprec.Radians(float64(a.sampleCount))

				color := Color{}
				positiveSamples := 0.0
				for lat := startLat; lat < endLat; lat += deltaLat {
					startLong := dprec.Radians(-dprec.Pi)
					endLong := dprec.Radians(dprec.Pi)
					deltaLong := (endLong - startLong) / (dprec.Radians(float64(a.sampleCount) * (dprec.Cos(lat) + 0.01)))

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

		a.image.Sides[i] = CubeImageSide{
			Texels: texels,
		}
	}
	return nil
}
