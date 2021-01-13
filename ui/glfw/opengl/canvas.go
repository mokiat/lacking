package opengl

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

const layerCount = 256

func NewCanvas() *Canvas {
	return &Canvas{
		layers: make([]stateLayer, layerCount),
	}
}

type Canvas struct {
	layers           []stateLayer
	layerDepth       int
	windowSize       ui.Size
	projectionMatrix sprec.Mat4
}

func (c *Canvas) Init() error {
	// create all shaders and stuff
	return nil
}

func (c *Canvas) Resize(size ui.Size) {
	c.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	)
}

func (c *Canvas) ResizeFramebuffer(size ui.Size) {
}

func (c *Canvas) Reset() {
	c.layerDepth = 0
	rootLayer := c.layerAt(c.layerDepth)
	rootLayer.Translation = ui.NewPosition(0, 0)
	rootLayer.ClipBounds = ui.Bounds{
		Position: ui.NewPosition(0, 0),
		Size:     c.windowSize,
	}

	gl.ClearColor(1.0, 0.5, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (c *Canvas) Push() {
	oldLayer := c.layerAt(c.layerDepth)
	if c.layerDepth++; c.layerDepth >= len(c.layers) {
		panic("maximum state depth exceeded: too many pushes")
	}
	newLayer := c.layerAt(c.layerDepth)
	*newLayer = *oldLayer
}

func (c *Canvas) Pop() {
	if c.layerDepth--; c.layerDepth < 0 {
		panic("minimum state depth exceeded: too many pops")
	}
}

func (c *Canvas) Translate(delta ui.Position) {
	layer := c.layerAt(c.layerDepth)
	layer.Translation = layer.Translation.Translate(delta.X, delta.Y)
}

func (c *Canvas) Clip(bounds ui.Bounds) {
	layer := c.layerAt(c.layerDepth)
	layer.ClipBounds = bounds.Translate(layer.Translation)
}

func (c *Canvas) layerAt(depth int) *stateLayer {
	return &c.layers[depth]
}

type stateLayer struct {
	Translation ui.Position
	ClipBounds  ui.Bounds
}
