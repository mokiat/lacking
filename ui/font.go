package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

// Use the following links as reference for Font terminology:
// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
// https://www.freetype.org/freetype2/docs/glyphs/glyphs-3.html

func newFontCollection(fonts []*Font) *FontCollection {
	return &FontCollection{
		fonts: fonts,
	}
}

// FontCollection represents a collection of Fonts.
type FontCollection struct {
	fonts []*Font
}

// Fonts returns all Fonts contained by this collection.
func (c *FontCollection) Fonts() []*Font {
	return c.fonts
}

// Destroy releases all resources held by this FontCollection.
func (c *FontCollection) Destroy() {
	for _, font := range c.fonts {
		font.Destroy()
	}
	c.fonts = nil
}

// Font represents a text Font.
type Font struct {
	// familyName holds the primary name of the font (e.g. Roboto, Monospaced)
	familyName string
	// subFamilyName holds the variant of the font (e.g. bold, italic)
	subFamilyName string

	// lineHeight holds the distance (positive) from one line baseline to the next.
	lineHeight float32
	// lineAscent holds the distance (positive) from the baseline to the top of the line.
	lineAscent float32
	// lineDescent holds the distance (positive) from the baseline to the bottom of the line.
	lineDescent float32

	// glyphs holds a map of supported glyphs.
	glyphs map[rune]*fontGlyph

	// texture holds the OpenGL texture to be used for font rendering
	texture render.Texture
}

// Family returns the family name of this Font.
// (e.g. Roboto, Open Sans)
func (f *Font) Family() string {
	return f.familyName
}

// SubFamily returns the sub-family name of this Font.
// (e.g. Italic, Bold)
func (f *Font) SubFamily() string {
	return f.subFamilyName
}

// LineHeight returns the height of a single line based on the font size.
func (f *Font) LineHeight(fontSize float32) float32 {
	return (f.lineAscent + f.lineDescent) * fontSize
}

// LineWidth returns the width of a single text line.
func (f *Font) LineWidth(characters []rune, fontSize float32) float32 {
	var result float32

	lastGlyph := (*fontGlyph)(nil)
	for _, ch := range characters {
		if glyph, ok := f.glyphs[ch]; ok {
			result += glyph.advance
			if lastGlyph != nil {
				result += lastGlyph.kerns[ch]
			}
			lastGlyph = glyph
		}
	}
	return result * fontSize
}

// LineIterator returns a new LineIterator over the specified text for the
// specified font size.
func (f *Font) LineIterator(characters []rune, fontSize float32) *LineIterator {
	return &LineIterator{
		font:     f,
		text:     characters,
		offset:   0,
		fontSize: fontSize,
	}
}

// TextSize returns the size it would take to draw the
// specified text string at the specified font size.
func (f *Font) TextSize(text string, fontSize float32) sprec.Vec2 {
	result := sprec.NewVec2(0, f.lineHeight)
	if len(text) == 0 {
		return sprec.Vec2Prod(result, fontSize)
	}

	currentWidth := float32(0.0)
	lastGlyph := (*fontGlyph)(nil)
	for _, ch := range text {
		if ch == '\r' {
			lastGlyph = nil
			continue
		}
		if ch == '\n' {
			result.X = sprec.Max(result.X, currentWidth)
			result.Y += f.lineHeight
			currentWidth = 0.0
			lastGlyph = nil
			continue
		}
		if glyph, ok := f.glyphs[ch]; ok {
			currentWidth += glyph.advance
			if lastGlyph != nil {
				currentWidth += lastGlyph.kerns[ch]
			}
			lastGlyph = glyph
		}
	}
	result.X = sprec.Max(result.X, currentWidth)
	return sprec.Vec2Prod(result, fontSize)
}

// Destroy releases all resources related to this Font.
func (f *Font) Destroy() {
	f.texture.Release()
}

type fontGlyph struct {
	// leftU holds the U texture coordinate of the left edge of the
	// glyph bounds.
	leftU float32
	// rightU holds the U texture coordinate of the right edge of the
	// glyph bounds.
	rightU float32
	// topV holds the V texture coordinate of the top edge of the
	// glyph bounds.
	topV float32
	// bottomV holds the V texture coordinate of the bottom edge of the
	// glyph bounds.
	bottomV float32

	// advance holds the distance to move from this glyph onto the next one and
	// includes the bearing values.
	advance float32
	// ascent holds the distance (positive) from the baseline to the glyph's top
	ascent float32
	// descent holds the distance (positive) from the baseline to the glyph's
	// bottom
	descent float32
	// leftBearing determines the room to leave to the left of the glyph
	leftBearing float32
	// rightBearing determines the room to leave to the right of the glyph
	rightBearing float32

	// kerns holds positional adjustments between two glyphs. A positive
	// value means to move them further apart.
	kerns map[rune]float32
}

// LineIterator represents an optimal way of evaluating the size of a text
// once character at a time.
type LineIterator struct {
	font     *Font
	offset   int
	text     []rune
	fontSize float32
	result   Character
}

// Next evaluates a character from the text and returns whether
// there was any character.
func (i *LineIterator) Next() bool {
	if i.offset >= len(i.text) {
		return false
	}

	char := i.text[i.offset]
	i.result = Character{
		Rune:  char,
		Kern:  0.0,
		Width: 0.0,
	}

	glyph, ok := i.font.glyphs[char]
	if ok {
		i.result.Width = glyph.advance * i.fontSize

		if i.offset > 0 {
			prevChar := i.text[i.offset-1]
			if prevGlyph, ok := i.font.glyphs[prevChar]; ok {
				i.result.Kern = prevGlyph.kerns[char]
			}
		}
	}

	i.offset++
	return true
}

// Character returns the last iterated character.
func (i *LineIterator) Character() Character {
	return i.result
}

// Character contains information on a character from a text iteration.
type Character struct {

	// Rune contains the representation of the character.
	Rune rune

	// Kern indicates the offset from the previous character.
	Kern float32

	// Width holds the horizontal size of the character.
	Width float32
}
