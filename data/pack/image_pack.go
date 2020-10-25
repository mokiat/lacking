package pack

import (
	"fmt"
	"math"
)

type ImageProvider interface {
	Image() *Image
}

type Image struct {
	Width  int
	Height int
	Texels [][]Color
}

func (i *Image) IsSquare() bool {
	return i.Width == i.Height
}

func (i *Image) Texel(x, y int) Color {
	return i.Texels[y][x]
}

func (i *Image) BilinearTexel(x, y float64) Color {
	floorX := math.Floor(x)
	ceilX := math.Ceil(x)
	floorY := math.Floor(y)
	ceilY := math.Ceil(y)
	fractX := x - floorX
	fractY := y - floorY

	topLeft := i.Texels[int(floorY)][int(floorX)]
	topRight := i.Texels[int(floorY)][int(ceilX)]
	bottomLeft := i.Texels[int(ceilY)][int(floorX)]
	bottomRight := i.Texels[int(ceilY)][int(ceilX)]

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
	if newWidth < i.Width/2 {
		image := i.Scale(i.Width/2, i.Height)
		return image.Scale(newWidth, newHeight)
	}
	if newHeight < i.Height/2 {
		image := i.Scale(i.Width, i.Height/2)
		return image.Scale(newWidth, newHeight)
	}

	newTexels := make([][]Color, newHeight)
	for y := 0; y < newHeight; y++ {
		newTexels[y] = make([]Color, newWidth)
		oldY := float64(y) * (float64(i.Height-1) / float64(newHeight-1))
		for x := 0; x < newWidth; x++ {
			oldX := float64(x) * (float64(i.Width-1) / float64(newWidth-1))
			newTexels[y][x] = i.BilinearTexel(oldX, oldY)
		}
	}
	return &Image{
		Width:  newWidth,
		Height: newHeight,
		Texels: newTexels,
	}
}

func (i *Image) RGBA8Data() []byte {
	data := make([]byte, 4*i.Width*i.Height)
	offset := 0
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			texel := i.Texel(x, i.Height-y-1)
			data[offset+0] = byte(255.0 * texel.R)
			data[offset+1] = byte(255.0 * texel.G)
			data[offset+2] = byte(255.0 * texel.B)
			data[offset+3] = byte(255.0 * texel.A)
			offset += 4
		}
	}
	return data
}

type CubeSide int

const (
	CubeSideFront CubeSide = iota
	CubeSideRear
	CubeSideLeft
	CubeSideRight
	CubeSideTop
	CubeSideBottom
)

type CubeImage struct {
	Dimension int
	Sides     [6]CubeImageSide
}

func (t *CubeImage) RGBA8Data(side CubeSide) []byte {
	data := make([]byte, 4*t.Dimension*t.Dimension)
	offset := 0
	texSide := t.Sides[side]
	for y := 0; y < t.Dimension; y++ {
		for x := 0; x < t.Dimension; x++ {
			texel := texSide.Texel(x, t.Dimension-y-1)
			data[offset+0] = byte(255.0 * texel.R)
			data[offset+1] = byte(255.0 * texel.G)
			data[offset+2] = byte(255.0 * texel.B)
			data[offset+3] = byte(255.0 * texel.A)
			offset += 4
		}
	}
	return data
}

type CubeImageSide struct {
	Texels [][]Color
}

func (s CubeImageSide) Texel(x, y int) Color {
	return s.Texels[y][x]
}

type CubeImageProvider interface {
	CubeImage() *CubeImage
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

type Color struct {
	R float64
	G float64
	B float64
	A float64
}
