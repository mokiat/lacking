package shading

type ShaderFunc[P any] func(palette P)

type GeometryVertexFunc ShaderFunc[GeometryVertexPalette]
type GeometryFragmentFunc ShaderFunc[GeometryFragmentPalette]

type ShadowVertexFunc ShaderFunc[ShadowVertexPalette]
type ShadowFragmentFunc ShaderFunc[ShadowFragmentPalette]

type ForwardVertexFunc ShaderFunc[ForwardVertexPalette]
type ForwardFragmentFunc func(palette *ForwardFragmentPalette)

type Palette struct {
	nodes []Node
}

func (p *Palette) Nodes() []Node {
	return p.nodes
}

func (p *Palette) ConstVec4(x, y, z, w float32) *Vec4Param {
	node := &ConstVec4Node{
		outVec: &Vec4Param{},
		x:      x,
		y:      y,
		z:      z,
		w:      w,
	}
	p.nodes = append(p.nodes, node)
	return node.outVec
}

func (p *Palette) MulVec4(inVec *Vec4Param, ratio float32) *Vec4Param {
	inVec.MarkUsed()
	node := &MulVec4Node{
		inVec:  inVec,
		ratio:  ratio,
		outVec: &Vec4Param{},
	}
	p.nodes = append(p.nodes, node)
	return node.outVec
}

type GeometryVertexPalette struct {
	Palette
}

type GeometryFragmentPalette struct {
	Palette
}

type ShadowVertexPalette struct {
	Palette
}

type ShadowFragmentPalette struct {
	Palette
}

type ForwardVertexPalette struct {
	Palette
}

func NewForwardFragmentPalette() *ForwardFragmentPalette {
	return &ForwardFragmentPalette{}
}

type ForwardFragmentPalette struct {
	Palette
}

func (p *ForwardFragmentPalette) OutputColor(inColor *Vec4Param) {
	inColor.MarkUsed()
	p.nodes = append(p.nodes, &OutputColorNode{
		inColor: inColor,
	})
}
