package asset

const UnspecifiedNodeIndex = int32(-1)

type Node struct {
	Name        string
	ParentIndex int32
	Translation [3]float64
	Rotation    [4]float64
	Scale       [3]float64
}

// NOTE: When doing armatures, a Mesh would have not just a single Node
// reference but an armature object that has references to multiple
// nodes and contains inverse transforms.
