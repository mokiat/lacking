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
	app := scope.Value(applicationKey{}).(*windowLifecycle)
	return app.OpenOverlay(instance)
}

var _ Overlay = (*overlayHandle)(nil)

type overlayHandle struct {
	lifecycle *windowLifecycle
	instance  Instance
	key       string
}

func (o *overlayHandle) Close() {
	o.lifecycle.CloseOverlay(o)
}
