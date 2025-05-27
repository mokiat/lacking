package backgrounddto

// Sky represents the background of the scene.
type Sky struct {

	// NodeID is the ID of the node that the sky is attached to.
	NodeID uint32

	// MaterialID is the ID of the material that will be used to render the
	// sky.
	MaterialID uint32
}
