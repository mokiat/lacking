package hierarchy

// EachNode traverses a node tree in DFS fashion, starting from the root.
func EachNode(root *Node, cb func(*Node)) {
	cb(root)
	for child := root.firstChild; child != nil; child = child.rightSibling {
		EachNode(child, cb)
	}
}
