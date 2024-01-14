package shading

type Node interface {
	OutputParams() []Parameter
}

type ConstVec4Node struct {
	outVec     *Vec4Param
	x, y, z, w float32
}

func (n *ConstVec4Node) OutputParams() []Parameter {
	return []Parameter{n.outVec}
}

func (n *ConstVec4Node) OutVec() *Vec4Param {
	return n.outVec
}

func (n *ConstVec4Node) X() float32 {
	return n.x
}

func (n *ConstVec4Node) Y() float32 {
	return n.y
}

func (n *ConstVec4Node) Z() float32 {
	return n.z
}

func (n *ConstVec4Node) W() float32 {
	return n.w
}

type MulVec4Node struct {
	inVec  *Vec4Param
	ratio  float32
	outVec *Vec4Param
}

func (n *MulVec4Node) OutputParams() []Parameter {
	return []Parameter{n.outVec}
}

func (n *MulVec4Node) InVec() *Vec4Param {
	return n.inVec
}

func (n *MulVec4Node) Ratio() float32 {
	return n.ratio
}

func (n *MulVec4Node) OutVec() *Vec4Param {
	return n.outVec
}

type OutputColorNode struct {
	inColor *Vec4Param
}

func (n *OutputColorNode) InColor() *Vec4Param {
	return n.inColor
}

func (OutputColorNode) OutputParams() []Parameter {
	return nil
}
