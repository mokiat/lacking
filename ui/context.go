package ui

import (
	"fmt"
	"image"
	"io"

	"golang.org/x/image/font/opentype"
)

func newContext(window *Window, locator ResourceLocator, graphics Graphics) *Context {
	return &Context{
		window:   window,
		graphics: graphics,
		locator:  locator,

		images:          make(map[string]Image),
		fontCollections: make(map[string]*FontCollection),
	}
}

// Context represents the lifecycle and resource allocation of an Element hierarchy.
type Context struct {
	window   *Window
	graphics Graphics
	locator  ResourceLocator

	images          map[string]Image
	fontCollections map[string]*FontCollection
	fonts           []Font
}

// Window returns the Window that this Context is a part of.
func (c *Context) Window() *Window {
	return c.window
}

// Schedule appends the specified function to be called from
// the main thread (main goroutine).
//
// This function is safe for concurrent use, though
// such use would not guarantee any order for the functions that are
// being concurrently added.
//
// This function can be called from both the main thread, as well as from
// other goroutines.
//
// There is a limit on the number of functions that can be queued within
// a given frame iteration. Once the buffer is full, new functions will be
// dropped.
func (c *Context) Schedule(fn func()) {
	c.window.Schedule(func() error {
		fn()
		return nil
	})
}

// CreateElement creates a new Element instance.
//
// The returned Element is not attached to anything and will not be
// drawn or processed in any way until it is attached to the Element
// hierarchy.
//
// Depending on what you allocate within the context of this Element,
// make sure to Delete this Element once done. Alternatively, as long as the
// Element is part of a hierarchy, you could leave that to the View
// that owns the hierarchy, which will clean everything in its hierarchy,
// once closed.
func (c *Context) CreateElement() *Element {
	return newElement(c)
}

// OpenImage opens the Image at the specified URI location.
//
// The URI is interpreted according to the used ResourceLocator.
//
// As the Image resource consumes resources, its lifecycle becomes linked
// to this Context. Once the owner of this Context is destroyed, the image
// will be released. Keep in mind that just dereferencing the owner is not
// sufficient, as cleanup would not be performed in such cases.
func (c *Context) OpenImage(uri string) (Image, error) {
	if result, ok := c.images[uri]; ok {
		return result, nil
	}

	in, err := c.locator.OpenResource(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open resource: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	result, err := c.graphics.CreateImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to create graphics image: %w", err)
	}
	c.images[uri] = result

	return result, nil
}

// CreateImage creates a new image resource.
//
// As the Image resource consumes resources, its lifecycle becomes linked
// to this Context. Once the owner of this Context is destroyed, the image
// will be released. Keep in mind that just dereferencing the owner is not
// sufficient, as cleanup would not be performed in such cases.
func (c *Context) CreateImage(img image.Image) (Image, error) {
	result, err := c.graphics.CreateImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to create graphics image: %w", err)
	}
	return result, nil
}

// OpenFontCollection opens the FontCollection at the specified URI location.
//
// The URI is interpreted according to the used ResourceLocator.
//
// As the FontCollection consumes resources, its lifecycle becomes linked
// to this Context. Once the owner of this Context is destroyed, the collection
// will be released. Keep in mind that just dereferencing the owner is not
// sufficient, as cleanup would not be performed in such cases.
func (c *Context) OpenFontCollection(uri string) (*FontCollection, error) {
	if result, ok := c.fontCollections[uri]; ok {
		return result, nil
	}

	in, err := c.locator.OpenResource(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open resource: %w", err)
	}
	defer in.Close()

	content, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource content: %w", err)
	}

	collection, err := opentype.ParseCollection(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font collection: %w", err)
	}

	var fonts []Font
	for i := 0; i < collection.NumFonts(); i++ {
		face, err := collection.Font(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get font: %w", err)
		}
		font, err := c.graphics.CreateFont(face)
		if err != nil {
			return nil, fmt.Errorf("failed to create font: %w", err)
		}
		fonts = append(fonts, font)
	}
	c.fonts = append(c.fonts, fonts...)

	result := newFontCollection(fonts)
	c.fontCollections[uri] = result
	return result, nil
}

// GetFont returns the Font with the specified family and sub-family name from
// the loaded family collections.
// If such a font cannot be found, the result will be indicated in the boolean
// flag.
func (c *Context) GetFont(family, subFamily string) (Font, bool) {
	for _, font := range c.fonts {
		if font.Family() == family && font.SubFamily() == subFamily {
			return font, true
		}
	}
	return nil, false
}
