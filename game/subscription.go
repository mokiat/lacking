package game

import "github.com/mokiat/gog/ds"

type UpdateCallback func(engine *Engine, scene *Scene, elapsedSeconds float64)

type UpdateSubscription struct {
	list     *ds.List[*UpdateSubscription]
	callback UpdateCallback
}

func (s *UpdateSubscription) Delete() {
	if s.list != nil {
		s.list.Remove(s)
		s.list = nil
	}
}
