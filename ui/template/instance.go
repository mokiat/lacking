package template

import "github.com/mokiat/lacking/ui"

// Instance represents the instance of a given Component.
type Instance struct {
	owner         *componentNode
	componentType ComponentType

	key          string
	data         interface{}
	layoutData   interface{}
	callbackData interface{}

	element  *ui.Element
	children []Instance
}
