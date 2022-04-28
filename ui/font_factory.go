package ui

import (
	"fmt"
	"strings"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
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

func newFontFactory(api render.API, renderer *canvasRenderer) *fontFactory {
	return &fontFactory{
		api:           api,
		renderer:      renderer,
		fontImageSize: 2048,
		buf:           &sfnt.Buffer{},
	}
}

type fontFactory struct {
	api            render.API
	renderer       *canvasRenderer
	fontImageSize  int
	colorTexture   render.Texture
	stencilTexture render.Texture
	framebuffer    render.Framebuffer
	buf            *sfnt.Buffer
}

func (f *fontFactory) Init() {
	switch f.api.Capabilities().Quality {
	case render.QualityHigh:
		f.fontImageSize = 2048
	case render.QualityMedium:
		f.fontImageSize = 1024
	case render.QualityLow:
		f.fontImageSize = 512
	}

	f.colorTexture = f.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           f.fontImageSize,
		Height:          f.fontImageSize,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})

	f.stencilTexture = f.api.CreateStencilTexture2D(render.StencilTexture2DInfo{
		Width:  f.fontImageSize,
		Height: f.fontImageSize,
	})

	f.framebuffer = f.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			f.colorTexture,
		},
		StencilAttachment: f.stencilTexture,
	})
}

func (f *fontFactory) Free() {
	defer f.colorTexture.Release()
	defer f.stencilTexture.Release()
	defer f.framebuffer.Release()
}

