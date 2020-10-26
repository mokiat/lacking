package shape

func NewStaticSphere(radius float32) StaticSphere {
	return StaticSphere{
		radius: radius,
	}
}

type StaticSphere struct {
	radius float32
}

func (s StaticSphere) Radius() float32 {
	return s.radius
}
