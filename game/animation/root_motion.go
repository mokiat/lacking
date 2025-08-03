package animation

import "github.com/mokiat/gomath/dprec"

// NewRootMotion creates a new RootMotion instance that is controlled by the
// specified root bone.
func NewRootMotion(node Node, bone string) *RootMotion {
	return &RootMotion{
		node: node,
		bone: bone,
	}
}

// RootMotion allows for the extraction of relative animation transformations
// from a root bone in order to move a character.
type RootMotion struct {
	node Node
	bone string
}

// DeltaTransform returns the transformation that occurred during the last
// animation tick for the root bone.
func (m *RootMotion) DeltaTransform() dprec.Mat4 {
	deltaTransform := m.node.BoneTransformDelta(m.bone)
	return dprec.TRSMat4(
		deltaTransform.Translation.ValueOrDefault(dprec.ZeroVec3()),
		deltaTransform.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		deltaTransform.Scale.ValueOrDefault(dprec.NewVec3(1.0, 1.0, 1.0)),
	)
}
