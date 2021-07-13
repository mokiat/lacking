package mat

const namespace = "github.com/mokiat/lacking/ui/mat"

// Alignment determines the positioning of child elements
// or text within a Layout or Control.
type Alignment int

const (
	AlignmentCenter Alignment = 1 + iota
	AlignmentLeft
	AlignmentRight
	AlignmentTop
	AlignmentBottom
)

// Relation determines relative to what is a position calculated.
type Relation int

const (
	RelationLeft Relation = 1 + iota
	RelationRight
	RelationTop
	RelationBottom
	RelationCenter
)
