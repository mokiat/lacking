package glsl

import "github.com/mokiat/gog/ds"

func newTranslationContext() *translationContext {
	stack := ds.NewStack[[]string](0)
	stack.Push([]string{})
	return &translationContext{
		names: make(map[string]string),
		stack: stack,
	}
}

type translationContext struct {
	names map[string]string
	stack *ds.Stack[[]string]
}

func (c *translationContext) Push() {
	c.stack.Push([]string{})
}

func (c *translationContext) Pop() {
	layer := c.stack.Pop()
	for _, name := range layer {
		delete(c.names, name)
	}
}

func (c *translationContext) RegisterIdentifier(srcName, dstName string) {
	c.names[srcName] = dstName
	layer := c.stack.Pop()
	layer = append(layer, srcName)
	c.stack.Push(layer)
}

func (c *translationContext) Identifier(srcName string) string {
	return c.names[srcName]
}
