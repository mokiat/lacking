package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

// Use the following links as reference for font terminology:
// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
// https://www.freetype.org/freetype2/docs/glyphs/glyphs-3.html

const (
	fontImageSize  = 2048
	fontImageCells = 16
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
		'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ', 'Ъ', 'ѝ', 'Ю', 'Я', '№', '<', '>', ' ', '\t',
	}
}

func NewFontFactory(renderer *Renderer) *FontFactory {
	return &FontFactory{
		renderer:            renderer,
		colorTexture:        opengl.NewTwoDTexture(),
		depthStencilTexture: opengl.NewTwoDTexture(),
		framebuffer:         opengl.NewFramebuffer(),
		buf:                 &sfnt.Buffer{},
	}
}

type FontFactory struct {
	renderer            *Renderer
	colorTexture        *opengl.TwoDTexture
	depthStencilTexture *opengl.TwoDTexture
	framebuffer         *opengl.Framebuffer
	buf                 *sfnt.Buffer
}

func (f *FontFactory) Init() {
	f.colorTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:          fontImageSize,
		Height:         fontImageSize,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		InternalFormat: gl.SRGB8_ALPHA8,
	})

	f.depthStencilTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:          fontImageSize,
		Height:         fontImageSize,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		InternalFormat: gl.DEPTH24_STENCIL8,
	})

	f.framebuffer.Allocate(opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&f.colorTexture.Texture,
		},
		DepthStencilAttachment: &f.depthStencilTexture.Texture,
	})
}

func (f *FontFactory) Free() {
	defer f.colorTexture.Release()
	defer f.depthStencilTexture.Release()
	defer f.framebuffer.Release()
}

