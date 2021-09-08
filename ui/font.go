package ui

func newFontCollection(fonts []Font) *FontCollection {
	return &FontCollection{
		fonts: fonts,
	}
}

// FontCollection represents a collection of Fonts.
type FontCollection struct {
	fonts []Font
}

// Fonts returns all Fonts contained by this collection.
func (c *FontCollection) Fonts() []Font {
	return c.fonts
}

// Font represents a text Font.
type Font interface {

	// Family returns the family name of this Font.
	// (e.g. Roboto, Open Sans)
	Family() string

	// SubFamily returns the sub-family name of this Font.
	// (e.g. Italic, Bold)
	SubFamily() string

	// TextSize returns the size it would take to draw the
	// specified text string at the specified font size.
	TextSize(text string, fontSize int) Size
}
