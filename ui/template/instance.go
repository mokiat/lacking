package template

import "github.com/mokiat/lacking/ui"

// Instance represents the instance of a given Component.
type Instance struct {
	key           string
	componentType string
	componentFunc ComponentFunc

	data         interface{}
	layoutData   interface{}
	callbackData interface{}
	children     []Instance

	element *ui.Element
}

func (i Instance) properties() Properties {
	return Properties{
		data:         i.data,
		layoutData:   i.layoutData,
		callbackData: i.callbackData,
		children:     i.children,
	}
}

func (i Instance) hasMatchingChild(instance Instance) bool {
	for _, child := range i.children {
		if child.key == instance.key && child.componentType == instance.componentType {
			return true
		}
	}
	return false
}
