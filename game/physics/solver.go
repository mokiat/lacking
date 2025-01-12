package physics

var (
	// ImpulseIterationCount controls the number of iterations for impulse
	// solutions by the solvers.
	ImpulseIterationCount = 8

	// NudgeIterationCount controls the number of iterations for nudge
	// solutions by the solvers.
	NudgeIterationCount = 8

	// ImpulseDriftAdjustmentRatio controls the amount by which impulses should
	// try to correct positional drift.
	//
	// This is the `beta` coefficient in the Baumgarte stabilization approach.
	ImpulseDriftAdjustmentRatio = 0.2

	// NudgeDriftAdjustmentRatio controls the amount by which nudges should
	// try to correct positional drift.
	//
	// The value here is accumulated over all iterations. In fact, the total
	// remaining error is proportional to (1.0 - ratio) ^ iterations.
	//
	// Some error should be left in order to avoid jitters due to imprecise
	// integration of the correction and to leave some drift for the
	// impulse solution.
	NudgeDriftAdjustmentRatio = 0.2
)
