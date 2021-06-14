package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

func newCamera(scene *Scene) *Camera {
	return &Camera{
		Node:     internal.NewNode(),
		scene:    scene,
		fov:      sprec.Degrees(120),
		fovMode:  graphics.FoVModeHorizontalPlus,
		exposure: 1.0,
	}
}

var _ graphics.Camera = (*Camera)(nil)

type Camera struct {
	internal.Node

	scene *Scene

	fov                 sprec.Angle
	fovMode             graphics.FoVMode
	aspectRatio         float32
	autoFocusEnabled    bool
	nearFocus           float32
	farFocus            float32
	autoExposureEnabled bool
	exposure            float32
}

func (c *Camera) FoV() sprec.Angle {
	return c.fov
}

func (c *Camera) SetFoV(angle sprec.Angle) {
	c.fov = angle
}

func (c *Camera) FoVMode() graphics.FoVMode {
	return c.fovMode
}

func (c *Camera) SetFoVMode(mode graphics.FoVMode) {
	c.fovMode = mode
}

func (c *Camera) AspectRatio() float32 {
	return c.aspectRatio
}

func (c *Camera) SetAspectRatio(ratio float32) {
	c.aspectRatio = ratio
}

func (c *Camera) AutoFocus() bool {
	return c.autoFocusEnabled
}

func (c *Camera) SetAutoFocus(enabled bool) {
	c.autoFocusEnabled = enabled
}

func (c *Camera) FocusRange() (float32, float32) {
	return c.nearFocus, c.farFocus
}

func (c *Camera) SetFocusRange(near, far float32) {
	c.nearFocus = near
	c.farFocus = far
}

func (c *Camera) AutoExposure() bool {
	return c.autoExposureEnabled
}

func (c *Camera) SetAutoExposure(enabled bool) {
	c.autoExposureEnabled = enabled
}

func (c *Camera) Exposure() float32 {
	return c.exposure
}

func (c *Camera) SetExposure(exposure float32) {
	c.exposure = exposure
}

func (c *Camera) Delete() {
}