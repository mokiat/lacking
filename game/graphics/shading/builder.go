package shading

// TODO: MeshFunc (used for vertex shader)

type ShadowFunc func(palette *ShadowPalette)

type GeometryFunc func(palette *GeometryPalette)

type EmissiveFunc func(palette *EmissivePalette)

type ForwardFunc func(palette *ForwardPalette)

type LightingFunc func(palette *LightingPalette)

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

type ShadowPalette struct {
	Palette
}

type GeometryPalette struct {
	Palette
}

type EmissivePalette struct {
	Palette
}

func NewForwardPalette() *ForwardPalette {
	return &ForwardPalette{}
}

type ForwardPalette struct {
	Palette
}

func (p *ForwardPalette) OutputColor(inColor *Vec4Param) {
	inColor.MarkUsed()
	p.nodes = append(p.nodes, &OutputColorNode{
		inColor: inColor,
	})
}

type LightingPalette struct {
	Palette
}
