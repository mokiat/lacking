package template

import (
	"fmt"
)

type Context struct {
	owner  *componentNode
	iScope *instanceScope
}

func (c *Context) Instance(componentType ComponentType, key string, settingsFn func()) Instance {
	c.iScope = &instanceScope{
		parent: c.iScope,
	}
	defer func() {
		c.iScope = c.iScope.parent
	}()

	c.iScope.instance = Instance{
		owner:         c.owner,
		componentType: componentType,
		key:           key,
	}
	settingsFn()
	return c.iScope.instance
}

func (c *Context) WithData(data interface{}) {
	if !isDataEqual(data, data) {
		panic(fmt.Errorf("cannot use non-comparable data type"))
	}
	c.iScope.instance.data = data
}

func (c *Context) WithLayoutData(layoutData interface{}) {
	if !isLayoutDataEqual(layoutData, layoutData) {
		panic(fmt.Errorf("cannot use non-comparable layout data type"))
	}
	c.iScope.instance.layoutData = layoutData
}

func (c *Context) WithCallbackData(callbackData interface{}) {
	c.iScope.instance.callbackData = callbackData
}

func (c *Context) WithChild(instance Instance) {
	c.iScope.instance.children = append(c.iScope.instance.children, instance)
}

func (c *Context) WithChildren(instances []Instance) {
	c.iScope.instance.children = append(c.iScope.instance.children, instances...)
}

type instanceScope struct {
	parent   *instanceScope
	instance Instance
}

func isDataEqual(a, b interface{}) bool {
	return a == b
}

func isLayoutDataEqual(a, b interface{}) bool {
	return a == b
}
