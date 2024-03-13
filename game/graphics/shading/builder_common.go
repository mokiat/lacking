package shading

type CommonBuilder interface {
	Value(value float32) Vec1Variable

	Vec1() Vec1Variable
	Vec2() Vec2Variable
	Vec3() Vec3Variable
	Vec4() Vec4Variable

	UniformVec1() Vec1Variable
	UniformVec2() Vec2Variable
	UniformVec3() Vec3Variable
	UniformVec4() Vec4Variable

	AssignVec1(target Vec1Variable, x Vec1Variable)
	AssignVec2(target Vec2Variable, x Vec1Variable, y Vec1Variable)
	AssignVec3(target Vec3Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable)
	AssignVec4(target Vec4Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable, w Vec1Variable)
}
