package internal

import "github.com/mokiat/gomath/sprec"

const (
	initialTextCharacterCount = 1024
	initialTextParagraphCount = 4
)

func newText() *Text {
	return &Text{
		characters: make([]rune, 0, initialTextCharacterCount),
		paragraphs: make([]Paragraph, 0, initialTextParagraphCount),
	}
}

type Text struct {
	font       *Font
	fontSize   float32
	color      sprec.Vec4
	characters []rune
	paragraphs []Paragraph
}

func (t *Text) Init(typography Typography) {
	t.font = typography.Font
	t.fontSize = typography.Size
	t.color = typography.Color
	t.characters = t.characters[:0]
	t.paragraphs = t.paragraphs[:0]
}

func (t *Text) Write(value string, position sprec.Vec2) {
	charOffset := len(t.characters)
	charCount := 0
	for _, ch := range value {
		t.characters = append(t.characters, ch)
		charCount++
	}
	t.paragraphs = append(t.paragraphs, Paragraph{
		position:   position,
		charOffset: charOffset,
		charCount:  charCount,
	})
}

type Paragraph struct {
	position   sprec.Vec2
	charOffset int
	charCount  int
}
