package ui

import (
	"image"

	"golang.org/x/image/font/opentype"
)

func newContext(parent *Context, window *Window, resMan *resourceManager) *Context {
	logger.Debug("Creating context.")
	return &Context{
		parent: parent,
		window: window,
		resMan: resMan,

		adhocImages: nil,
		namedImages: make(map[string]*Image),

		adhocFonts: nil,
		namedFonts: make(map[string]*Font),

		adhocFontCollections: nil,
		namedFontCollections: make(map[string]*FontCollection),
	}
}

// Context represents the lifecycle and resource allocation of an Element hierarchy.
type Context struct {
	parent *Context
	window *Window
	resMan *resourceManager

	adhocImages []*Image
	namedImages map[string]*Image

	adhocFonts []*Font
	namedFonts map[string]*Font

	adhocFontCollections []*FontCollection
	namedFontCollections map[string]*FontCollection
}

// Window returns the Window that this Context is a part of.
func (c *Context) Window() *Window {
	return c.window
}

// Schedule appends the specified function to be called from
// the UI thread (UI goroutine).
//
// This function is safe for concurrent use, though such use would not
// guarantee any order for the functions that are being concurrently added.
//
// This function can be called from both the UI thread, as well as from
// other goroutines.
//
// There is a limit to the number of functions that can be queued within
// a given frame iteration. Once the buffer is full, new functions will be
// dropped.
func (c *Context) Schedule(fn func()) {
	c.window.Schedule(fn)
}

// CreateContext returns a new Context that is a child of the current
// Context and as such can reuse resources held by the current context.
func (c *Context) CreateContext() *Context {
	return newContext(c, c.window, c.resMan)
}

// CreateImage creates a new Image resource.
//
// The Image will be destroyed once this Context is destroyed.
func (c *Context) CreateImage(img image.Image) (*Image, error) {
	logger.Debug("Creating ad-hoc image.")
	result := c.resMan.CreateImage(img)
	c.adhocImages = append(c.adhocImages, result)
	return result, nil
}

// OpenImage opens the Image at the specified URI location.
//
// The URI is interpreted according to the used ResourceLocator.
//
// The Image will be destroyed once this Context is destroyed.
func (c *Context) OpenImage(uri string) (*Image, error) {
	if result, ok := c.findImage(uri); ok {
		return result, nil
	}
	logger.Debug("Opening named image (%q).", uri)
	result, err := c.resMan.OpenImage(uri)
	if err != nil {
		return nil, err
	}
	c.namedImages[uri] = result
	return result, nil
}

// CreateFont creates a new Font resource.
//
// The Font will be destroyed once this Context is destroyed.
func (c *Context) CreateFont(font *opentype.Font) (*Font, error) {
	logger.Debug("Creating ad-hoc font.")
	result, err := c.resMan.CreateFont(font)
	if err != nil {
		return nil, err
	}
	c.adhocFonts = append(c.adhocFonts, result)
	return result, nil
}

// OpenFont opens the Font at the specified URI location.
//
// The URI is interpreted according to the used ResourceLocator.
//
// The Font will be destroyed once this Context is destroyed.
func (c *Context) OpenFont(uri string) (*Font, error) {
	if result, ok := c.findFont(uri); ok {
		return result, nil
	}
	logger.Debug("Opening named font (%q).", uri)
	result, err := c.resMan.OpenFont(uri)
	if err != nil {
		return nil, err
	}
	c.namedFonts[uri] = result
	return result, nil
}

// CreateFontCollection creates a new FontCollection resource.
//
// The FontCollection will be destroyed once this Context is destroyed.
func (c *Context) CreateFontCollection(collection *opentype.Collection) (*FontCollection, error) {
	logger.Debug("Creating ad-hoc font collection.")
	result, err := c.resMan.CreateFontCollection(collection)
	if err != nil {
		return nil, err
	}
	c.adhocFontCollections = append(c.adhocFontCollections, result)
	return result, nil
}

// OpenFontCollection opens the FontCollection at the specified URI location.
//
// The URI is interpreted according to the used ResourceLocator.
//
// The FontCollection will be destroyed once this Context is destroyed.
func (c *Context) OpenFontCollection(uri string) (*FontCollection, error) {
	if result, ok := c.findFontCollection(uri); ok {
		return result, nil
	}
	logger.Debug("Opening named font collection (%q).", uri)
	result, err := c.resMan.OpenFontCollection(uri)
	if err != nil {
		return nil, err
	}
	c.namedFontCollections[uri] = result
	return result, nil
}

// GetFont returns the Font with the specified family and sub-family name from
// all of the created or loaded fonts and/or family collections.
func (c *Context) GetFont(family, subFamily string) (*Font, bool) {
	matches := func(font *Font) bool {
		return font.Family() == family && font.SubFamily() == subFamily
	}
	for _, font := range c.adhocFonts {
		if matches(font) {
			return font, true
		}
	}
	for _, font := range c.namedFonts {
		if matches(font) {
			return font, true
		}
	}
	for _, collection := range c.adhocFontCollections {
		for _, font := range collection.fonts {
			if matches(font) {
				return font, true
			}
		}
	}
	for _, collection := range c.namedFontCollections {
		for _, font := range collection.fonts {
			if matches(font) {
				return font, true
			}
		}
	}
	return nil, false
}

// Destroy releases all resources held by this Context.
func (c *Context) Destroy() {
	logger.Debug("Destroying context.")

	for _, image := range c.adhocImages {
		logger.Debug("Destroying ad-hoc image.")
		image.Destroy()
	}
	c.adhocImages = nil

	for uri, image := range c.namedImages {
		logger.Debug("Destroying named image (%q).", uri)
		image.Destroy()
	}
	c.namedImages = make(map[string]*Image)

	for _, font := range c.adhocFonts {
		logger.Debug("Destroying ad-hoc font.")
		font.Destroy()
	}
	c.adhocFonts = nil

	for uri, font := range c.namedFonts {
		logger.Debug("Destroying named font (%q).", uri)
		font.Destroy()
	}
	c.namedFonts = make(map[string]*Font)

	for _, collection := range c.adhocFontCollections {
		logger.Debug("Destroying ad-hoc font collection.")
		collection.Destroy()
	}
	c.adhocFontCollections = nil

	for uri, collection := range c.namedFontCollections {
		logger.Debug("Destroying named font collection (%q).", uri)
		collection.Destroy()
	}
}

func (c *Context) findImage(uri string) (*Image, bool) {
	if result, ok := c.namedImages[uri]; ok {
		return result, true
	}
	if c.parent != nil {
		return c.parent.findImage(uri)
	}
	return nil, false
}

func (c *Context) findFont(uri string) (*Font, bool) {
	if result, ok := c.namedFonts[uri]; ok {
		return result, true
	}
	if c.parent != nil {
		return c.parent.findFont(uri)
	}
	return nil, false
}

func (c *Context) findFontCollection(uri string) (*FontCollection, bool) {
	if result, ok := c.namedFontCollections[uri]; ok {
		return result, true
	}
	if c.parent != nil {
		return c.parent.findFontCollection(uri)
	}
	return nil, false
}
