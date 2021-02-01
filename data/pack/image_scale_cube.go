package pack

import (
	"fmt"
	"hash"
	"sync"
)

func ScaleCubeImage(imageProvider CubeImageProvider, dimension int) *ScaleCubeImageAction {
	return &ScaleCubeImageAction{
		imageProvider: imageProvider,
		dimension:     dimension,
	}
}

var _ CubeImageProvider = (*ScaleCubeImageAction)(nil)

type ScaleCubeImageAction struct {
	imageProvider CubeImageProvider
	dimension     int

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *CubeImage
}

func (a *ScaleCubeImageAction) Describe() string {
	return "scale_cube_image"
}

func (a *ScaleCubeImageAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "scale_cube_image", HashableParams{
		"dimension": a.dimension,
		"image":     a.imageProvider,
	})
}

func (a *ScaleCubeImageAction) CubeImage(ctx *Context) (*CubeImage, error) {
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

func (a *ScaleCubeImageAction) run(ctx *Context) (*CubeImage, error) {
	srcImage, err := a.imageProvider.CubeImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cube image: %w", err)
	}
	return srcImage.Scale(a.dimension), nil
}
