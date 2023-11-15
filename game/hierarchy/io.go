package hierarchy

// NodeSource represents an abstraction that is able to apply its transform
// to a node.
type NodeSource interface {

	// ApplyTo requests that any transform be applied to the specified node.
	ApplyTo(node *Node)
}

// NodeTarget represents an abstraction that is able to modify its transform
// based on a node's positioning.
type NodeTarget interface {

	// ApplyFrom requests that the node's transform be applied to the receiver.
	ApplyFrom(node *Node)
}
