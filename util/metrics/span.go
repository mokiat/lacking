package metrics

import "time"

var (
	spans     = make(map[string]*Span)
	spanLayer = 0
	spanList  []*Span
)

func BeginSpan(name string) *Span {
	if span, ok := spans[name]; ok {
		span.startTime = time.Now()
		return span
	}
	result := &Span{
		name:      name,
		layer:     spanLayer,
		startTime: time.Now(),
	}
	spanLayer++
	spans[name] = result
	return result
}

func Spans() []*Span {
	if len(spans) == len(spanList) {
		return spanList
	}
	spanList = make([]*Span, 0, len(spans))
	for _, span := range spans {
		spanList = append(spanList, span)
	}
	return spanList
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
}
