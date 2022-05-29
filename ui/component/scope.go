package component

import (
	"github.com/mokiat/lacking/ui"
)

// Scope represents a component sub-hierarchy region.
type Scope interface {

	// Context returns the ui.Context that is applicable for this component scope.
	Context() *ui.Context

	// Value returns the stored arbitrary value for the specified arbitrary key.
	// This is a mechanism through which external frameworks can attach metadata
	// to scopes.
	Value(key interface{}) interface{}
}

// ValueScope creates a new Scope that extends the specified parent scope
// by adding the specified key-value pair.
func ValueScope(parent Scope, key, value interface{}) Scope {
	return &valueScope{
		parent: parent,
		key:    key,
		value:  value,
	}
}

type valueScope struct {
	parent Scope
	key    interface{}
	value  interface{}
}

func (s *valueScope) Context() *ui.Context {
	if s.parent == nil {
		return nil
	}
	return s.parent.Context()
}

func (s *valueScope) Value(key interface{}) interface{} {
	if s.key == key {
		return s.value
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.Value(key)
}

// ContextScope returns a new Scope that extends the specified parent scope
// but uses a different ui.Context. This can be used to have sections of the
// UI use a dedicated resource set that will be freed once the hierarchy
// is destroyed.
func ContextScope(parent Scope, ctx *ui.Context) Scope {
	return &contextScope{
		parent: parent,
		ctx:    ctx,
	}
}

type contextScope struct {
	parent Scope
	ctx    *ui.Context
}

func (s *contextScope) Context() *ui.Context {
	return s.ctx
}

func (s *contextScope) Value(key interface{}) interface{} {
	if s.parent == nil {
		return nil
	}
	return s.parent.Value(key)
}

// ContextScoped will cause a wrapped component to receive a Scope that has a
// dedicated ui.Context with a lifecycle equal to the lifecycle of the
// component instance.
func ContextScoped(delegate Component) Component {
	ctxs := make(map[*componentNode]*ui.Context)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties, scope Scope) Instance {
			ctx := ctxs[renderCtx.node]
			if renderCtx.isFirstRender() {
				ctx = scope.Context().CreateContext()
				ctxs[renderCtx.node] = ctx
			}
			if renderCtx.isLastRender() {
				defer func() {
					ctx.Destroy()
					delete(ctxs, renderCtx.node)
				}()
			}
			return delegate.componentFunc(props, ContextScope(scope, ctx))
		},
	}
}
