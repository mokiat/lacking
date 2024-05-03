package mdl

import "github.com/mokiat/gomath/sprec"

type Armature struct {
	joints []*Joint
}

func (a *Armature) Joints() []*Joint {
	return a.joints
}

func (a *Armature) AddJoint(joint *Joint) {
	a.joints = append(a.joints, joint)
}

type Joint struct {
	node              Node
	inverseBindMatrix sprec.Mat4
}

func (j *Joint) Node() Node {
	return j.node
}

func (j *Joint) SetNode(node Node) {
	j.node = node
}

func (j *Joint) InverseBindMatrix() sprec.Mat4 {
	return j.inverseBindMatrix
}

func (j *Joint) SetInverseBindMatrix(matrix sprec.Mat4) {
	j.inverseBindMatrix = matrix
}
