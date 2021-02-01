package pack

import (
	"fmt"
	"hash"
	"sync"

	"github.com/mokiat/gomath/dprec"
)

func BuildCubeSideFromEquirectangular(side CubeSide, imageProvider ImageProvider) *BuildCubeSideFromEquirectangularAction {
	return &BuildCubeSideFromEquirectangularAction{
		side:          side,
		imageProvider: imageProvider,
	}
}

var _ ImageProvider = (*BuildCubeSideFromEquirectangularAction)(nil)

type BuildCubeSideFromEquirectangularAction struct {
	side          CubeSide
	imageProvider ImageProvider

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *Image
}

func (a *BuildCubeSideFromEquirectangularAction) Describe() string {
	return fmt.Sprintf("build_cube_side_from_equirectangular(side: %d)", a.side)
}

func (a *BuildCubeSideFromEquirectangularAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "build_cube_side_from_equirectangular", HashableParams{
		"side":  int(a.side),
		"image": a.imageProvider,
	})
}

func (a *BuildCubeSideFromEquirectangularAction) Image(ctx *Context) (*Image, error) {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	a.resultMutex.Lock()
	defer a.resultMutex.Unlock()

	digest, err := CalculateDigest(a)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}
	if EqualDigests(digest, a.resultDigest) {
		return a.result, nil
	}

	result, err := a.run(ctx)
	if err != nil {
		return nil, err
	}

	a.result = result
	a.resultDigest = digest
	return result, nil
}

func (a *BuildCubeSideFromEquirectangularAction) run(ctx *Context) (*Image, error) {
	srcImage, err := a.imageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get equirectangular image: %w", err)
	}

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

	return &Image{
		Width:  dimension,
		Height: dimension,
		Texels: texels,
	}, nil
}
