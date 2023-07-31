package app

// Cursor represents the visual aspect of a pointer on the screen.
type Cursor interface {

	// Destroy releases all resources allocated for this cursor.
	Destroy()
}

// CursorDefinition can be used to describe a new cursor.
type CursorDefinition struct {

	// Path specifies the location of the image resource.
	Path string

	// HotspotX specifies the X click position within the cursor image.
	HotspotX int

	// HotspotY specifies the Y click position within the cursor image.
	HotspotY int
}
