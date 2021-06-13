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

type Camera interface {
	Node

	// FoV returns the field-of-view angle for this camera.
	FoV() sprec.Angle

	// SetFoV changes the field-of-view angle setting of this camera.
	SetFoV(angle sprec.Angle)

	// FoVMode returns the mode of field-of-view. This determines how the
	// FoV setting is used to calculate the final image in the vertical
	// and horizontal directions.
	FoVMode() FoVMode

	// SetFoVMode changes the field-of-view mode of this camera.
	SetFoVMode(mode FoVMode)

	// AspectRatio returns the aspect ratio to be maintained when rendering
	// with this camera. This setting works in combination with FoVMode and
	// FoV settings.
	AspectRatio() float32

	// SetAspectRatio changes the aspect ratio of this camera.
	SetAspectRatio(ratio float32)

	// AutoFocus returns whether this camera will try and automatically
	// focus on objects.
	AutoFocus() bool

	// SetAutoFocus changes whether this camera should attempt to automatically
	// focus on object in the scene.
	SetAutoFocus(enabled bool)

	// FocusRange changes the range from near to far in which the image
	// will be in focus.
	FocusRange() (float32, float32)

	// SetFocusRange changes the focus range for this camera.
	SetFocusRange(near, far float32)

	// AutoExposure returns whether this camera will try and automatically
	// adjust the exposure setting to maintain a proper brightness of
	// the final image.
	AutoExposure() bool

	// SetAutoExposure changes whether this cammera should attempt to
	// do automatic exposure adjustment.
	SetAutoExposure(enabled bool)

	// Exposure returns the exposure setting of this camera.
	Exposure() float32

	// SetExposure changes the exposure setting of this camera.
	// Smaller values mean that the final image will be darker and
	// higher values mean that the final image will be brighter with
	// 1.0 being the starting point.
	SetExposure(exposure float32)

	// Delete removes this camera from the scene.
	Delete()
}
