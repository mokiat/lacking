package game

import "github.com/mokiat/lacking/util/datastruct"

type UpdateCallback func(engine *Engine, scene *Scene)

type UpdateSubscription struct {
	list     *datastruct.List[*UpdateSubscription]
	callback UpdateCallback
}

func (s *UpdateSubscription) Delete() {
	if s.list != nil {
		s.list.Remove(s)
		s.list = nil
	}
}
