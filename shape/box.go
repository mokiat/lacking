package shape

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

func (b StaticBox) Width() float32 {
	return b.width
}

func (b StaticBox) Height() float32 {
	return b.height
}

func (b StaticBox) Length() float32 {
	return b.length
}
