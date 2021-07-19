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

// WithContext can be used during the instantiation of an Application
// in order to configure a context object.
//
// This is a helper function in place of RegisterContext. While currently not
// enforced, you should use this function during the instantiation of your
// root component.
// Using it at a later point during the lifecycle of your application could
// indicate an improper usage of contexts. You may consider using reducers
// and global state instead.
func WithContext(context interface{}) {
	RegisterContext(context)
}
