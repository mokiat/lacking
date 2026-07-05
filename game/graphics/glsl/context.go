package glsl

import (
	"fmt"

	"github.com/mokiat/gog/ds"
)

func newTranslationContext() *translationContext {
	stack := ds.EmptyStack[[]string]()
	stack.Push([]string{})
	return &translationContext{
		names: make(map[string]string),
		stack: stack,
	}
}

type translationContext struct {
	names     map[string]string
	stack     *ds.Stack[[]string]
	freeIndex int
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

func (c *translationContext) CreateIdentifier(srcName string) string {
	c.freeIndex++
	dstName := fmt.Sprintf("uVar%d", c.freeIndex)
	c.RegisterIdentifier(srcName, dstName)
	return dstName
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
