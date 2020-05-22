package graphics

func newPipeline() *Pipeline {
	items := make([]Item, 1024)
	for i := range items {
		items[i] = createItem()
	}
	sequences := make([]Sequence, 32)
	for i := range sequences {
		sequences[i] = createSequence(items)
	}
	return &Pipeline{
		preRenderActions:  make([]func(), 0, 32),
		postRenderActions: make([]func(), 0, 32),
		items:             items,
		sequences:         sequences[:0],
	}
}

type Pipeline struct {
	preRenderActions  []func()
	postRenderActions []func()

	items     []Item
	itemIndex int

	sequences      []Sequence
	activeSequence *Sequence
}

func (p *Pipeline) BeginSequence() *Sequence {
	if p.activeSequence != nil {
		panic("previous sequence has not been ended")
	}

	sequencesLen := len(p.sequences)
	if sequencesLen == cap(p.sequences) {
		panic("max number of render sequences reached")
	}

	p.sequences = p.sequences[:sequencesLen+1]
	p.activeSequence = &p.sequences[sequencesLen]
	p.activeSequence.reset(p.itemIndex)
	return p.activeSequence
}

func (p *Pipeline) EndSequence(sequence *Sequence) {
	p.itemIndex = sequence.itemEndIndex
	p.activeSequence = nil
}

func (p *Pipeline) SchedulePreRender(action func()) {
	actionsLen := len(p.preRenderActions)
	if actionsLen == cap(p.preRenderActions) {
		panic("maximum number of pre-render actions reached")
	}
	p.preRenderActions = p.preRenderActions[:actionsLen+1]
	p.preRenderActions[actionsLen] = action
}

func (p *Pipeline) SchedulePostRender(action func()) {
	actionsLen := len(p.postRenderActions)
	if actionsLen == cap(p.postRenderActions) {
		panic("maximum number of post-render actions reached")
	}
	p.postRenderActions = p.postRenderActions[:actionsLen+1]
	p.postRenderActions[actionsLen] = action
}

func (p *Pipeline) rewind() {
	p.itemIndex = 0
	p.preRenderActions = p.preRenderActions[:0]
	p.postRenderActions = p.postRenderActions[:0]
	p.sequences = p.sequences[:0]
	p.activeSequence = nil
}

func (p *Pipeline) sequencesView() []Sequence {
	return p.sequences
}

func (p *Pipeline) preRenderActionsView() []func() {
	return p.preRenderActions
}

func (p *Pipeline) postRenderActionsView() []func() {
	return p.postRenderActions
}
