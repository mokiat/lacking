package asset

import "github.com/mokiat/gomath/sprec"

// Armature represents the definition of a skeleton.
type Armature struct {

	// Joints is the collection of joints that make up the armature.
	Joints []Joint
}

// Joint represents a single joint in an armature.
type Joint struct {

	// NodeIndex is the index of the node that is associated with the joint.
	NodeIndex uint32

	// InverseBindMatrix is the matrix that transforms the joint from its
	// local space to the space of the mesh.
	InverseBindMatrix sprec.Mat4
}
