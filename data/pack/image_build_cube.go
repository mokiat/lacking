package pack

import (
	"fmt"
	"hash"
	"sync"
)

func BuildCubeImage() *BuildCubeImageAction {
	return &BuildCubeImageAction{}
}

var _ CubeImageProvider = (*BuildCubeImageAction)(nil)

type BuildCubeImageAction struct {
	frontImageProvider  ImageProvider
	rearImageProvider   ImageProvider
	leftImageProvider   ImageProvider
	rightImageProvider  ImageProvider
	topImageProvider    ImageProvider
	bottomImageProvider ImageProvider
	dimension           int

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *CubeImage
}

func (a *BuildCubeImageAction) WithDimension(dimension int) *BuildCubeImageAction {
	a.dimension = dimension
	return a
}

func (a *BuildCubeImageAction) WithFrontImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.frontImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) WithRearImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.rearImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) WithLeftImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.leftImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) WithRightImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.rightImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) WithTopImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.topImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) WithBottomImage(imageProvider ImageProvider) *BuildCubeImageAction {
	a.bottomImageProvider = imageProvider
	return a
}

func (a *BuildCubeImageAction) Describe() string {
	return "build_cube_image"
}

func (a *BuildCubeImageAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "build_cube_image", HashableParams{
		"front_image":  a.frontImageProvider,
		"rear_image":   a.rearImageProvider,
		"left_image":   a.leftImageProvider,
		"right_image":  a.rightImageProvider,
		"top_image":    a.topImageProvider,
		"bottom_image": a.bottomImageProvider,
	})
}

func (a *BuildCubeImageAction) CubeImage(ctx *Context) (*CubeImage, error) {
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

func (a *BuildCubeImageAction) run(ctx *Context) (*CubeImage, error) {
	frontImage, err := a.frontImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get front image: %w", err)
	}
	if !frontImage.IsSquare() {
		return nil, fmt.Errorf("front image is not a square (%d, %d)", frontImage.Width, frontImage.Height)
	}
	rearImage, err := a.rearImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rear image: %w", err)
	}
	if !rearImage.IsSquare() {
		return nil, fmt.Errorf("rear image is not a square (%d, %d)", rearImage.Width, rearImage.Height)
	}
	leftImage, err := a.leftImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get left image: %w", err)
	}
	if !leftImage.IsSquare() {
		return nil, fmt.Errorf("left image is not a square (%d, %d)", leftImage.Width, leftImage.Height)
	}
	rightImage, err := a.rightImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get right image: %w", err)
	}
	if !rightImage.IsSquare() {
		return nil, fmt.Errorf("right image is not a square (%d, %d)", rightImage.Width, rightImage.Height)
	}
	topImage, err := a.topImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top image: %w", err)
	}
	if !topImage.IsSquare() {
		return nil, fmt.Errorf("top image is not a square (%d, %d)", topImage.Width, topImage.Height)
	}
	bottomImage, err := a.bottomImageProvider.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bottom image: %w", err)
	}
	if !bottomImage.IsSquare() {
		return nil, fmt.Errorf("bottom image is not a square (%d, %d)", bottomImage.Width, bottomImage.Height)
	}

	if a.dimension > 0 {
		frontImage = frontImage.Scale(a.dimension, a.dimension)
		rearImage = rearImage.Scale(a.dimension, a.dimension)
		leftImage = leftImage.Scale(a.dimension, a.dimension)
		rightImage = rightImage.Scale(a.dimension, a.dimension)
		topImage = topImage.Scale(a.dimension, a.dimension)
		bottomImage = bottomImage.Scale(a.dimension, a.dimension)
	} else {
		areSameDimension := frontImage.Width == rearImage.Width &&
			frontImage.Width == leftImage.Width &&
			frontImage.Width == rightImage.Width &&
			frontImage.Width == topImage.Width &&
			frontImage.Width == bottomImage.Width
		if !areSameDimension {
			return nil, fmt.Errorf("images are not of the same size")
		}
	}

	image := &CubeImage{
		Dimension: frontImage.Width,
	}
	image.Sides[CubeSideFront] = CubeImageSide{
		Texels: frontImage.Texels,
	}
	image.Sides[CubeSideRear] = CubeImageSide{
		Texels: rearImage.Texels,
	}
	image.Sides[CubeSideLeft] = CubeImageSide{
		Texels: leftImage.Texels,
	}
	image.Sides[CubeSideRight] = CubeImageSide{
		Texels: rightImage.Texels,
	}
	image.Sides[CubeSideTop] = CubeImageSide{
		Texels: topImage.Texels,
	}
	image.Sides[CubeSideBottom] = CubeImageSide{
		Texels: bottomImage.Texels,
	}
	return image, nil
}
