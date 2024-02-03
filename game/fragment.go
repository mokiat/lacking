package game

// FragmentDefinition describes a fragment of a game scene.
type FragmentDefinition struct {
	Nodes []FragmentNodeDefinition
}

// FragmentNodeDefinition describes a node within a fragment of a game scene.
type FragmentNodeDefinition struct {
	ParentIndex   int
	IsStationary  bool
	IsInseparable bool
}
