package component

import (
	"image"

	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui"
	"golang.org/x/image/font/opentype"
)

// RootScope returns the main scope for the UI hierarchy. Resources loaded from
// the ui.Context of this scope will not be released until the UI has completed.
func RootScope() Scope {
	return rootScope
}

// Scope represents a component sub-hierarchy region.
type Scope interface {

	// Context returns the ui.Context that is applicable for this component scope.
	Context() *ui.Context

	// Value returns the stored arbitrary value for the specified arbitrary key.
	// This is a mechanism through which external frameworks can attach metadata
	// to scopes.
	Value(key interface{}) interface{}
}

// GetScopeValue is a helper function that retrieves the value as the specified
// generic param type from the specified scope using the provided key.
//
// If there is no value with the specified key in the Scope or if the value
// is not of the correct type then nil is returned.
func GetScopeValue[T any](scope Scope, key interface{}) T {
	value, ok := scope.Value(key).(T)
	if !ok {
		var defaultValue T
		return defaultValue
	}
	return value
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
	ctxSet := make(map[*componentNode]*ui.Context)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties, scope Scope) Instance {
			ctx := ctxSet[renderCtx.node]
			if renderCtx.isFirstRender() {
				ctx = scope.Context().CreateContext()
				ctxSet[renderCtx.node] = ctx
			}
			if renderCtx.isLastRender() {
				defer func() {
					ctx.Destroy()
					delete(ctxSet, renderCtx.node)
				}()
			}
			scope = ContextScope(scope, ctx)
			renderCtx.node.scope = scope
			return delegate.componentFunc(props, scope)
		},
	}
}

// Window uses the specified scope to retrieve the Window that owns that
// particular scope.
func Window(scope Scope) *ui.Window {
	return scope.Context().Window()
}

// OpenImage uses the ui.Context from the specified scope to load image at
// the specified uri location.
//
// If loading of the image fails for some reason, this function logs an error
// and returns nil.
func OpenImage(scope Scope, uri string) *ui.Image {
	img, err := scope.Context().OpenImage(uri)
	if err != nil {
		log.Error("[component] Error opening image (%q): %v", uri, err)
		return nil
	}
	return img
}

// CreateImage uses the ui.Context from the specified scope to create a new
// image.
//
// If the image could not be created for some reason, this function logs an
// error and returns nil.
func CreateImage(scope Scope, img image.Image) *ui.Image {
	result, err := scope.Context().CreateImage(img)
	if err != nil {
		log.Error("[component] Error creating image: %v", err)
		return nil
	}
	return result
}

// OpenFont uses the ui.Context from the specified scope to load the font at
// the specified uri location.
//
// If the font cannot be loaded for some reason, this function logs an error
// and returns nil.
func OpenFont(scope Scope, uri string) *ui.Font {
	font, err := scope.Context().OpenFont(uri)
	if err != nil {
		log.Error("[component] Error opening font (%q): %v", uri, err)
		return nil
	}
	return font
}

// CreateFont uses the ui.Context from the specified scope to create a new
// ui.Font based on the passed opentype Font.
//
// If creation of the font fails, this function logs an error and returns nil.
func CreateFont(scope Scope, otFont *opentype.Font) *ui.Font {
	font, err := scope.Context().CreateFont(otFont)
	if err != nil {
		log.Error("[component] Error creating font: %v", err)
		return nil
	}
	return font
}

// OpenFontCollection uses the ui.Context from the specified scope to load the
// font collection at the specified uri location.
//
// If the collection cannot be loaded for some reason, this function logs an
// error and returns nil.
func OpenFontCollection(scope Scope, uri string) *ui.FontCollection {
	collection, err := scope.Context().OpenFontCollection(uri)
	if err != nil {
		log.Error("[component] Error opening font collection (%q): %v", uri, err)
		return nil
	}
	return collection
}

// GetFont uses the ui.Context from the specified scope to retrieve the font
// with the specified family and style.
//
// This function returns nil and logs a warning if it is unable to find the
// requested font (fonts need to have been loaded beforehand through one
// of the other Font functions).
func GetFont(scope Scope, family, style string) *ui.Font {
	font, found := scope.Context().GetFont(family, style)
	if !found {
		log.Warn("[component] Unable to find font (%q / %q)", family, style)
		return nil
	}
	return font
}