func (f *fontFactory) CreateFont(font *opentype.Font) (*Font, error) {
	f.api.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: f.framebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  f.fontImageSize,
			Height: f.fontImageSize,
		},
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationDontCare,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 0.0},
			},
		},
		DepthLoadOp:       render.LoadOperationDontCare,
		DepthStoreOp:      render.StoreOperationDontCare,
		StencilLoadOp:     render.LoadOperationClear,
		StencilStoreOp:    render.StoreOperationDontCare,
		StencilClearValue: 0x00,
	})
	f.renderer.onBegin(NewSize(f.fontImageSize, f.fontImageSize))

	cellSize := float32(f.fontImageSize) / float32(fontImageCells)
	// Use 4% padding to ensure that glyphs don't touch
	padding := 0.04 * cellSize
	contentSize := cellSize - 2.0*padding
	reader := fontReader{
		buf:  f.buf,
		font: font,
		// One em is roughly the size of the area where a single glyph is drawn
		// though with modern fonts glyphs can overflow that area.
		// Still, we use that to determine roughly how many pixels we'd like for
		// each glyph (ppem) so that we don't get many fixed point rounding errors.
		ppem: fixed.I(int(contentSize)),
	}

	metrics := reader.Metrics()

	// We normalize the font based on the maximum ascent and maximum descent
	// in order to have consistent font sizes, irrelevant of the font design.
	scale := 1.0 / (fixedToFloat(metrics.Ascent) + fixedToFloat(metrics.Descent))

	resultGlyphs := make(map[rune]*fontGlyph)

	for i, ch := range supportedCharacters {
		chIndex := reader.GlyphIndex(ch)
		segments, advance := reader.LoadGlyph(chIndex)
		bounds := segments.Bounds()

		glyphWidth := fixedToFloat(bounds.Max.X - bounds.Min.X)
		glyphHeight := fixedToFloat(bounds.Max.Y - bounds.Min.Y)
		determinant := (glyphWidth - glyphHeight) * contentSize

		cellStartX := float32(i%fontImageCells)*cellSize + padding
		cellEndX := cellStartX + contentSize
		cellStartY := float32(i/fontImageCells)*cellSize + padding
		cellEndY := cellStartY + contentSize

		// Make sure to preserve glyph proportions, otherwise mipmapping
		// causes bad artifacts on elongated glyphs.
		var boxPosition sprec.Vec2
		var boxSize sprec.Vec2
		if glyphWidth > glyphHeight {
			boxPosition = sprec.NewVec2(
				cellStartX,
				cellStartY+determinant/(2.0*glyphWidth),
			)
			boxSize = sprec.NewVec2(
				contentSize,
				(contentSize*glyphHeight)/glyphWidth,
			)
		} else {
			boxPosition = sprec.NewVec2(
				cellStartX-determinant/(2.0*glyphHeight),
				cellStartY,
			)
			boxSize = sprec.NewVec2(
				(contentSize*glyphWidth)/glyphHeight,
				contentSize,
			)
		}

		f.renderer.Push()
		f.renderer.SetClipBounds(
			cellStartX, cellEndX,
			cellStartY, cellEndY,
		)
		f.renderer.SetTransform(sprec.Mat4MultiProd(
			sprec.TranslationMat4(boxPosition.X, boxPosition.Y, 0.0),
			sprec.ScaleMat4(boxSize.X/glyphWidth, boxSize.Y/glyphHeight, 1.0),
			sprec.TranslationMat4(-fixedToFloat(bounds.Min.X), -fixedToFloat(bounds.Min.Y), 0.0),
		))

		shape := f.renderer.Shape()
		shape.Begin(Fill{
			Color: White(),
			Rule:  FillRuleNonZero,
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
		shape.End()
		f.renderer.Pop()

		resultKerns := make(map[rune]float32)
		for _, targetCh := range supportedCharacters {
			targetChIndex := reader.GlyphIndex(targetCh)

			if kern := reader.Kern(chIndex, targetChIndex); kern != 0 {
				resultKerns[targetCh] = fixedToFloat(kern) * scale
			}
		}

		resultGlyphs[ch] = &fontGlyph{
			leftU:   boxPosition.X / float32(f.fontImageSize),
			rightU:  (boxPosition.X + boxSize.X) / float32(f.fontImageSize),
			topV:    1.0 - (boxPosition.Y)/float32(f.fontImageSize),
			bottomV: 1.0 - (boxPosition.Y+boxSize.Y)/float32(f.fontImageSize),

			advance:      fixedToFloat(advance) * scale,
			ascent:       -fixedToFloat(bounds.Min.Y) * scale,
			descent:      fixedToFloat(bounds.Max.Y) * scale,
			leftBearing:  fixedToFloat(bounds.Min.X) * scale,
			rightBearing: fixedToFloat(advance-bounds.Max.X) * scale,

			kerns: resultKerns,
		}
	}
	f.renderer.onEnd()

	resultTexture := f.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           f.fontImageSize,
		Height:          f.fontImageSize,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeLinear,
		Mipmapping:      true,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	f.api.CopyContentToTexture(render.CopyContentToTextureInfo{
		Texture:         resultTexture,
		TextureLevel:    0,
		TextureX:        0,
		TextureY:        0,
		FramebufferX:    0,
		FramebufferY:    0,
		Width:           f.fontImageSize,
		Height:          f.fontImageSize,
		GenerateMipmaps: true,
	})

	f.api.EndRenderPass()

	return &Font{
		familyName:    reader.FontFamilyName(),
		subFamilyName: reader.FontSubFamilyName(),

		lineHeight:  fixedToFloat(metrics.Height) * scale,
		lineAscent:  fixedToFloat(metrics.Ascent) * scale,
		lineDescent: fixedToFloat(metrics.Descent) * scale,

		glyphs: resultGlyphs,

		texture: resultTexture,
	}, nil
}

type fontReader struct {
	buf  *sfnt.Buffer
	font *sfnt.Font
	ppem fixed.Int26_6
}

func (r *fontReader) FontFamilyName() string {
	familyName, err := r.font.Name(r.buf, 1)
	if err != nil {
		panic(fmt.Errorf("failed to get family name: %w", err))
	}
	return strings.ToLower(familyName)
}

func (r *fontReader) FontSubFamilyName() string {
	subFamilyName, err := r.font.Name(r.buf, 2)
	if err != nil {
		panic(fmt.Errorf("failed to get sub-family name: %w", err))
	}
	return strings.ToLower(subFamilyName)
}

func (r *fontReader) Metrics() xfont.Metrics {
	metrics, err := r.font.Metrics(r.buf, r.ppem, xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to get font metrics: %w", err))
	}
	return metrics
}

func (r *fontReader) GlyphIndex(ch rune) sfnt.GlyphIndex {
	chIndex, err := r.font.GlyphIndex(r.buf, ch)
	if err != nil {
		panic(fmt.Errorf("failed to find glyph (%c) index: %w", ch, err))
	}
	return chIndex
}

func (r *fontReader) LoadGlyph(chIndex sfnt.GlyphIndex) (sfnt.Segments, fixed.Int26_6) {
	segments, err := r.font.LoadGlyph(r.buf, chIndex, r.ppem, nil)
	if err != nil {
		panic(fmt.Errorf("failed to load glyph (%d): %w", chIndex, err))
	}

	advance, err := r.font.GlyphAdvance(r.buf, chIndex, r.ppem, xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to load glyph (%d) advance: %w", chIndex, err))
	}

	return segments, advance
}

func (r *fontReader) Kern(ch1Index, ch2Index sfnt.GlyphIndex) fixed.Int26_6 {
	kern, err := r.font.Kern(r.buf, ch1Index, ch2Index, r.ppem, xfont.HintingNone)
	if err != nil {
		panic(fmt.Errorf("failed to find kern (%d - %d): %w", ch1Index, ch2Index, err))
	}
	return kern
}

func fixedToFloat(value fixed.Int26_6) float32 {
	if value > 0 {
		return float32(value>>6) + float32(value&0x3F)/float32(64)
	} else {
		return -float32(-value>>6) - float32(-value&0x3F)/float32(64)
	}
}
