package mdl

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mdouchement/hdr/codec/rgbe"
	_ "golang.org/x/image/tiff"

	"github.com/mdouchement/hdr"
	"github.com/mokiat/goexr/exr"
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

func ParseImage(in io.Reader) (*Image, error) {
	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return BuildImageResource(img), nil
}

func BuildImageResource(img image.Image) *Image {
	imgStartX := img.Bounds().Min.X
	imgStartY := img.Bounds().Min.Y
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	image := NewImage(width, height)
	for y := range height {
		for x := range width {
			atX := imgStartX + x
			atY := imgStartY + y
			switch img := img.(type) {
			case hdr.Image:
				r, g, b, a := img.HDRAt(atX, atY).HDRPixel()
				image.SetTexel(x, y, Color{
					R: r,
					G: g,
					B: b,
					A: a,
				})
			case *exr.RGBAImage:
				clr := img.At(atX, atY).(exr.RGBAColor)
				image.SetTexel(x, y, Color{
					R: float64(clr.R),
					G: float64(clr.G),
					B: float64(clr.B),
					A: float64(clr.A),
				})
			default:
				c := color.NRGBAModel.Convert(img.At(atX, atY)).(color.NRGBA)
				image.SetTexel(x, y, Color{
					R: float64(float64(c.R) / 255.0),
					G: float64(float64(c.G) / 255.0),
					B: float64(float64(c.B) / 255.0),
					A: float64(float64(c.A) / 255.0),
				})
			}
		}
	}
	return image
}

func BuildCubeSideFromEquirectangular(side CubeSide, srcImage *Image) *Image {
	dimension := max(1, srcImage.Height()/2)
	dstImage := NewImage(dimension, dimension)

	uv := dprec.ZeroVec2()
	startUV := dprec.NewVec2(0.0, 1.0)
	deltaUV := dprec.NewVec2(1.0/float64(dimension-1), -1.0/float64(dimension-1))

	uv.Y = startUV.Y
	for y := range dimension {
		uv.X = startUV.X
		for x := range dimension {
			texel := srcImage.TexelUVBilinear(UVWToEquirectangularUV(CubeUVToUVW(side, uv)))
			dstImage.SetTexel(x, y, texel)
			uv.X += deltaUV.X
		}
		uv.Y += deltaUV.Y
	}
	return dstImage
}

func BuildIrradianceCubeImage(srcImage *CubeImage, sampleCount int) *CubeImage {
	dstImage := NewCubeImage(srcImage.size)
	var group sync.WaitGroup
	for i := range srcImage.sides {
		group.Add(1)
		go func() {
			defer group.Done()
			projectIrradianceCubeImageSide(srcImage, dstImage, CubeSide(i), sampleCount)
		}()
	}
	group.Wait()
	return dstImage
}

func projectIrradianceCubeImageSide(srcImage, dstImage *CubeImage, side CubeSide, sampleCount int) {
	dimension := srcImage.size

	startLat := dprec.Degrees(-90.0)
	endLat := dprec.Degrees(90.0)
	deltaLat := (endLat - startLat) / dprec.Radians(float64(sampleCount))

	startLong := dprec.Degrees(-180.0)
	endLong := dprec.Degrees(180.0)

	dstSide := dstImage.Side(side)
	uv := dprec.ZeroVec2()
	startU := 0.0
	deltaU := 1.0 / float64(dimension-1)
	startV := 1.0
	deltaV := -1.0 / float64(dimension-1)

	uv.Y = startV
	for y := range dimension {
		uv.X = startU

		for x := range dimension {
			uvw := CubeUVToUVW(side, uv)

			var color Color
			var positiveSamples int
			for lat := startLat; lat < endLat; lat += deltaLat {
				latitudeCS := dprec.Cos(lat)
				latitudeSN := dprec.Sin(lat)

				deltaLong := (endLong - startLong) / (dprec.Radians(float64(sampleCount) * (latitudeCS + 0.01)))
				for long := startLong; long < endLong; long += deltaLong {
					longitudeCS := dprec.Cos(long)
					longitudeSN := dprec.Sin(long)

					direction := dprec.NewVec3(
						longitudeSN*latitudeCS,
						longitudeCS*latitudeCS,
						latitudeSN,
					)
					if dot := dprec.Vec3Dot(uvw, direction); dot > 0.0 {
						positiveSamples++
						srcColor := srcImage.TexelUVW(direction)
						color.R += srcColor.R * dot
						color.G += srcColor.G * dot
						color.B += srcColor.B * dot
					}
				}
			}

			if positiveSamples > 0 {
				// NOTE: Unlike some tutorials, we scale by 2*Pi instead of just Pi.
				// This is because we have uniform sampling across the semi-sphere
				// and we then need to scale according to the surface of a unit
				// semisphere, which is 2 * Pi, not Pi.
				// If you consider this integration as a Monte Carlo integration,
				// the volume term is 2 * Pi, not Pi.
				weight := 2.0 * dprec.Pi / float64(positiveSamples)
				color.R *= weight
				color.G *= weight
				color.B *= weight
			}
			color.A = 1.0
			dstSide.SetTexel(x, y, color)

			uv.X += deltaU
		}
		uv.Y += deltaV
	}
}
