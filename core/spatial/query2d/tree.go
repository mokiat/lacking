package query2d

// InvalidTreeItemID is an identifier that can be used by user
// code to mark an identifier as invalid. Such an identifier will
// never be returned by the library but must also never be passed to the
// library.
const InvalidTreeItemID = TreeItemID(0xFFFFFFFF)

// TreeItemID is an identifier used to control the placement of an item
// into a tree.
type TreeItemID uint32

// TreeStats represents the current state of a tree.
type TreeStats struct {

	// NodeCount is the total number of nodes in the tree.
	NodeCount uint32

	// ItemCount is the total number of items in the tree.
	ItemCount uint32

	// ItemCountPerDepth contains the number of items at each depth level.
	ItemCountPerDepth []uint32
}

// TreeVisitStats represents statistics on the last visit operation
// performed on a tree.
type TreeVisitStats struct {

	// NodeCountVisited is the number of nodes that were visited during the last
	// visit operation.
	NodeCountVisited uint32

	// NodeCountAccepted is the number of nodes that were determined relevant
	// during the last visit operation.
	NodeCountAccepted uint32

	// NodeCountRejected is the number of nodes that were determined irrelevant
	// during the last visit operation.
	NodeCountRejected uint32

	// ItemCountVisited is the number of items that were visited during the last
	// visit operation.
	ItemCountVisited uint32

	// ItemCountAccepted is the number of items that were determined relevant
	// during the last visit operation.
	ItemCountAccepted uint32

	// ItemCountRejected is the number of items that were determined irrelevant
	// during the last visit operation.
	ItemCountRejected uint32
}
