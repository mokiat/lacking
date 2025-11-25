package component

// Overlay represents a UI Element that stays on top of all existing
// elements and is first to receive events.
type Overlay interface {

	// Close closes this Overlay.
	Close()
}

// OpenOverlay opens a new Overlay that will take the appearance described
// in the specified instance.
func OpenOverlay(scope Scope, instance Instance) Overlay {
	app := TypedValue[*applicationComponent](scope)
	return app.openOverlay(instance)
}

// CloseOverlay closes the Overlay associated with the specified scope.
func CloseOverlay(scope Scope) {
	app := TypedValue[*applicationComponent](scope)
	overlay := TypedValue[*overlayHandle](scope)
	app.closeOverlay(overlay)
}

var _ Overlay = (*overlayHandle)(nil)

type overlayHandle struct {
	app      *applicationComponent
	instance Instance
}

func (o *overlayHandle) Close() {
	o.app.closeOverlay(o)
}
