package template

var renderCtx renderContext

type renderContext struct {
	node        *componentNode
	firstRender bool
	lastRender  bool
	stateIndex  int
}

func (c renderContext) isFirstRender() bool {
	return c.firstRender
}

func (c renderContext) isLastRender() bool {
	return c.lastRender
}
