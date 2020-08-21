package world

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/resource"
)

func NewLayout(size float32, depth int) *Layout {
	root := newNode(sprec.ZeroVec3(), size)
	addNodeChildren(root, depth-1)
	return &Layout{
		root: root,
	}
}

type Layout struct {
	root        *Node
	environment *Environment
}

func (l *Layout) SetEnvironment(env *Environment) {
	l.environment = env
}

func (l *Layout) Environment() *Environment {
	return l.environment
}

func (l *Layout) CreateRenderable(matrix sprec.Mat4, radius float32, model *resource.Model) *Renderable {
	renderable := &Renderable{
		Matrix: matrix,
		Radius: radius,
		Model:  model,
	}
	l.InvalidateRenderable(renderable)
	return renderable
}

func (l *Layout) DeleteRenderable(renderable *Renderable) {
	renderable.detach()
}

func (l *Layout) InvalidateRenderable(renderable *Renderable) {
	renderable.detach()
	if node, ok := l.root.FindNode(renderable.Matrix.Translation(), renderable.Radius); ok {
		renderable.attach(node.renderableList)
	}
}

func newNode(position sprec.Vec3, size float32) *Node {
	return &Node{
		position:       position,
		size:           size,
		renderableList: &Renderable{},
	}
}

type Node struct {
	position sprec.Vec3
	size     float32
	children [8]*Node

	renderableList *Renderable
}

func (n *Node) FindNode(position sprec.Vec3, size float32) (*Node, bool) {
	return n, true // TODO
}

func (n *Node) IsVisibleFrom(camera *Camera) bool {
	return true // TODO
}

func addNodeChildren(node *Node, depth int) {
	if depth <= 0 {
		return
	}
	offsets := [8]sprec.Vec3{
		sprec.NewVec3(-1.0, 1.0, 1.0),
		sprec.NewVec3(1.0, 1.0, 1.0),
		sprec.NewVec3(-1.0, 1.0, -1.0),
		sprec.NewVec3(1.0, 1.0, -1.0),
		sprec.NewVec3(-1.0, -1.0, 1.0),
		sprec.NewVec3(1.0, -1.0, 1.0),
		sprec.NewVec3(-1.0, -1.0, -1.0),
		sprec.NewVec3(1.0, -1.0, -1.0),
	}
	for i := range node.children {
		child := newNode(
			sprec.Vec3Sum(node.position, sprec.Vec3Prod(offsets[i], node.size/4.0)),
			node.size/2.0,
		)
		addNodeChildren(child, depth-1)
		node.children[i] = child
	}
}
