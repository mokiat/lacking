package internal

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	}
}

type FontFactory struct {
	renderer            *Renderer
	colorTexture        *opengl.TwoDTexture
	depthStencilTexture *opengl.TwoDTexture
	framebuffer         *opengl.Framebuffer
}

func (f *FontFactory) Init() {
	f.colorTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:          fontImageSize,
		Height:         fontImageSize,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		InternalFormat: gl.RGBA8,
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

	result := NewFont()

	familyName, err := font.Name(buf, 1)
	if err != nil {
		panic(fmt.Errorf("failed to get family name: %w", err))
	}
	result.familyName = strings.ToLower(familyName)

	subFamilyName, err := font.Name(buf, 2)
	if err != nil {
		panic(fmt.Errorf("failed to get sub-family name: %w", err))
	}
	result.subFamilyName = strings.ToLower(subFamilyName)

	cellSize := (fontImageSize / fontImageCells)
	fontSize := pickOptimalFontSize(font, cellSize)

	metrics, err := font.Metrics(buf, fixed.I(fontSize), xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to get font metrics: %w", err))
	}

	scale := 1.0 / float32(metrics.Ascent.Round()+metrics.Descent.Round())
	result.lineHeight = float32(metrics.Height.Round()) * scale
	result.lineAscent = float32(metrics.Ascent.Round()) * scale
	result.lineDescent = float32(metrics.Descent.Round()) * scale

	f.renderer.Begin(Target{
		Framebuffer: f.framebuffer,
		Size:        sprec.NewVec2(float32(fontImageSize), float32(fontImageSize)),
	})

	f.framebuffer.ClearColor(0, sprec.NewVec4(0.0, 0.0, 0.0, 0.0))
	f.framebuffer.ClearDepth(1.0)
	f.framebuffer.ClearStencil(0)

	for i, ch := range supportedCharacters {
		chIndex, err := font.GlyphIndex(buf, ch)
		if err != nil {
			panic(fmt.Errorf("failed to find char index: %w", err))
		}

		segments, err := font.LoadGlyph(buf, chIndex, fixed.I(fontSize), nil)
		if err != nil {
			panic(fmt.Errorf("failed to load glyph index: %w", err))
		}

		bounds, advance, err := font.GlyphBounds(buf, chIndex, fixed.I(fontSize), xfont.HintingNone)
		if err != nil {
			panic(fmt.Errorf("failed to get glyph bounds: %w", err))
		}

		leftPx := (i % fontImageCells) * cellSize
		rightPx := leftPx + bounds.Max.Sub(bounds.Min).X.Floor()
		topPx := (i / fontImageCells) * cellSize
		bottomPx := topPx + bounds.Max.Sub(bounds.Min).Y.Floor()

		f.renderer.SetClipBounds(
			0, float32(fontImageSize),
			0, float32(fontImageSize),
		)

		f.renderer.SetTransform(
			sprec.TranslationMat4(float32(leftPx), float32(rightPx), 0.0),
		)

		shape := f.renderer.BeginShape(Fill{
			color: sprec.NewVec4(1.0, 0.0, 0.0, 1.0),
		})

		const scale = 1

		for _, segment := range segments {
			switch segment.Op {
			case sfnt.SegmentOpMoveTo:
				// log.Printf("move to (%d, %d)\n",
				// 	segment.Args[0].X.Floor()*scale,
				// 	segment.Args[0].Y.Floor()*scale+400,
				// )
				shape.MoveTo(
					sprec.NewVec2(
						float32(segment.Args[0].X.Floor()*scale),
						float32(segment.Args[0].Y.Floor()*scale),
					),
				)

			case sfnt.SegmentOpLineTo:
				// log.Printf("line to (%d, %d)\n",
				// 	segment.Args[0].X.Floor()*scale,
				// 	segment.Args[0].Y.Floor()*scale+400,
				// )
				shape.LineTo(
					sprec.NewVec2(
						float32(segment.Args[0].X.Floor()*scale),
						float32(segment.Args[0].Y.Floor()*scale),
					),
				)

			case sfnt.SegmentOpQuadTo:
				// log.Printf("quad to (%d, %d) (%d, %d)\n",
				// 	segment.Args[0].X.Floor()*scale,
				// 	segment.Args[0].Y.Floor()*scale+400,
				// 	segment.Args[1].X.Floor()*scale,
				// 	segment.Args[1].Y.Floor()*scale+400,
				// )
				shape.QuadTo(
					sprec.NewVec2(
						float32(segment.Args[0].X.Floor()*scale),
						float32(segment.Args[0].Y.Floor()*scale),
					),
					sprec.NewVec2(
						float32(segment.Args[1].X.Floor()*scale),
						float32(segment.Args[1].Y.Floor()*scale),
					),
				)

			case sfnt.SegmentOpCubeTo:
				// log.Printf("cube to (%d, %d) (%d, %d) (%d, %d)\n",
				// 	segment.Args[0].X.Floor()*scale,
				// 	segment.Args[0].Y.Floor()*scale+400,
				// 	segment.Args[1].X.Floor()*scale,
				// 	segment.Args[1].Y.Floor()*scale+400,
				// 	segment.Args[2].X.Floor()*scale,
				// 	segment.Args[2].Y.Floor()*scale+400,
				// )
				shape.CubeTo(
					sprec.NewVec2(
						float32(segment.Args[0].X.Floor()*scale),
						float32(segment.Args[0].Y.Floor()*scale),
					),
					sprec.NewVec2(
						float32(segment.Args[1].X.Floor()*scale),
						float32(segment.Args[1].Y.Floor()*scale),
					),
					sprec.NewVec2(
						float32(segment.Args[2].X.Floor()*scale),
						float32(segment.Args[2].Y.Floor()*scale),
					),
				)

			default:
				panic("unknown segment")
			}
		}

		f.renderer.EndShape(shape)

		result.glyphs[ch] = &fontGlyph{
			leftU:        float32(leftPx) / float32(fontImageSize),
			rightU:       float32(rightPx) / float32(fontImageSize),
			topV:         float32(topPx) / float32(fontImageSize),
			bottomV:      float32(bottomPx) / float32(fontImageSize),
			advance:      float32(advance.Round()) * scale,
			ascent:       float32(-bounds.Min.Y.Round()) * scale,
			descent:      float32(bounds.Max.Y.Round()) * scale,
			leftBearing:  float32(bounds.Min.X.Round()) * scale,
			rightBearing: float32((advance - bounds.Max.X).Round()) * scale,
			kerns:        make(map[rune]float32),
		}

		for _, targetCh := range supportedCharacters {
			targetChIndex, err := font.GlyphIndex(buf, targetCh)
			if err != nil {
				panic(fmt.Errorf("failed to find char index: %w", err))
			}
			kern, err := font.Kern(buf, chIndex, targetChIndex, fixed.I(fontSize), xfont.HintingNone)
			if err != nil {
				panic(fmt.Errorf("failed to find kern: %w", err))
			}
			if kern.Ceil() == 0 {
				continue
			}
			result.glyphs[ch].kerns[targetCh] = float32(kern.Ceil()) * scale
		}

		if true {
			break
		}
	}

	f.renderer.End()

	result.texture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:           fontImageSize,
		Height:          fontImageSize,
		WrapS:           gl.CLAMP_TO_EDGE,
		WrapT:           gl.CLAMP_TO_EDGE,
		MinFilter:       gl.LINEAR_MIPMAP_LINEAR,
		MagFilter:       gl.LINEAR,
		UseAnisotropy:   false,
		GenerateMipmaps: true,
		InternalFormat:  gl.SRGB8_ALPHA8,
	})

	gl.CopyTextureSubImage2D(result.texture.ID(), 0, 0, 0, 0, 0, fontImageSize, fontImageSize)

	elapsedTime := time.Since(startTime)
	fmt.Printf("Font creation time: %s\n", elapsedTime)

	return result
}

