package metric

import (
	"slices"
	"time"
)

var (
	rootSpan           Span
	recordedIterations int

	rootRegion *Region
	iterations int

	currentRegion  *Region
	resetRequested bool
)

func init() {
	rootSpan = Span{
		Name: "frame",
	}
	rootRegion = &Region{
		name: "frame",
	}
	currentRegion = rootRegion
}

// FlameTree returns the current flame graph tree and the number of
// recorded iterations.
func FrameTree() (Span, int) {
	resetRequested = true
	return rootSpan, recordedIterations
}

// BeginFrame starts a new frame. This should be called at the beginning of
// each game frame.
func BeginFrame() {
	iterations++
	currentRegion = rootRegion
	rootRegion.startTime = time.Now()
}

// EndFrame ends the current frame.
func EndFrame() {
	rootRegion.duration += time.Since(rootRegion.startTime)
	updateSpan(&rootSpan, rootRegion)
	recordedIterations = iterations
	if resetRequested {
		iterations = 0
		resetRegion(rootRegion)
		resetRequested = false
	}
}

// Span represents a span in a flame graph. Unlike a region, it represents
// a summary of the overall time spent in a region.
type Span struct {
	Name     string
	Children []Span
	Duration time.Duration
}

// BeginRegion starts a new monitoring region. The region must be ended with
// a call to End.
func BeginRegion(name string) *Region {
	index := slices.IndexFunc(currentRegion.children, func(candidate *Region) bool {
		return candidate.name == name
	})
	var region *Region
	if index >= 0 {
		region = currentRegion.children[index]
	} else {
		region = &Region{
			name:   name,
			parent: currentRegion,
		}
		currentRegion.children = append(currentRegion.children, region)
	}
	region.startTime = time.Now()
	currentRegion = region
	return region
}

// Region represents a monitoring region and is using during profiling.
type Region struct {
	name      string
	parent    *Region
	children  []*Region
	startTime time.Time
	duration  time.Duration
}

// End ends the current region.
func (r *Region) End() {
	r.duration += time.Since(r.startTime)
	currentRegion = r.parent
}

func updateSpan(span *Span, region *Region) {
	span.Name = region.name
	span.Duration = region.duration
	if missing := len(region.children) - len(span.Children); missing > 0 {
		span.Children = append(span.Children, make([]Span, missing)...)
	}
	for i := range region.children {
		updateSpan(&span.Children[i], region.children[i])
	}
}

func resetRegion(region *Region) {
	region.duration = 0
	for i := range region.children {
		resetRegion(region.children[i])
	}
}
