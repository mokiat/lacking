package pack

import "fmt"

type ScaleCubeImageAction struct {
	imageProvider CubeImageProvider
	dimension     int
	image         *CubeImage
}

func (a *ScaleCubeImageAction) Describe() string {
	return fmt.Sprintf("scale_cube_image(size: %d)", a.dimension)
}

func (a *ScaleCubeImageAction) CubeImage() *CubeImage {
	if a.image == nil {
		panic("reading data from unprocessed action")
	}
	return a.image
}

func (a *ScaleCubeImageAction) Run() error {
	srcImage := a.imageProvider.CubeImage()
	a.image = srcImage.Scale(a.dimension)
	return nil
}
