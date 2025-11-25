package timestep

import "time"

// NewSegmenter creates a new Segmenter that produces the specified
// fixed-interval updates.
func NewSegmenter(interval time.Duration) *Segmenter {
	return &Segmenter{
		interval:          interval,
		accumulationLimit: time.Second,
		accumulatedDelta:  0,
	}
}

// Segmenter is a mechanism to get reliable fixed-interval ticks, sincefor some
// algorithms (like physics) having consistent intervals is a necessity.
type Segmenter struct {
	interval          time.Duration
	accumulationLimit time.Duration
	accumulatedDelta  time.Duration

	stepCallback   StepCallback
	fixedCallback  UpdateCallback
	interpCallback InterpolationCallback
}

// Reset clears any accumulated time.
func (t *Segmenter) Reset() {
	t.accumulatedDelta = 0
}

// AccumulationLimit returns the limit on how much time can be accumulated.
func (t *Segmenter) AccumulationLimit() time.Duration {
	return t.accumulationLimit
}

// SetAccumulationLimit sets a limit on how much time can be accumulated.
// This is useful to avoid spiral of death scenarios when the application is
// running too slow.
func (t *Segmenter) SetAccumulationLimit(limit time.Duration) {
	t.accumulationLimit = limit
}

// StepCallback returns the callback that is called before fixed time steps are
// made to inform of the number of steps pending.
func (t *Segmenter) StepCallback() StepCallback {
	return t.stepCallback
}

// SetStepCallback sets the callback that is called before fixed time steps are
// made to inform of the number of steps pending.
func (t *Segmenter) SetStepCallback(callback StepCallback) {
	t.stepCallback = callback
}

// FixedCallback returns the callback that is called on each fixed time step.
func (t *Segmenter) FixedCallback() UpdateCallback {
	return t.fixedCallback
}

// SetFixedCallback sets the callback that is called on each fixed time step.
func (t *Segmenter) SetFixedCallback(callback UpdateCallback) {
	t.fixedCallback = callback
}

// InterpCallback returns the callback that is called whenever the timestep has
// overstepped the actual time and linear interpolation needs to be performed.
func (t *Segmenter) InterpCallback() InterpolationCallback {
	return t.interpCallback
}

// SetInterpCallback sets the callback that is called whenever the timestep has
// overstepped the actual time and linear interpolation needs to be performed.
func (t *Segmenter) SetInterpCallback(callback InterpolationCallback) {
	t.interpCallback = callback
}

// Update should be called with the actual delta time elapsed. The callback
// function will be called a number of times with fixed delta intervals.
func (t *Segmenter) Update(delta time.Duration) {
	t.accumulatedDelta += delta
	t.accumulatedDelta = min(t.accumulatedDelta, t.accumulationLimit)

	steps := t.accumulatedDelta.Seconds() / t.interval.Seconds()
	if t.stepCallback != nil {
		t.stepCallback(steps)
	}

	for t.accumulatedDelta >= t.interval {
		if t.fixedCallback != nil {
			t.fixedCallback(t.interval)
		}
		t.accumulatedDelta -= t.interval
	}

	if t.interpCallback != nil {
		fraction := t.accumulatedDelta.Seconds() / t.interval.Seconds()
		t.interpCallback(fraction)
	}
}
