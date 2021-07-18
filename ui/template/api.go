package template

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/mokiat/lacking/ui"
)

func New(component Component, setupFn func()) Instance {
	dslCtx = &dslContext{
		parent: dslCtx,
	}
	defer func() {
		dslCtx = dslCtx.parent
	}()

	dslCtx.instance = Instance{
		componentType: component.componentType,
		componentFunc: component.componentFunc,
	}
	setupFn()
	return dslCtx.instance
}

func WithData(data interface{}) {
	dslCtx.instance.data = data
}

func WithLayoutData(layoutData interface{}) {
	dslCtx.instance.layoutData = layoutData
}

func WithCallbackData(callbackData interface{}) {
	dslCtx.instance.callbackData = callbackData
}

func WithChild(key string, instance Instance) {
	instance.key = key
	dslCtx.instance.children = append(dslCtx.instance.children, instance)
}

func WithChildren(children []Instance) {
	dslCtx.instance.children = children
}

func Once(fn func()) {
	if renderCtx.isFirstRender() {
		fn()
	}
}

func Defer(fn func()) {
	if renderCtx.isLastRender() {
		fn()
	}
}

func UseState(fn func() interface{}) *State {
	if renderCtx.firstRender {
		renderCtx.node.states[renderCtx.stateDepth] = append(renderCtx.node.states[renderCtx.stateDepth], State{
			node:  renderCtx.node,
			value: fn(),
		})
	}
	log.Println("use_state", renderCtx.stateDepth, renderCtx.stateIndex)
	result := &renderCtx.node.states[renderCtx.stateDepth][renderCtx.stateIndex]
	renderCtx.stateIndex++
	return result
}

func Window() *ui.Window {
	return uiCtx.Window()
}

func OpenImage(uri string) ui.Image {
	img, err := uiCtx.OpenImage(uri)
	if err != nil {
		panic(fmt.Errorf("failed to open image %q: %w", uri, err))
	}
	return img
}

func OpenFontCollection(uri string) {
	if _, err := uiCtx.OpenFontCollection(uri); err != nil {
		panic(fmt.Errorf("failed to open font collection %q: %w", uri, err))
	}
}

func GetFont(family, style string) ui.Font {
	font, found := uiCtx.GetFont(family, style)
	if !found {
		panic(fmt.Errorf("could not find font %q / %q", family, style))
	}
	return font
}

func InitGlobalState(state *ReducedState) {
	rootState = state
}

func NewReducedState(reducer Reducer) *ReducedState {
	result := &ReducedState{
		reducer: reducer,
		value:   reducer(nil, nil),
	}
	globalStates = append(globalStates, result)
	return result
}

func Dispatch(action interface{}) {
	invalidateGlobalNodes := false
	for _, state := range globalStates {
		newValue := state.reducer(state, action)
		if newValue != state.value {
			state.value = newValue
			invalidateGlobalNodes = true
		}
	}
	if invalidateGlobalNodes {
		for _, node := range globalStateNodes {
			node.reconcile(node.instance)
		}
	}
}

var globalStateNodes []*componentNode

type Reducer func(state *ReducedState, action interface{}) interface{}

type ConnectFunc func(props Properties, rootState *ReducedState) (data interface{}, callbackData interface{})

func Connect(delegate Component, connectFn ConnectFunc) Component {
	_, file, line, _ := runtime.Caller(1)
	return Component{
		componentType: fmt.Sprintf("%s#%d", file, line),
		componentFunc: func(props Properties) Instance {
			Once(func() {
				globalStateNodes = append(globalStateNodes, renderCtx.node)
			})

			Defer(func() {
				for i, node := range globalStateNodes {
					if node == renderCtx.node {
						globalStateNodes[i] = globalStateNodes[len(globalStateNodes)-1]
						globalStateNodes = globalStateNodes[:len(globalStateNodes)-1]
					}
				}
			})

			data, callbackData := connectFn(props, rootState)
			return delegate.componentFunc(Properties{
				data:         data,
				layoutData:   props.layoutData,
				callbackData: callbackData,
				children:     props.children,
			})
		},
	}
}

func After(duration time.Duration, fn func()) {
	node := renderCtx.node
	time.AfterFunc(duration, func() {
		uiCtx.Schedule(func() {
			if node.isValid() {
				fn()
			}
		})
	})
}
