package internal

import "github.com/mokiat/gomath/sprec"

const (
	initialTextCharacterCount = 1024
)

func newText() *Text {
	return &Text{
		characters: make([]rune, 0, initialTextCharacterCount),
	}
}

type Text struct {
	font       *Font
	fontSize   float32
	color      sprec.Vec4
	characters []rune
}

func (t *Text) Init(font *Font, fontSize float32, color sprec.Vec4) {
	t.font = font
	t.fontSize = fontSize
	t.color = color
	t.characters = t.characters[:0]
}

func (t *Text) Write(value string) {
	for _, ch := range value {
		t.characters = append(t.characters, ch)
	}
}
