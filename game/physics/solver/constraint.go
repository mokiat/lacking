package solver

// Constraint represents the algorithm necessary to solve a single-object
// constraint.
type Constraint interface {
	// Reset clears the internal cache state for this constraint solver.
	//
	// This is called at the start of every iteration.
	Reset(ctx Context)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its object.
	//
	// This is called multiple times per iteration.
	ApplyImpulses(ctx Context)

	// ApplyNudges is called by the physics engine to instruct the solver to
	// apply the necessary nudges to its object.
	//
	// This is called multiple times per iteration.
	ApplyNudges(ctx Context)
}

// PairConstraint represents the algorithm necessary to solve
// a double-object constraint.
type PairConstraint interface {
	// Reset clears the internal cache state for this constraint solver.
	//
	// This is called at the start of every iteration.
	Reset(ctx PairContext)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its objects.
	//
	// This is called multiple times per iteration.
	ApplyImpulses(ctx PairContext)

	// ApplyNudges is called by the physics engine to instruct the solver to
	// apply the necessary nudges to its objects.
	//
	// This is called multiple times per iteration.
	ApplyNudges(ctx PairContext)
}
