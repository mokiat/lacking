package app

import "github.com/mokiat/lacking/app"

// GraphicsEngine indicates the type of graphics engine that should be
// enabled by the window.
type GraphicsEngine string

const (
	// GraphicsEngineOpenGL indicates that the window should bind the
	// OpenGL syscall functions.
	GraphicsEngineOpenGL GraphicsEngine = "gl"

	// GraphicsEngineVulkan indicates that the window should bind the
	// Vulkan syscall functions.
	GraphicsEngineVulkan GraphicsEngine = "vulkan"
)

// NewConfig creates a new Config object that contains the minimum
// required settings.
func NewConfig(title string, width, height int) *Config {
	return &Config{
		title:          title,
		width:          width,
		height:         height,
		swapInterval:   1,
		cursorVisible:  true,
		graphicsEngine: GraphicsEngineOpenGL,
	}
}

// Config represents an application window configuration.
type Config struct {
	title          string
	width          int
	height         int
	minWidth       *int
	maxWidth       *int
	minHeight      *int
	maxHeight      *int
	swapInterval   int
	maximized      bool
	cursorVisible  bool
	cursor         *app.CursorDefinition
	icon           string
	graphicsEngine GraphicsEngine
}

// SetMinSize sets a minimum size for the window.
// Specifying a non-positive value for any dimension disables this setting.
func (c *Config) SetMinSize(width, height int) {
	if width > 0 && height > 0 {
		c.minWidth = &width
		c.minHeight = &height
	} else {
		c.minWidth = nil
		c.minHeight = nil
	}
}

// MinSize returns the minimum size for the window.
// This method returns (0, 0) if a minimum size is not specified.
func (c *Config) MinSize() (int, int) {
	if c.minWidth == nil || c.maxWidth == nil {
		return 0, 0
	}
	return *c.minWidth, *c.minHeight
}

// SetMaxSize sets a maximum size for the window.
// Specifying a non-positive value for any dimension disables this setting.
func (c *Config) SetMaxSize(width, height int) {
	if width > 0 && height > 0 {
		c.maxWidth = &width
		c.maxHeight = &height
	} else {
		c.maxWidth = nil
		c.maxHeight = nil
	}
}

// MaxSize returns the maximum size for the window.
// This method returns (0, 0) if a maximum size is not specified.
func (c *Config) MaxSize() (int, int) {
	if c.maxWidth == nil || c.maxHeight == nil {
		return 0, 0
	}
	return *c.maxWidth, *c.maxHeight
}

// SetGraphicsEngine configures the desired graphics engine.
func (c *Config) SetGraphicsEngine(engine GraphicsEngine) {
	c.graphicsEngine = engine
}

// GraphicsEngine returns the graphics engine that will be
// used. By default this is GraphicsEngineOpenGL.
func (c *Config) GraphicsEngine() GraphicsEngine {
	return c.graphicsEngine
}

// SetVSync indicates whether v-sync should be enabled.
func (c *Config) SetVSync(vsync bool) {
	if vsync {
		c.swapInterval = 1
	} else {
		c.swapInterval = 0
	}
}

// VSync returns whether v-sync will be enabled.
func (c *Config) VSync() bool {
	return c.swapInterval != 0
}

// SetMaximized specifies whether the window should be
// created in maximized state.
func (c *Config) SetMaximized(maximized bool) {
	c.maximized = maximized
}

// Maximized returns whether the window will be created in
// maximized state.
func (c *Config) Maximized() bool {
	return c.maximized
}

// SetCursorVisible specifies whether the cursor should be
// displayed when moved over the window.
func (c *Config) SetCursorVisible(visible bool) {
	c.cursorVisible = visible
}

// CursorVisible returns whether the cursor will be shown
// when hovering over the window.
func (c *Config) CursorVisible() bool {
	return c.cursorVisible
}

// SetCursor configures a custom cursor to be used.
// Specifying nil disables the custom cursor.
func (c *Config) SetCursor(definition *app.CursorDefinition) {
	c.cursor = definition
}

// Cursor returns the cursor configuration for this application.
func (c *Config) Cursor() *app.CursorDefinition {
	return c.cursor
}

// SetIcon specifies the filepath to an icon image that will
// be used for the application.
//
// An empty string value indicates that no icon should be used.
func (c *Config) SetIcon(icon string) {
	c.icon = icon
}

// Icon returns the filepath location of an icon image that
// will be used by the application.
func (c *Config) Icon() string {
	return c.icon
}
