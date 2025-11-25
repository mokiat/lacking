package component

import (
	"fmt"
	"image"
	"log/slog"

	"github.com/mokiat/lacking/ui"
	"golang.org/x/image/font/opentype"
)

// OpenImage uses the ui.Context from the specified scope to load image at
// the specified uri location.
//
// If loading of the image fails for some reason, this function logs an error
// and returns nil.
func OpenImage(scope Scope, uri string) *ui.Image {
	img, err := scope.Context().OpenImage(uri)
	if err != nil {
		logger.Error("Error opening named image",
			slog.String("uri", uri),
			slog.String("error", err.Error()),
		)
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
		logger.Error("Error creating ad-hoc image",
			slog.String("error", err.Error()),
		)
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
		logger.Error("Error opening named font",
			slog.String("uri", uri),
			slog.String("error", err.Error()),
		)
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
		logger.Error("Error creating ad-hoc font",
			slog.String("error", err.Error()),
		)
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
		logger.Error("Error opening named font collection",
			slog.String("uri", uri),
			slog.String("error", err.Error()),
		)
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
		logger.Warn("Unable to find font",
			slog.String("family", family),
			slog.String("style", style),
		)
		return nil // TODO: Return no-op value.
	}
	return font
}

// OpenSound uses the ui.Context from the specified scope to load the sound at
// the specified uri location.
//
// If the sound cannot be loaded for some reason, this function logs an error
// and returns nil.
func OpenSound(scope Scope, uri string) *ui.Sound {
	sound, err := scope.Context().OpenSound(uri)
	if err != nil {
		logger.Error("Error opening named sound",
			slog.String("uri", uri),
			slog.String("error", err.Error()),
		)
		return nil // TODO: Return no-op value.
	}
	return sound
}

// ContextScope returns a new Scope that extends the specified parent scope
// but uses a different ui.Context. This can be used to have sections of the
// UI use a dedicated resource set.
//
// NOTE: The returned scope is NOT responsible for managing the lifecycle of
// the provided ui.Context. The caller is responsible for ensuring that the
// context remains valid for the duration of the scope's usage and is properly
// released afterwards.
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

func (s *contextScope) Value(key any) any {
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

func (c *contextScopedComponent) HandleCreate(ref Renderable, scope Scope, properties Properties) {
	context := scope.Context().CreateContext()
	c.contexts[ref] = context
	scope = ContextScope(scope, context)
	c.Component.HandleCreate(ref, scope, properties)
}

func (c *contextScopedComponent) HandleUpdate(ref Renderable, scope Scope, properties Properties) {
	context := c.contexts[ref]
	scope = ContextScope(scope, context)
	c.Component.HandleUpdate(ref, scope, properties)
}

func (c *contextScopedComponent) HandleDelete(ref Renderable) {
	c.Component.HandleDelete(ref)
	context := c.contexts[ref]
	delete(c.contexts, ref)
	context.Destroy()
}
