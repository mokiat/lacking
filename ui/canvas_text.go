package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
	initialTextCharacterCount = 1024
	initialTextParagraphCount = 4
)

func newText(state *canvasState, shaders ShaderCollection) *Text {
	return &Text{
		state: state,

		textMesh:     newTextMesh(maxVertexCount),
		textMaterial: newMaterial(shaders.TextMaterial),

		characters: make([]rune, 0, initialTextCharacterCount),
		paragraphs: make([]paragraph, 0, initialTextParagraphCount),
	}
}

// Text represents a module for drawing text.
type Text struct {
	state        *canvasState
	commandQueue render.CommandQueue

	textMesh     *TextMesh
	textMaterial *Material
	textPipeline render.Pipeline

	engaged bool

	clipBounds      sprec.Vec4
	transformMatrix sprec.Mat4

	font       *Font
	fontSize   float32
	color      sprec.Vec4
	characters []rune
	paragraphs []paragraph
}

func (t *Text) onCreate(api render.API, commandQueue render.CommandQueue) {
	t.commandQueue = commandQueue
	t.textMesh.Allocate(api)
	t.textMaterial.Allocate(api)
	t.textPipeline = api.CreatePipeline(render.PipelineInfo{
		Program:                     t.textMaterial.program,
		VertexArray:                 t.textMesh.vertexArray,
		Topology:                    render.TopologyTriangles,
		Culling:                     render.CullModeNone,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		StencilTest:                 false,
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
}

func (t *Text) onDestroy() {
	defer t.textMesh.Release()
	defer t.textMaterial.Release()
	defer t.textPipeline.Release()
}

func (t *Text) onBegin() {
	t.textMesh.Reset()
}

func (t *Text) onEnd() {
	t.textMesh.Update()
}

// Begin starts a new text sequence using the specified typography settings.
// Make sure to use End when finished with the text.
func (t *Text) Begin(typography Typography) {
	if t.engaged {
		panic("text already started")
	}
	t.engaged = true

	currentLayer := t.state.currentLayer
	t.clipBounds = sprec.NewVec4(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	t.transformMatrix = currentLayer.Transform

	t.font = typography.Font
	t.fontSize = typography.Size
	t.color = uiColorToVec(typography.Color)
	t.characters = t.characters[:0]
	t.paragraphs = t.paragraphs[:0]
}

// End marks the end of the text and pushes all collected data for
// drawing.
func (t *Text) End() {
	if !t.engaged {
		panic("text already ended")
	}
	t.engaged = false

	vertexOffset := t.textMesh.Offset()
	for _, paragraph := range t.paragraphs {
		offset := paragraph.position
		lastGlyph := (*fontGlyph)(nil)

		paragraphChars := t.characters[paragraph.charOffset : paragraph.charOffset+paragraph.charCount]
		for _, ch := range paragraphChars {
			lineHeight := t.font.lineHeight * t.fontSize
			lineAscent := t.font.lineAscent * t.fontSize
			if ch == '\r' {
				offset.X = paragraph.position.X
				lastGlyph = nil
				continue
			}
			if ch == '\n' {
				offset.X = paragraph.position.X
				offset.Y += lineHeight
				lastGlyph = nil
				continue
			}

			if glyph, ok := t.font.glyphs[ch]; ok {
				advance := glyph.advance * t.fontSize
				leftBearing := glyph.leftBearing * t.fontSize
				rightBearing := glyph.rightBearing * t.fontSize
				ascent := glyph.ascent * t.fontSize
				descent := glyph.descent * t.fontSize

				vertTopLeft := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							leftBearing,
							lineAscent-ascent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.leftU, glyph.topV),
				}
				vertTopRight := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							advance-rightBearing,
							lineAscent-ascent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.rightU, glyph.topV),
				}
				vertBottomLeft := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							leftBearing,
							lineAscent+descent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.leftU, glyph.bottomV),
				}
				vertBottomRight := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							advance-rightBearing,
							lineAscent+descent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.rightU, glyph.bottomV),
				}

				t.textMesh.Append(vertTopLeft)
				t.textMesh.Append(vertBottomLeft)
				t.textMesh.Append(vertBottomRight)

				t.textMesh.Append(vertTopLeft)
				t.textMesh.Append(vertBottomRight)
				t.textMesh.Append(vertTopRight)

				offset.X += advance
				if lastGlyph != nil {
					offset.X += lastGlyph.kerns[ch] * t.fontSize
				}
				lastGlyph = glyph
			}
		}
	}
	vertexCount := t.textMesh.Offset() - vertexOffset

	t.commandQueue.BindPipeline(t.textPipeline)
	t.commandQueue.Uniform4f(t.textMaterial.clipDistancesLocation, t.clipBounds.Array())
	t.commandQueue.Uniform4f(t.textMaterial.colorLocation, t.color.Array())
	t.commandQueue.UniformMatrix4f(t.textMaterial.projectionMatrixLocation, t.state.projectionMatrix.ColumnMajorArray())
	t.commandQueue.UniformMatrix4f(t.textMaterial.transformMatrixLocation, t.transformMatrix.ColumnMajorArray())
	t.commandQueue.TextureUnit(0, t.font.texture)
	t.commandQueue.Uniform1i(t.textMaterial.textureLocation, 0)
	t.commandQueue.Draw(vertexOffset, vertexCount, 1)
}

// Line draws a text line at the specified position.
func (t *Text) Line(value string, position sprec.Vec2) {
	charOffset := len(t.characters)
	charCount := 0
	for _, ch := range value {
		t.characters = append(t.characters, ch)
		charCount++
	}
	t.paragraphs = append(t.paragraphs, paragraph{
		position:   position,
		charOffset: charOffset,
		charCount:  charCount,
	})
}

// Typography configures how text is to be drawn.
type Typography struct {

	// Font specifies the font to be used.
	Font *Font

	// Size specifies the font size.
	Size float32

	// Color indicates the color of the text.
	Color Color
}

type paragraph struct {
	position   sprec.Vec2
	charOffset int
	charCount  int
}
