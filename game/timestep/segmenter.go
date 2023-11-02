package timestep

import "time"

// InterpCallback is called whenever the timestep has overstepped the actual
// time and linear interpolation needs to be performed.
type InterpCallback func(alpha float64)

// NewSegmenter creates a new Segmenter that produces the specified
// fixed-interval updates.
func NewSegmenter(interval time.Duration) *Segmenter {
	return &Segmenter{
		interval:         interval,
		accumulatedDelta: 0,
	}
}

// Segmenter is a mechanism to get reliable fixed-interval ticks, sincefor some
// algorithms (like physics) having consistent intervals is a necessity.
type Segmenter struct {
	interval         time.Duration
	accumulatedDelta time.Duration
}

// Update should be called with the actual delta time elapsed. The callback
// function will be called a number of times with fixed delta intervals.
func (t *Segmenter) Update(delta time.Duration, fixedCallback UpdateCallback, interpCallback InterpCallback) {
	t.accumulatedDelta += delta
	if t.accumulatedDelta > time.Second {
		t.accumulatedDelta = t.interval // too slow; ease load
	}
	for t.accumulatedDelta > 0 {
		fixedCallback(t.interval)
		t.accumulatedDelta -= t.interval
	}
	if t.accumulatedDelta < 0 {
		excessTime := t.accumulatedDelta + t.interval
		interpCallback(excessTime.Seconds() / t.interval.Seconds())
	}
}
