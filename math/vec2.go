package math

func NewVec2(x, y float32) Vec2 {
	return Vec2{
		X: x,
		Y: y,
	}
}

func Vec2Sum(a, b Vec2) Vec2 {
	return Vec2{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func Vec2Diff(a, b Vec2) Vec2 {
	return Vec2{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}

func Vec2Quot(vector Vec2, value float32) Vec2 {
	return Vec2{
		X: vector.X / value,
		Y: vector.Y / value,
	}
}
func Vec2Dot(a, b Vec2) float32 {
	return a.X*b.X + a.Y*b.Y
}

func UnitVec2(vector Vec2) Vec2 {
	return Vec2Quot(vector, vector.Length())
}

type Vec2 struct {
	X float32
	Y float32
}

func (v Vec2) IsZero() bool {
	return Eq32(v.X, 0.0) && Eq32(v.Y, 0.0)
}

func (v Vec2) Length() float32 {
	return Sqrt32(Vec2Dot(v, v))
}
