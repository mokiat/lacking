package template

import (
	"fmt"
	"reflect"
	"runtime"
)

type Component struct {
	componentType string
	componentFunc ComponentFunc
}

func Plain(fn ComponentFunc) Component {
	_, file, line, _ := runtime.Caller(1)
	return Component{
		componentType: fmt.Sprintf("%s#%d", file, line),
		componentFunc: fn,
	}
}

type ComponentFunc func(props Properties) Instance

func ShallowCached(delegate Component) Component {
	var (
		oldData        interface{}
		oldLayoutData  interface{}
		oldChildren    []Instance
		cachedInstance Instance
	)

	_, file, line, _ := runtime.Caller(1)
	return Component{
		componentType: fmt.Sprintf("%s#%d", file, line),
		componentFunc: func(props Properties) Instance {
			shouldCallDelegate := renderCtx.lastRender ||
				((oldData == nil) && (oldLayoutData == nil) && (oldChildren == nil)) ||
				!isDataShallowEqual(oldData, props.data) ||
				!isLayoutDataShallowEqual(oldLayoutData, props.layoutData) ||
				!areChildrenEqual(oldChildren, props.children)
			if !shouldCallDelegate {
				return cachedInstance
			}

			oldData = props.data
			oldLayoutData = props.layoutData
			oldChildren = props.children
			cachedInstance = delegate.componentFunc(props)
			return cachedInstance
		},
	}
}

func DeepCached(delegate Component) Component {
	var (
		oldData        interface{}
		oldLayoutData  interface{}
		oldChildren    []Instance
		cachedInstance Instance
	)

	_, file, line, _ := runtime.Caller(1)
	return Component{
		componentType: fmt.Sprintf("%s#%d", file, line),
		componentFunc: func(props Properties) Instance {
			shouldCallDelegate := renderCtx.lastRender ||
				((oldData == nil) && (oldLayoutData == nil) && (oldChildren == nil)) ||
				!isDataDeepEqual(oldData, props.data) ||
				!isLayoutDataDeepEqual(oldLayoutData, props.layoutData) ||
				!areChildrenEqual(oldChildren, props.children)
			if !shouldCallDelegate {
				return cachedInstance
			}

			oldData = props.data
			oldLayoutData = props.layoutData
			oldChildren = props.children
			cachedInstance = delegate.componentFunc(props)
			return cachedInstance
		},
	}
}

func isDataShallowEqual(oldData, newData interface{}) bool {
	return newData == oldData
}

func isDataDeepEqual(oldData, newData interface{}) bool {
	return reflect.DeepEqual(newData, oldData)
}

func isLayoutDataShallowEqual(oldLayoutData, newLayoutData interface{}) bool {
	return newLayoutData == oldLayoutData
}

func isLayoutDataDeepEqual(oldLayoutData, newLayoutData interface{}) bool {
	return reflect.DeepEqual(newLayoutData, oldLayoutData)
}

func areChildrenEqual(oldChildren, newChildren []Instance) bool {
	if len(newChildren) != len(oldChildren) {
		return false
	}
	for i := range newChildren {
		if newChildren[i].key != oldChildren[i].key {
			return false
		}
		if newChildren[i].componentType != oldChildren[i].componentType {
			return false
		}
	}
	return true
}
