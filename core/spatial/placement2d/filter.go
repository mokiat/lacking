package placement2d

import "github.com/mokiat/gog/opt"

// Filter represents a set of criteria to filter 2D shapes in a scene.
type Filter struct {

	// Mask is a bitmask used to filter shapes based on their assigned layers.
	Mask opt.T[uint32]

	// SkipDynamic indicates whether dynamic shapes should be excluded from the
	// results.
	SkipDynamic bool

	// SkipStatic indicates whether static shapes should be excluded from the
	// results.
	SkipStatic bool
}
