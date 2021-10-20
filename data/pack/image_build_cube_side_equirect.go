package pack

import "github.com/mokiat/gomath/dprec"

type BuildCubeSideFromEquirectangularAction struct {
	side          CubeSide
	imageProvider ImageProvider
	image         *Image
}

func (*BuildCubeSideFromEquirectangularAction) Describe() string {
	return "build_cube_side_from_equirectangular()"
}

func (a *BuildCubeSideFromEquirectangularAction) Image() *Image {
	if a.image == nil {
		panic("reading data from unprocessed action")
	}
	return a.image
}

func (a *BuildCubeSideFromEquirectangularAction) Run() error {
	srcImage := a.imageProvider.Image()

	dimension := srcImage.Height / 2
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
			texels[y][x] = srcImage.TexelUV(UVWToEquirectangularUV(CubeUVToUVW(a.side, uv)))
			uv.X += deltaU
		}
		uv.Y += deltaV
	}

	a.image = &Image{
		Width:  dimension,
		Height: dimension,
		Texels: texels,
	}
	return nil
}

func BuildCubeSideFromEquirectangular(srcImage *Image, side CubeSide) *Image {
	dimension := srcImage.Height / 2
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
			texels[y][x] = srcImage.TexelUV(UVWToEquirectangularUV(CubeUVToUVW(side, uv)))
			uv.X += deltaU
		}
		uv.Y += deltaV
	}

	return &Image{
		Width:  dimension,
		Height: dimension,
		Texels: texels,
	}
}
