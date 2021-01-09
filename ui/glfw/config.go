package glfw

type GraphicsEngine int

const (
	GraphicsEngineOpenGL GraphicsEngine = 1 + iota

	GraphicsEngineVulkan
)

func NewAppConfig(title string, width, height int) *AppConfig {
	return &AppConfig{
		title:          title,
		width:          width,
		height:         height,
		graphicsEngine: GraphicsEngineOpenGL,
		swapInterval:   1,
	}
}

type AppConfig struct {
	title          string
	width          int
	height         int
	graphicsEngine GraphicsEngine
	swapInterval   int
}

func (c *AppConfig) SetGraphicsEngine(engine GraphicsEngine) {
	c.graphicsEngine = engine
}

func (c *AppConfig) SetVSync(vsync bool) {
	if vsync {
		c.swapInterval = 1
	} else {
		c.swapInterval = 0
	}
}
