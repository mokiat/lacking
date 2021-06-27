package internal

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

// Use the following links as reference for font terminology:
// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
// https://www.freetype.org/freetype2/docs/glyphs/glyphs-3.html

const (
	fontImageSize = 2048
)

var supportedCharacters []rune

func init() {
	supportedCharacters = []rune{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
		'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F',
		'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V',
		'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '!', '"',
		'#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', ':', ';', '<',
		'=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~', 'а', 'б',
		'в', 'г', 'д', 'е', 'ж', 'з', 'и', 'й', 'к', 'л', 'м', 'н', 'о', 'п', 'р', 'с',
		'т', 'у', 'ф', 'х', 'ц', 'ч', 'ш', 'щ', 'ъ', 'ь', 'ю', 'я', 'А', 'Б', 'В', 'Г',
		'Д', 'Е', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М', 'Н', 'О', 'П', 'Р', 'С', 'Т', 'У',
		'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ', 'Ъ', 'ѝ', 'Ю', 'Я', '№', '<', '>',
	}
}

func NewFont(familyName, subFamilyName string) *Font {
	return &Font{
		texture:       opengl.NewTwoDTexture(),
		familyName:    strings.ToLower(familyName),
		subFamilyName: strings.ToLower(subFamilyName),
		glyphs:        make(map[rune]*fontGlyph),
	}
}

var _ ui.Font = (*Font)(nil)

type Font struct {
	familyName    string
	subFamilyName string

	// lineHeight holds the distance (positive) from one line baseline to the next.
	lineHeight float32
	// lineAscent holds the distance (positive) from the baseline to the top of the line.
	lineAscent float32
	// lineDescent holds the distance (positive) from the baseline to the bottom of the line.
	lineDescent float32
	// glyphs holds a map of supported glyphs.
	glyphs map[rune]*fontGlyph

	texture *opengl.TwoDTexture
}

func (f *Font) Family() string {
	return f.familyName
}

func (f *Font) SubFamily() string {
	return f.subFamilyName
}

var buf = &sfnt.Buffer{}

func (f *Font) Allocate(font *opentype.Font) {
	log.Printf("font: %s / %s", f.familyName, f.subFamilyName)

	src := image.NewUniform(color.White)
	dst := image.NewNRGBA(image.Rect(0, 0, fontImageSize, fontImageSize))

	cellSize := (fontImageSize / 16)
	fontSize := pickOptimalFontSize(font, cellSize)
	log.Printf("selected font size: %s", fontSize)

	metrics, err := font.Metrics(buf, fontSize, xfont.HintingNone)
	if err != nil {
		panic(err)
	}
	log.Printf("ascent: %s, descent: %s, lineheight: %s", metrics.Ascent, metrics.Descent, metrics.Height)

	mIndex, err := font.GlyphIndex(buf, 'M')
	if err != nil {
		panic(fmt.Errorf("failed to get index: %w", err))
	}

	bounds, _, err := font.GlyphBounds(buf, mIndex, fontSize, xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to get bounds: %w", err))
	}

	log.Printf("char M ascent: %s, descent: %s, height: %s", bounds.Min.Y, bounds.Max.Y, bounds.Max.Y-bounds.Min.Y)

	// font.Metrics(b *sfnt.Buffer, ppem fixed.Int26_6, h xfont.Hinting)

	// font.GlyphBounds(b *sfnt.Buffer, x sfnt.GlyphIndex, ppem fixed.Int26_6, h xfont.Hinting)

	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size:    24.0 * 4,
		DPI:     72.0, // normal screen dpi
		Hinting: xfont.HintingNone,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create face: %w", err))
	}
	defer face.Close()

	drawer := xfont.Drawer{
		Src:  src,
		Dst:  dst,
		Face: face,
	}
	for i, ch := range supportedCharacters {
		cellX := i % 16
		cellY := i / 16
		drawer.Dot = fixed.P(cellX*32*4+2*4, cellY*32*4+26*4)
		drawer.DrawString(string(ch))
	}

	f.texture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             fontImageSize,
		Height:            fontImageSize,
		MinFilter:         gl.LINEAR,
		MagFilter:         gl.LINEAR,
		UseAnisotropy:     false,
		GenerateMipmaps:   false,
		InternalFormat:    gl.SRGB8_ALPHA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              dst.Pix,
	})
}

func (f *Font) Release() {
	f.texture.Release()
}

func pickOptimalFontSize(font *opentype.Font, cellSize int) fixed.Int26_6 {
	minFontSize := 1
	maxFontSize := cellSize
	for minFontSize < maxFontSize-1 {
		avgFontSize := (minFontSize + maxFontSize) / 2
		metrics, err := font.Metrics(buf, fixed.I(avgFontSize), xfont.HintingNone)
		if err != nil {
			panic(fmt.Errorf("failed to get font metrics: %w", err))
		}
		if (metrics.Ascent + metrics.Descent).Ceil() > cellSize {
			maxFontSize = avgFontSize - 1
		} else {
			minFontSize = avgFontSize
		}
	}
	return fixed.I(minFontSize)
}

type fontGlyph struct {
	// lowerLeftU holds the U texture coordinate of the lower-left corner of the
	// glyph bounds.
	lowerLeftU float32
	// lowerLeftV holds the V texture coordinate of the lower-left corner of the
	// glyph bounds.
	lowerLeftV float32
	// upperRightU holds the U texture coordinate of the upper-right corner of the
	// glyph bounds.
	upperRightU float32
	// upperRightV holds the V texture coordinate of the upper-right corner of the
	// glyph bounds.
	upperRightV float32
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
