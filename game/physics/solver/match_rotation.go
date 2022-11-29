package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewMatchRotation creates a new constraint solver that keeps
// two bodies oriented in the same direction on all axis.
func NewMatchRotation() physics.DBConstraintSolver {
	// TODO: Do a three-jacobian solution here
	return NewDBCombined(
		NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3()),
		NewMatchDirections().
			SetPrimaryDirection(dprec.BasisZVec3()).
			SetSecondaryDirection(dprec.BasisZVec3()),
	)
}
