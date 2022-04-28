package component

import (
	"fmt"
	"image"

	"github.com/mokiat/lacking/ui"
)

var uiCtx *ui.Context

// Window returns the underlying ui Window object.
func Window() *ui.Window {
	return uiCtx.Window()
}

// OpenImage delegates to the UI window context to open
// the specified image.
func OpenImage(uri string) *ui.Image {
	img, err := uiCtx.OpenImage(uri)
	if err != nil {
		panic(fmt.Errorf("failed to open image %q: %w", uri, err))
	}
	return img
}

// CreateImage delegates to the UI window context to create
// the specified image.
func CreateImage(img image.Image) *ui.Image {
	result, err := uiCtx.CreateImage(img)
	if err != nil {
		panic(fmt.Errorf("failed to create image: %w", err))
	}
	return result
}

// OpenFontCollection delegates to the UI window context to open
// the specified font collection.
func OpenFontCollection(uri string) {
	if _, err := uiCtx.OpenFontCollection(uri); err != nil {
		panic(fmt.Errorf("failed to open font collection %q: %w", uri, err))
	}
}

// GetFont retrieves the font with the specified family and style.
//
// Keep in mind that the necessary fonts should have been loaded via
// OpenFontCollection beforehand, otherwise this method will panic if
// it is unable to find the requested font.
func GetFont(family, style string) *ui.Font {
	font, found := uiCtx.GetFont(family, style)
	if !found {
		panic(fmt.Errorf("could not find font %q / %q", family, style))
	}
	return font
}
