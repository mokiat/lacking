package metrics

import "time"

var (
	spans     = make(map[string]*Span)
	spanLayer = 0
	spanList  []*Span
)

func BeginSpan(name string) *Span {
	span, ok := spans[name]
	if !ok {
		span = &Span{
			name: name,
		}
		spans[name] = span
	}
	span.startTime = time.Now()
	span.layer = spanLayer
	spanLayer++
	return span
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
