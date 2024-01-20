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

func FrameTree() (Span, int) {
	resetRequested = true
	return rootSpan, recordedIterations
}

func BeginFrame() {
	iterations++
	currentRegion = rootRegion
	rootRegion.startTime = time.Now()
}

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

type Span struct {
	Name     string
	Children []Span
	Duration time.Duration
}

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

type Region struct {
	name      string
	parent    *Region
	children  []*Region
	startTime time.Time
	duration  time.Duration
}

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
