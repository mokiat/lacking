package pack

import "fmt"

type BuildCubeImageAction struct {
	frontImage  ImageProvider
	rearImage   ImageProvider
	leftImage   ImageProvider
	rightImage  ImageProvider
	topImage    ImageProvider
	bottomImage ImageProvider
	dimension   int
	image       *CubeImage
}

type BuildCubeImageOption func(a *BuildCubeImageAction)

func WithFrontImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.frontImage = image
	}
}

func WithRearImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.rearImage = image
	}
}

func WithLeftImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.leftImage = image
	}
}

func WithRightImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.rightImage = image
	}
}

func WithTopImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.topImage = image
	}
}

func WithBottomImage(image ImageProvider) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.bottomImage = image
	}
}

func WithDimension(dimension int) BuildCubeImageOption {
	return func(a *BuildCubeImageAction) {
		a.dimension = dimension
	}
}

func (*BuildCubeImageAction) Describe() string {
	return "build_cube_image()"
}

func (a *BuildCubeImageAction) CubeImage() *CubeImage {
	if a.image == nil {
		panic("reading data from unprocessed action")
	}
	return a.image
}

func (a *BuildCubeImageAction) Run() error {
	frontImage := a.frontImage.Image()
	if !frontImage.IsSquare() {
		return fmt.Errorf("front image is not a square (%d, %d)", frontImage.Width, frontImage.Height)
	}
	rearImage := a.rearImage.Image()
	if !rearImage.IsSquare() {
		return fmt.Errorf("rear image is not a square (%d, %d)", rearImage.Width, rearImage.Height)
	}
	leftImage := a.leftImage.Image()
	if !leftImage.IsSquare() {
		return fmt.Errorf("left image is not a square (%d, %d)", leftImage.Width, leftImage.Height)
	}
	rightImage := a.rightImage.Image()
	if !rightImage.IsSquare() {
		return fmt.Errorf("right image is not a square (%d, %d)", rightImage.Width, rightImage.Height)
	}
	topImage := a.topImage.Image()
	if !topImage.IsSquare() {
		return fmt.Errorf("top image is not a square (%d, %d)", topImage.Width, topImage.Height)
	}
	bottomImage := a.bottomImage.Image()
	if !bottomImage.IsSquare() {
		return fmt.Errorf("bottom image is not a square (%d, %d)", bottomImage.Width, bottomImage.Height)
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
			return fmt.Errorf("images are not of the same size")
		}
	}

	a.image = &CubeImage{
		Dimension: frontImage.Width,
	}
	a.image.Sides[CubeSideFront] = CubeImageSide{
		Texels: frontImage.Texels,
	}
	a.image.Sides[CubeSideRear] = CubeImageSide{
		Texels: rearImage.Texels,
	}
	a.image.Sides[CubeSideLeft] = CubeImageSide{
		Texels: leftImage.Texels,
	}
	a.image.Sides[CubeSideRight] = CubeImageSide{
		Texels: rightImage.Texels,
	}
	a.image.Sides[CubeSideTop] = CubeImageSide{
		Texels: topImage.Texels,
	}
	a.image.Sides[CubeSideBottom] = CubeImageSide{
		Texels: bottomImage.Texels,
	}
	return nil
}

func BuildCube(frontImg, rearImg, leftImg, rightImg, topImg, bottomImg *Image, dimension int) (*CubeImage, error) {
	if !frontImg.IsSquare() {
		return nil, fmt.Errorf("front image is not a square (%d, %d)", frontImg.Width, frontImg.Height)
	}
	if !rearImg.IsSquare() {
		return nil, fmt.Errorf("rear image is not a square (%d, %d)", rearImg.Width, rearImg.Height)
	}
	if !leftImg.IsSquare() {
		return nil, fmt.Errorf("left image is not a square (%d, %d)", leftImg.Width, leftImg.Height)
	}
	if !rightImg.IsSquare() {
		return nil, fmt.Errorf("right image is not a square (%d, %d)", rightImg.Width, rightImg.Height)
	}
	if !topImg.IsSquare() {
		return nil, fmt.Errorf("top image is not a square (%d, %d)", topImg.Width, topImg.Height)
	}
	if !bottomImg.IsSquare() {
		return nil, fmt.Errorf("bottom image is not a square (%d, %d)", bottomImg.Width, bottomImg.Height)
	}

	if dimension > 0 {
		frontImg = frontImg.Scale(dimension, dimension)
		rearImg = rearImg.Scale(dimension, dimension)
		leftImg = leftImg.Scale(dimension, dimension)
		rightImg = rightImg.Scale(dimension, dimension)
		topImg = topImg.Scale(dimension, dimension)
		bottomImg = bottomImg.Scale(dimension, dimension)
	} else {
		areSameDimension := frontImg.Width == rearImg.Width &&
			frontImg.Width == leftImg.Width &&
			frontImg.Width == rightImg.Width &&
			frontImg.Width == topImg.Width &&
			frontImg.Width == bottomImg.Width
		if !areSameDimension {
			return nil, fmt.Errorf("images are not of the same size")
		}
	}

	result := &CubeImage{
		Dimension: frontImg.Width,
	}
	result.Sides[CubeSideFront] = CubeImageSide{
		Texels: frontImg.Texels,
	}
	result.Sides[CubeSideRear] = CubeImageSide{
		Texels: rearImg.Texels,
	}
	result.Sides[CubeSideLeft] = CubeImageSide{
		Texels: leftImg.Texels,
	}
	result.Sides[CubeSideRight] = CubeImageSide{
		Texels: rightImg.Texels,
	}
	result.Sides[CubeSideTop] = CubeImageSide{
		Texels: topImg.Texels,
	}
	result.Sides[CubeSideBottom] = CubeImageSide{
		Texels: bottomImg.Texels,
	}
	return result, nil
}
