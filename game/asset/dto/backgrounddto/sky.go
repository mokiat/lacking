package backgrounddto

// Sky represents the background of the scene.
type Sky struct {

	// NodeIndex is the index of the node that the sky is attached to.
	NodeIndex uint32

	// MaterialID is the ID of the material that will be used to render the
	// sky.
	MaterialID uint32
}
