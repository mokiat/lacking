package mat

import "github.com/mokiat/lacking/ui"

// Layout represents an algorithm through which Elements are
// positioned on the screen relative to their parents.
type Layout interface {

	// Apply applies this layout to the specified Element.
	Apply(element *ui.Element)
}
