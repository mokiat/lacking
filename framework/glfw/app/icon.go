package app

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
)

func openImage(path string) (image.Image, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}
