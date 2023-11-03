package component

import (
	"fmt"
	"image"
	"reflect"

	"github.com/mokiat/lacking/ui"
	"golang.org/x/image/font/opentype"
)

// RootScope initializes a new scope associated with the specified window.
//
// One would usually use this method to acquire a root scope to be later
// used in Initialize to bootstrap the framework.
func RootScope(window *ui.Window) Scope {
	return ContextScope(nil, window.Context())
}

// Scope represents a component sub-hierarchy region.
type Scope interface {

	// Context returns the ui.Context that is applicable for this component scope.
	Context() *ui.Context

	// Value returns the stored arbitrary value for the specified arbitrary key.
	// This is a mechanism through which external frameworks can attach metadata
	// to scopes.
	Value(key any) any
}

// GetScopeValue is a helper function that retrieves the value as the specified
// generic param type from the specified scope using the provided key.
//
// If there is no value with the specified key in the Scope or if the value
// is not of the correct type then the zero value for that type is returned.
func GetScopeValue[T any](scope Scope, key any) T {
	value, ok := scope.Value(key).(T)
	if !ok {
		var zeroValue T
		return zeroValue
	}
	return value
}

// TypedValueScope returns a ValueScope that uses the value's type as the
// key.
func TypedValueScope[T any](parent Scope, value T) Scope {
	return ValueScope(parent, reflect.TypeOf(value), value)
}

// TypedValue returns the value in the specified scope associated with the
// generic type.
//
// If there is no value with the specified type in the Scope then the zero
// value for that type is returned.
func TypedValue[T any](scope Scope) T {
	var zeroValue T
	value, ok := scope.Value(reflect.TypeOf(zeroValue)).(T)
	if !ok {
		return zeroValue
	}
	return value
}

// ValueScope creates a new Scope that extends the specified parent scope
// by adding the specified key-value pair.
func ValueScope(parent Scope, key, value any) Scope {
	return &valueScope{
		parent: parent,
		key:    key,
		value:  value,
	}
}

type valueScope struct {
	parent Scope
	key    any
	value  any
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
	return &contextScopedComponent{
		Component: delegate,
		contexts:  make(map[Renderable]*ui.Context),
	}
}

type contextScopedComponent struct {
	Component
	contexts map[Renderable]*ui.Context
}

func (c *contextScopedComponent) TypeName() string {
	return fmt.Sprintf("context-scoped(%s)", c.Component.TypeName())
}

func (c *contextScopedComponent) Allocate(scope Scope, invalidate InvalidateFunc) Renderable {
	context := scope.Context().CreateContext()
	scope = ContextScope(scope, context)
	ref := c.Component.Allocate(scope, invalidate)
	c.contexts[ref] = context
	return ref
}

func (c *contextScopedComponent) NotifyDelete(ref Renderable) {
	context := c.contexts[ref]
	delete(c.contexts, ref)
	context.Destroy()
	c.Component.NotifyDelete(ref)
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
		logger.Error("Error opening named image (%q): %v!", uri, err)
		return nil // TODO: Return no-op value.
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
		logger.Error("Error creating ad-hoc image: %v!", err)
		return nil // TODO: Return no-op value.
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
		logger.Error("Error opening named font (%q): %v!", uri, err)
		return nil // TODO: Return no-op value.
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
		logger.Error("Error creating ad-hoc font: %v!", err)
		return nil // TODO: Return no-op value.
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
		logger.Error("Error opening named font collection (%q): %v!", uri, err)
		return nil // TODO: Return no-op value.
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
		logger.Warn("Unable to find font (%q - %q)!", family, style)
		return nil // TODO: Return no-op value.
	}
	return font
}
