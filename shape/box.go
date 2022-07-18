package shape

import "github.com/mokiat/gomath/sprec"

func NewStaticBox(width, height, length float32) StaticBox {
	return StaticBox{
		width:  width,
		height: height,
		length: length,
	}
}

type StaticBox struct {
	width  float32
	height float32
	length float32
}

func (b StaticBox) BoundingSphereRadius() float32 {
	return sprec.Sqrt(b.Width()*b.Width()+b.Height()*b.Height()+b.Length()*b.Length()) / 2.0
}

func (b StaticBox) Width() float32 {
	return b.width
}

func (b StaticBox) Height() float32 {
	return b.height
}

func (b StaticBox) Length() float32 {
	return b.length
}