func NewFont() *Font {
	return &Font{
		texture: opengl.NewTwoDTexture(),
		glyphs:  make(map[rune]*fontGlyph),
	}
}

var _ ui.Font = (*Font)(nil)

var buf = &sfnt.Buffer{}

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

func (f *Font) Allocate(font *opentype.Font) {
	startTime := time.Now()

	familyName, err := font.Name(buf, 1)
	if err != nil {
		panic(fmt.Errorf("failed to get family name: %w", err))
	}
	f.familyName = strings.ToLower(familyName)

	subFamilyName, err := font.Name(buf, 2)
	if err != nil {
		panic(fmt.Errorf("failed to get sub-family name: %w", err))
	}
	f.subFamilyName = strings.ToLower(subFamilyName)

	src := image.NewUniform(color.White)
	dst := image.NewNRGBA(image.Rect(0, 0, fontImageSize, fontImageSize))

	cellSize := (fontImageSize / fontImageCells)
	fontSize := pickOptimalFontSize(font, cellSize)

	metrics, err := font.Metrics(buf, fixed.I(fontSize), xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to get font metrics: %w", err))
	}

	scale := 1.0 / float32(metrics.Ascent.Round()+metrics.Descent.Round())
	f.lineHeight = float32(metrics.Height.Round()) * scale
	f.lineAscent = float32(metrics.Ascent.Round()) * scale
	f.lineDescent = float32(metrics.Descent.Round()) * scale

	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72.0, // normal screen dpi
		Hinting: xfont.HintingNone,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create face: %w", err))
	}
	defer face.Close()

	// TODO: Render the font yourself using OpenGL once you have
	// narrowed down polygon rendering.

	for i, ch := range supportedCharacters {
		chIndex, err := font.GlyphIndex(buf, ch)
		if err != nil {
			panic(fmt.Errorf("failed to find char index: %w", err))
		}

		drawRect, mask, maskPos, _, ok := face.Glyph(fixed.P(0, 0), ch)
		if !ok {
			panic(fmt.Errorf("failed to find glyph for character %c", ch))
		}

		leftPx := (i % fontImageCells) * cellSize
		rightPx := leftPx + drawRect.Dx()
		topPx := (i / fontImageCells) * cellSize
		bottomPx := topPx + drawRect.Dy()

		draw.DrawMask(
			dst,
			image.Rect(leftPx, topPx, rightPx, bottomPx),
			src,
			image.Pt(0, 0),
			mask,
			maskPos,
			draw.Src,
		)

		bounds, advance, err := font.GlyphBounds(buf, chIndex, fixed.I(fontSize), xfont.HintingNone)
		if err != nil {
			panic(fmt.Errorf("failed to get glyph bounds: %w", err))
		}

		f.glyphs[ch] = &fontGlyph{
			leftU:        float32(leftPx) / float32(fontImageSize),
			rightU:       float32(rightPx) / float32(fontImageSize),
			topV:         float32(topPx) / float32(fontImageSize),
			bottomV:      float32(bottomPx) / float32(fontImageSize),
			advance:      float32(advance.Round()) * scale,
			ascent:       float32(-bounds.Min.Y.Round()) * scale,
			descent:      float32(bounds.Max.Y.Round()) * scale,
			leftBearing:  float32(bounds.Min.X.Round()) * scale,
			rightBearing: float32((advance - bounds.Max.X).Round()) * scale,
			kerns:        make(map[rune]float32),
		}

		for _, targetCh := range supportedCharacters {
			targetChIndex, err := font.GlyphIndex(buf, targetCh)
			if err != nil {
				panic(fmt.Errorf("failed to find char index: %w", err))
			}
			kern, err := font.Kern(buf, chIndex, targetChIndex, fixed.I(fontSize), xfont.HintingNone)
			if err != nil {
				panic(fmt.Errorf("failed to find kern: %w", err))
			}
			if kern.Ceil() == 0 {
				continue
			}
			f.glyphs[ch].kerns[targetCh] = float32(kern.Ceil()) * scale
		}
	}

	f.texture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             fontImageSize,
		Height:            fontImageSize,
		WrapS:             gl.CLAMP_TO_EDGE,
		WrapT:             gl.CLAMP_TO_EDGE,
		MinFilter:         gl.LINEAR_MIPMAP_LINEAR,
		MagFilter:         gl.LINEAR,
		UseAnisotropy:     false,
		GenerateMipmaps:   true,
		InternalFormat:    gl.SRGB8_ALPHA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              dst.Pix,
	})

	elapsedTime := time.Since(startTime)
	fmt.Printf("Font creation time: %s\n", elapsedTime)
}

func (f *Font) Destroy() {
	f.texture.Release()
}

func pickOptimalFontSize(font *opentype.Font, cellSize int) int {
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
	return minFontSize
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