func (f *FontFactory) CreateFont(font *opentype.Font) *Font {
	startTime := time.Now()

	result := &Font{
		familyName:    f.fontFamilyName(font),
		subFamilyName: f.fontSubFamilyName(font),
		texture:       opengl.NewTwoDTexture(),
		glyphs:        make(map[rune]*fontGlyph),
	}

	cellSize := (fontImageSize / fontImageCells)
	// Use 4% padding to ensure that glyphs don't touch
	padding := cellSize / 25
	contentSize := cellSize - 2*padding
	// One em is roughly the size of the area where a single glyph is drawn
	// though with modern fonts glyphs can overflow that area.
	// Still, we use that to determine roughly how many pixels we'd like for
	// each glyph (ppem) so that we don't get many fixed point rounding errors.
	ppem := fixed.I(contentSize)

	metrics, err := font.Metrics(f.buf, ppem, xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to get font metrics: %w", err))
	}

	// We normalize the font based on the maximum ascent and maximum descent
	// in order to have consistent font sizes, irrelevant of the font design.
	scale := 1.0 / (fixedToFloat(metrics.Ascent) + fixedToFloat(metrics.Descent))
	result.lineHeight = fixedToFloat(metrics.Height) * scale
	result.lineAscent = fixedToFloat(metrics.Ascent) * scale
	result.lineDescent = fixedToFloat(metrics.Descent) * scale

	f.framebuffer.ClearColor(0, sprec.NewVec4(0.0, 0.0, 0.0, 0.0))
	f.framebuffer.ClearDepth(1.0)
	f.framebuffer.ClearStencil(0)
	f.renderer.Begin(Target{
		Framebuffer: f.framebuffer,
		Size:        sprec.NewVec2(float32(fontImageSize), float32(fontImageSize)),
	})

	for i, ch := range supportedCharacters {
		chIndex, err := font.GlyphIndex(f.buf, ch)
		if err != nil {
			panic(fmt.Errorf("failed to find char (%c) index: %w", ch, err))
		}

		segments, err := font.LoadGlyph(f.buf, chIndex, ppem, nil)
		if err != nil {
			panic(fmt.Errorf("failed to load glyph (%c): %w", ch, err))
		}

		bounds, advance, err := font.GlyphBounds(f.buf, chIndex, ppem, xfont.HintingNone)
		if err != nil {
			panic(fmt.Errorf("failed to get glyph (%c) bounds: %w", ch, err))
		}

		cellStartX := float32((i%fontImageCells)*cellSize) + float32(padding)
		cellEndX := cellStartX + float32(contentSize)
		cellStartY := float32((i/fontImageCells)*cellSize) + float32(padding)
		cellEndY := cellStartY + float32(contentSize)

		glyphWidth := fixedToFloat(bounds.Max.X - bounds.Min.X)
		glyphHeight := fixedToFloat(bounds.Max.Y - bounds.Min.Y)
		determinant := (glyphWidth - glyphHeight) * float32(contentSize)

		// Make sure to preserve glyph proportions, otherwise mipmapping
		// causes bad artifacts on elongated glyphs.
		var boxPosition sprec.Vec2
		var boxSize sprec.Vec2
		if glyphWidth > glyphHeight {
			boxPosition = sprec.NewVec2(cellStartX, cellStartY+determinant/(2.0*glyphWidth))
			boxSize = sprec.NewVec2(float32(contentSize), (float32(contentSize)*glyphHeight)/glyphWidth)
		} else {
			boxPosition = sprec.NewVec2(cellStartX-determinant/(2.0*glyphHeight), cellStartY)
			boxSize = sprec.NewVec2((float32(contentSize)*glyphWidth)/glyphHeight, float32(contentSize))
		}

		f.renderer.SetClipBounds(
			cellStartX, cellEndX,
			cellStartY, cellEndY,
		)
		f.renderer.SetTransform(sprec.Mat4MultiProd(
			sprec.TranslationMat4(
				float32(boxPosition.X),
				float32(boxPosition.Y),
				0.0,
			),
			sprec.ScaleMat4(
				float32(boxSize.X)/fixedToFloat(bounds.Max.X-bounds.Min.X),
				float32(boxSize.Y)/fixedToFloat(bounds.Max.Y-bounds.Min.Y),
				1.0,
			),
			sprec.TranslationMat4(
				-fixedToFloat(bounds.Min.X),
				-fixedToFloat(bounds.Min.Y),
				0.0,
			),
		))

		shape := f.renderer.BeginShape(Fill{
			color: sprec.NewVec4(1.0, 1.0, 1.0, 1.0),
			mode:  StencilModeNonZero,
		})

		for _, segment := range segments {
			switch segment.Op {
			case sfnt.SegmentOpMoveTo:
				shape.MoveTo(
					sprec.NewVec2(
						fixedToFloat(segment.Args[0].X),
						fixedToFloat(segment.Args[0].Y),
					),
				)
			case sfnt.SegmentOpLineTo:
				shape.LineTo(
					sprec.NewVec2(
						fixedToFloat(segment.Args[0].X),
						fixedToFloat(segment.Args[0].Y),
					),
				)
			case sfnt.SegmentOpQuadTo:
				shape.QuadTo(
					sprec.NewVec2(
						fixedToFloat(segment.Args[0].X),
						fixedToFloat(segment.Args[0].Y),
					),
					sprec.NewVec2(
						fixedToFloat(segment.Args[1].X),
						fixedToFloat(segment.Args[1].Y),
					),
				)
			case sfnt.SegmentOpCubeTo:
				shape.CubeTo(
					sprec.NewVec2(
						fixedToFloat(segment.Args[0].X),
						fixedToFloat(segment.Args[0].Y),
					),
					sprec.NewVec2(
						fixedToFloat(segment.Args[1].X),
						fixedToFloat(segment.Args[1].Y),
					),
					sprec.NewVec2(
						fixedToFloat(segment.Args[2].X),
						fixedToFloat(segment.Args[2].Y),
					),
				)
			default:
				panic(fmt.Errorf("unknown segment operation %d", segment.Op))
			}
		}

		f.renderer.EndShape(shape)

		result.glyphs[ch] = &fontGlyph{
			leftU:   boxPosition.X / float32(fontImageSize),
			rightU:  (boxPosition.X + boxSize.X) / float32(fontImageSize),
			topV:    1.0 - (boxPosition.Y)/float32(fontImageSize),
			bottomV: 1.0 - (boxPosition.Y+boxSize.Y)/float32(fontImageSize),

			advance:      fixedToFloat(advance) * scale,
			ascent:       -fixedToFloat(bounds.Min.Y) * scale,
			descent:      fixedToFloat(bounds.Max.Y) * scale,
			leftBearing:  fixedToFloat(bounds.Min.X) * scale,
			rightBearing: fixedToFloat(advance-bounds.Max.X) * scale,

			kerns: make(map[rune]float32),
		}

		for _, targetCh := range supportedCharacters {
			targetChIndex, err := font.GlyphIndex(f.buf, targetCh)
			if err != nil {
				panic(fmt.Errorf("failed to find char (%c) index: %w", ch, err))
			}

			kern, err := font.Kern(f.buf, chIndex, targetChIndex, ppem, xfont.HintingNone)
			if err != nil {
				panic(fmt.Errorf("failed to find kern (%c - %c): %w", ch, targetCh, err))
			}

			if kern != 0 {
				result.glyphs[ch].kerns[targetCh] = fixedToFloat(kern) * scale
			}
		}
	}

	f.renderer.End()

	result.texture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:              fontImageSize,
		Height:             fontImageSize,
		WrapS:              gl.CLAMP_TO_EDGE,
		WrapT:              gl.CLAMP_TO_EDGE,
		MinFilter:          gl.LINEAR_MIPMAP_LINEAR,
		MagFilter:          gl.LINEAR,
		UseAnisotropy:      false,
		PlaceholderMipmaps: true,
		GenerateMipmaps:    false,
		InternalFormat:     gl.SRGB8_ALPHA8,
	})

	gl.TextureBarrier()
	gl.CopyTextureSubImage2D(result.texture.ID(), 0, 0, 0, 0, 0, fontImageSize, fontImageSize)
	gl.GenerateTextureMipmap(result.texture.ID())

	elapsedTime := time.Since(startTime)
	fmt.Printf("Font creation time: %s\n", elapsedTime)

	return result
}

func (f *FontFactory) fontFamilyName(font *opentype.Font) string {
	familyName, err := font.Name(f.buf, 1)
	if err != nil {
		panic(fmt.Errorf("failed to get family name: %w", err))
	}
	return strings.ToLower(familyName)
}

func (f *FontFactory) fontSubFamilyName(font *opentype.Font) string {
	subFamilyName, err := font.Name(f.buf, 2)
	if err != nil {
		panic(fmt.Errorf("failed to get sub-family name: %w", err))
	}
	return strings.ToLower(subFamilyName)
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

func fixedToFloat(value fixed.Int26_6) float32 {
	if value > 0 {
		return float32(value>>6) + float32(value&0x3F)/float32(64)
	} else {
		return -float32(-value>>6) - float32(-value&0x3F)/float32(64)
	}
}
