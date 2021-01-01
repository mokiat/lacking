package ui

import "github.com/mokiat/gomath/sprec"

type Canvas interface {
	Push()
	Pop()

	Clip(bounds Bounds)

	Translate(delta sprec.Vec2)

	SetTypeface(font Typeface)
	Typeface() Typeface

	SetStrokeActive(enabled bool)
	IsStrokeActive() bool

	SetStrokeColor(color Color)
	StrokeColor() Color

	SetStrokeSize(size int)
	StrokeSize() int

	SetFillActive(active bool)
	IsFillActive() bool

	SetFillColor(color Color)
	FillColor() Color

	DrawRectangle(position, size sprec.Vec2)
	DrawCircle(position sprec.Vec2, radius float32)
	DrawLine(start, end sprec.Vec2)
	DrawText(text string, position sprec.Vec2)
	DrawImage(image Image, position sprec.Vec2)
}
