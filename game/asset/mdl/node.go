package mdl

import (
	"iter"
	"slices"

	"github.com/mokiat/gomath/dprec"
)

func NewNode(name string) *Node {
	return &Node{
		Object:      NewObject(),
		name:        name,
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
		scale:       dprec.NewVec3(1.0, 1.0, 1.0),
	}
}

type Node struct {
	*Object
	name string

	metadata Metadata

	attachments []any

	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3

	parent *Node
	nodes  []*Node
}

func (n *Node) Metadata() Metadata {
	return n.metadata
}

func (n *Node) SetMetadata(metadata Metadata) {
	n.metadata = metadata
}

func (n *Node) Attachments() []any {
	return n.attachments
}

func (n *Node) ClearAttachments() {
	n.attachments = nil
}

func (n *Node) AddAttachment(attachment any) {
	n.attachments = append(n.attachments, attachment)
}

func (n *Node) RemoveAttachment(attachment any) {
	n.attachments = slices.DeleteFunc(n.attachments, func(candidate any) bool {
		return candidate == attachment
	})
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) SetName(name string) {
	n.name = name
}

func (n *Node) Translation() dprec.Vec3 {
	return n.translation
}

func (n *Node) SetTranslation(translation dprec.Vec3) {
	n.translation = translation
}

func (n *Node) Rotation() dprec.Quat {
	return n.rotation
}

func (n *Node) SetRotation(rotation dprec.Quat) {
	n.rotation = rotation
}

func (n *Node) Scale() dprec.Vec3 {
	return n.scale
}

func (n *Node) SetScale(scale dprec.Vec3) {
	n.scale = scale
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) SetParent(parent *Node) {
	n.parent = parent
}

func (n *Node) Nodes() []*Node {
	return n.nodes
}

func (n *Node) AddNode(node *Node) {
	node.SetParent(n)
	n.nodes = append(n.nodes, node)
}

func (n *Node) RemoveNode(node *Node) {
	if node.Parent() == n {
		node.SetParent(nil)
	}
	n.nodes = slices.DeleteFunc(n.nodes, func(candidate *Node) bool {
		return candidate == node
	})
}

func NodeAttachmentsOfType[T any](node *Node) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, attachment := range node.attachments {
			if value, ok := attachment.(T); ok {
				if !yield(value) {
					return
				}
			}
		}
	}
}
