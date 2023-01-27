package graphics

import "github.com/mokiat/gomath/sprec"

// FoVMode determines how the camera field of view is calculated
// in the horizontal and vertical directions.
type FoVMode string

const (
	// FoVModeAnamorphic will abide to the aspect ratio setting
	// and will pillerbox the image if necessary.
	FoVModeAnamorphic FoVMode = "anamorphic"

	// FoVModeHorizontalPlus will apply the FoV setting to the
	// vertical direction and will adjust the horizontal direction
	// accordingly to preserve the screen's aspect ratio.
	FoVModeHorizontalPlus FoVMode = "horizontal-plus"

	// FoVModeVertialMinus will apply the FoV setting to the
	// horizontal direction and will adjust the vertical direction
	// accordingly to preserve the screen's aspect ratio.
	FoVModeVertialMinus FoVMode = "vertical-minus"

	// FoVModePixelBased will use an orthogonal projection that
	// will match in side the screen pixel size.
	FoVModePixelBased FoVMode = "pixel-based"
)

func newCamera(scene *Scene) *Camera {
	return &Camera{
		Node:              newNode(),
		fov:               sprec.Degrees(120),
		fovMode:           FoVModeHorizontalPlus,
		maxExposure:       10000.0,
		minExposure:       0.00001,
		autoExposureSpeed: 2.0,
		exposure:          1.0,
	}
}

// Camera represents a 3D camera.
type Camera struct {
	Node

	fov                 sprec.Angle
	fovMode             FoVMode
	aspectRatio         float32
	autoFocusEnabled    bool
	nearFocus           float32
	farFocus            float32
	autoExposureEnabled bool
	autoExposureSpeed   float32
	maxExposure         float32
	minExposure         float32
	exposure            float32
}

// FoV returns the field-of-view angle for this camera.
func (c *Camera) FoV() sprec.Angle {
	return c.fov
}

// SetFoV changes the field-of-view angle setting of this camera.
func (c *Camera) SetFoV(angle sprec.Angle) {
	c.fov = angle
}

// FoVMode returns the mode of field-of-view. This determines how the
// FoV setting is used to calculate the final image in the vertical
// and horizontal directions.
func (c *Camera) FoVMode() FoVMode {
	return c.fovMode
}

// SetFoVMode changes the field-of-view mode of this camera.
func (c *Camera) SetFoVMode(mode FoVMode) {
	c.fovMode = mode
}

// AspectRatio returns the aspect ratio to be maintained when rendering
// with this camera. This setting works in combination with FoVMode and
// FoV settings.
func (c *Camera) AspectRatio() float32 {
	return c.aspectRatio
}

// SetAspectRatio changes the aspect ratio of this camera.
func (c *Camera) SetAspectRatio(ratio float32) {
	c.aspectRatio = ratio
}

// AutoFocus returns whether this camera will try and automatically
// focus on objects.
func (c *Camera) AutoFocus() bool {
	return c.autoFocusEnabled
}

// SetAutoFocus changes whether this camera should attempt to automatically
// focus on object in the scene.
func (c *Camera) SetAutoFocus(enabled bool) {
	c.autoFocusEnabled = enabled
}

// FocusRange changes the range from near to far in which the image
// will be in focus.
func (c *Camera) FocusRange() (float32, float32) {
	return c.nearFocus, c.farFocus
}

// SetFocusRange changes the focus range for this camera.
func (c *Camera) SetFocusRange(near, far float32) {
	c.nearFocus = near
	c.farFocus = far
}

// AutoExposure returns whether this camera will try and automatically
// adjust the exposure setting to maintain a proper brightness of
// the final image.
func (c *Camera) AutoExposure() bool {
	return c.autoExposureEnabled
}

// SetAutoExposure changes whether this cammera should attempt to
// do automatic exposure adjustment.
func (c *Camera) SetAutoExposure(enabled bool) {
	c.autoExposureEnabled = enabled
}

// AutoExposureSpeed returns how fast the camera will adjust its exposure.
func (c *Camera) AutoExposureSpeed() float32 {
	return c.autoExposureSpeed
}

// SetAutoExposureSpeed changes the speed at which the camera will automatically
// adjust its exposure.
func (c *Camera) SetAutoExposureSpeed(speed float32) {
	c.autoExposureSpeed = speed
}

// MaximumExposure returns the maximum exposure that the camera may use
// during AutoExposure.
func (c *Camera) MaximumExposure() float32 {
	return c.maxExposure
}

// SetMaximumExposure changes the maximum exposure for this camera.
func (c *Camera) SetMaximumExposure(maxExposure float32) {
	c.maxExposure = maxExposure
}

// MinimumExposure returns the maximum exposure that the camera may use
// during AutoExposure.
func (c *Camera) MinimumExposure() float32 {
	return c.minExposure
}

// SetMinimumExposure changes the maximum exposure for this camera.
func (c *Camera) SetMinimumExposure(minExposure float32) {
	c.minExposure = minExposure
}

// Exposure returns the exposure setting of this camera.
func (c *Camera) Exposure() float32 {
	return c.exposure
}

// SetExposure changes the exposure setting of this camera.
// Smaller values mean that the final image will be darker and
// higher values mean that the final image will be brighter with
// 1.0 being the starting point.
func (c *Camera) SetExposure(exposure float32) {
	c.exposure = exposure
}

// Delete removes this camera from the scene.
func (c *Camera) Delete() {
}
