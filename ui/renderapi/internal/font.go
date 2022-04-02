package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
)

// Use the following links as reference for Font terminology:
// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
// https://www.freetype.org/freetype2/docs/glyphs/glyphs-3.html

var _ ui.Font = (*Font)(nil)

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

func (f *Font) Family() string {
	return f.familyName
}

func (f *Font) SubFamily() string {
	return f.subFamilyName
}

func (f *Font) TextSize(text string, fontSize int) ui.Size {
	if len(text) == 0 {
		return ui.NewSize(0, 0)
	}

	result := sprec.NewVec2(0, f.lineAscent+f.lineDescent)
	currentWidth := float32(0.0)
	lastGlyph := (*fontGlyph)(nil)
	for _, ch := range text {
		if ch == '\r' {
			result.X = sprec.Max(result.X, currentWidth)
			currentWidth = 0.0
			lastGlyph = nil
			continue
		}
		if ch == '\n' {
			result.X = sprec.Max(result.X, currentWidth)
			result.Y += f.lineHeight - (f.lineAscent + f.lineDescent)
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
	result = sprec.Vec2Prod(result, float32(fontSize))

	return ui.NewSize(int(result.X), int(result.Y))
}

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
