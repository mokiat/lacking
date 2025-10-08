package hierarchy

// EachNode traverses a node tree in DFS fashion, starting from the root.
func EachNode(root *Node, cb func(*Node)) {
	cb(root)
	for child := root.firstChild; child != nil; child = child.rightSibling {
		EachNode(child, cb)
	}
}

// SetActive controls the active state of a node and all its children.
func SetActive(root *Node, active bool) {
	type activable interface {
		SetActive(bool)
	}
	EachNode(root, func(n *Node) {
		if act, ok := n.source.(activable); ok {
			act.SetActive(active)
		}
		if act, ok := n.target.(activable); ok {
			act.SetActive(active)
		}
	})
}
