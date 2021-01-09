package ui

// Control represents an abstract user interface
// control. Most methods on this interface should be
// safe for use from end users.
// Make sure you call the methods only from the UI
// thread as otherwise the behavior is unpredictable.
type Control interface {

	// Element returns the hierarchy element that manages
	// this control. This method is intended for use by
	// internal logic and control implementations.
	// End users should not need to use this.
	Element() *Element

	// ID returns the unique identifier for the control
	// if such has been specified.
	ID() string

	// AddStyle adds a new style to this control.
	// If the style has already been added, then this
	// operation does nothing.
	AddStyle(name string)

	// RemoveStyle removes a style from this control.
	// If the style is not applied to the control, then
	// this method does nothing.
	RemoveStyle(name string)

	// Styles returns a slice of all styles applied to
	// this control.
	// Styles are a mechanism through which configurations
	// and layout data can be shared across controls.
	// They are applied in the order in which they were
	// added to the control, where latter styles override
	// former ones.
	Styles() []string

	// SetLayoutData changes the layout configuration for
	// this control.
	SetLayoutData(data LayoutData)

	// LayoutData returns the layout configuration for
	// this control.
	LayoutData() LayoutData

	// Parent returns the parent control that holds this
	// control.
	// Note that the returned control is not necessary the
	// immediate parent element that references this control.
	// It is possible for a parent control to manage intermediate
	// elements (e.g. rows, headers) that hold the control.
	// If this is the top-most control, nil is returned.
	Parent() Control

	// SetVisible controls whether the control should be
	// displayed.
	// Setting this to false does not cause controls to
	// be repositioned and instead it just prevents the
	// control from being rendered and from receiving
	// input events.
	SetVisible(visible bool)

	// IsVisible returns whether this control should be
	// rendered and receive events.
	IsVisible() bool

	// SetMaterialized controls whether the control is at
	// all considered as existing.
	// Setting this to false causes the control to behave
	// as though it has been deleted.
	SetMaterialized(materialized bool)

	// IsMaterialized returns whether this control is
	// considered as existing or not.
	IsMaterialized() bool

	// SetEnabled controls whether this control receives
	// input events.
	// Setting this to false means that the control
	// would not react to events like mouse or keyboard
	// inputs and depending on the implementation might
	// be rendered in a different way to indicate that
	// it is not enabled.
	SetEnabled(enabled bool)

	// IsEnabled returns whether the control is to
	// receive input events.
	IsEnabled() bool

	// Delete removes this control. It can no longer and
	// should no longer be used in any way.
	// If the control is already deleted, this method
	// does nothing.
	Delete()
}
