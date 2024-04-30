package asset

// Sky represents the background of the scene.
type Sky struct {

	// NodeIndex is the index of the node that the sky is attached to.
	NodeIndex uint32

	// MaterialIndex is the index of the material that will be used to render the
	// sky.
	MaterialIndex uint32
}
