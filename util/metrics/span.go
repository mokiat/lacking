package metrics

import (
	"time"

	"github.com/mokiat/lacking/util/datastruct"
	"golang.org/x/exp/slices"
)

var (
	spanPool     = datastruct.NewDynamicPool[Span]()
	spans        []*Span
	spanLayer    = 0
	spanSnapshot []Span
)

func BeginSpan(name string) *Span {
	span := spanPool.Fetch()
	span.name = name
	span.startTime = time.Now()
	span.layer = spanLayer
	spans = append(spans, span)
	spanLayer++
	return span
}

func Spans() []Span {
	return spanSnapshot
}

type Span struct {
	name      string
	layer     int
	startTime time.Time
	endTime   time.Time
}

func (s *Span) Name() string {
	return s.name
}

func (s *Span) StartTime() time.Time {
	return s.startTime
}

func (s *Span) EndTime() time.Time {
	return s.endTime
}

func (s *Span) Depth() int {
	return s.layer
}

func (s *Span) End() {
	s.endTime = time.Now()
	spanLayer--
	if spanLayer == 0 {
		spanSnapshot = slices.Grow(spanSnapshot, len(spans))
		spanSnapshot = spanSnapshot[:len(spans)]
		for i, span := range spans {
			spanSnapshot[i] = *span
			spanPool.Restore(span)
		}
		spans = spans[:0]
	}
}
