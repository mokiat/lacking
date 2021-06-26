package app

// Cursor represents the visual aspect of a pointer on the screen.
type Cursor interface {

	// Delete releases all resources allocated for this cursor.
	Delete()
}

// CursorDefinition can be used to describe a new cursor.
type CursorDefinition struct {
	Path     string
	HotspotX int
	HotspotY int
}
