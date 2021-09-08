package ui

// Canvas is an interface that represents a mechanism
// through which a Control can render itself to the screen.
type Canvas interface {

	// Push records the current state and creates a new
	// state layer. Changes done in the new layer will
	// not affect the former layer.
	Push()

	// Pop restores the former state layer and configures
	// the drawing state accordingly.
	Pop()

	// Translate moves the drawing position by the specified
	// delta amount.
	Translate(delta Position)

	// Clip sets new clipping bounds. Pixels from draw operations
	// that are outside the clipping bounds will not be drawn.
	//
	// Initially the clipping bounds are equal to the window size.
	Clip(bounds Bounds)

	// Shape returns the shape rendering module.
	Shape() Shape

	// Contour returns the contour rendering module.
	Contour() Contour

	// Text returns the text rendering module.
	Text() Text
}

// Shape represents a module for drawing solid shapes.
type Shape interface {

	// Begin starts a new solid shape using the specified fill settings.
	// Make sure to use End when finished with the shape.
	Begin(fill Fill)

	// MoveTo positions the cursor to the specified position.
	MoveTo(position Position)

	// LineTo creates a direct line from the last cursor position
	// to the newly specified position.
	LineTo(position Position)

	// QuadTo creates a quadratic Bezier curve from the last cursor
	// position to the newly specified position by going past the
	// specified control point.
	QuadTo(control, position Position)

	// CubeTo creates a cubic Bezier curve from the last cursor position
	// to the newly specified position by going past the two specified
	// control points.
	CubeTo(control1, control2, position Position)

	// Rectangle is a helper function that draws a rectangle at the
	// specified position and size using a sequence of MoveTo and LineTo
	// instructions.
	Rectangle(position Position, size Size)

	// Triangle is a helper function that draws a triangle with the
	// specified corners, using a sequence of MoveTo and LineTo
	// instructions.
	Triangle(a, b, c Position)

	// Circle is a helper function that draws a circle at the
	// specified position and with the specified radius using a
	// sequence of Shape instructions (whether MoveTo, LineTo or
	// Bezier curves are used is up to the implementation).
	Circle(position Position, radius int)

	// RoundRectangle is a helper function that draws a rounded
	// rectangle at the specified position and with the specified size
	// and corner radiuses.
	RoundRectangle(position Position, size Size, roundness RectRoundness)

	// End marks the end of the shape and pushes all collected data for
	// drawing.
	End()
}

// Contour represents a module for drawing curved lines.
type Contour interface {

	// Begin starts a new contour.
	// Make sure to use End when finished with the contour.
	Begin()

	// MoveTo positions the cursor to the specified position and
	// marks the specified stroke setting for that point.
	MoveTo(position Position, stroke Stroke)

	// LineTo creates a direct line from the last cursor position
	// to the newly specified position and sets the specified
	// stroke for the new position.
	LineTo(position Position, stroke Stroke)

	// QuadTo creates a quadratic Bezier curve from the last cursor
	// position to the newly specified position by going past the
	// specified control point. The target position is assigned
	// the specified stroke setting.
	QuadTo(control, position Position, stroke Stroke)

	// CubeTo creates a cubic Bezier curve from the last cursor position
	// to the newly specified position by going past the two specified
	// control points. The target position is assigned
	// the specified stroke setting.
	CubeTo(control1, control2, position Position, stroke Stroke)

	// CloseLoop makes an automatic line connection back to the starting
	// point, as specified via MoveTo.
	CloseLoop()

	// Rectangle is a helper function that draws the outline of a rectangle
	// at the specified position and size using a sequence of MoveTo and LineTo
	// instructions.
	Rectangle(position Position, size Size, stroke Stroke)

	// Triangle is a helper function that draws the outline of a triangle
	// with the specified corners, using a sequence of MoveTo and LineTo
	// instructions.
	Triangle(a, b, c Position, stroke Stroke)

	// Circle is a helper function that draws a circle at the
	// specified position and with the specified radius using a
	// sequence of Shape instructions (whether MoveTo, LineTo or
	// Bezier curves are used is up to the implementation).
	Circle(position Position, radius int, stroke Stroke)

	// RoundRectangle is a helper function that draws a rounded
	// rectangle at the specified position and with the specified size
	// and corner radiuses.
	RoundRectangle(position Position, size Size, roundness RectRoundness, stroke Stroke)

	// End marks the end of the contour and pushes all collected data for
	// drawing.
	End()
}

// Text represents a module for drawing text.
type Text interface {

	// Begin starts a new text sequence using the specified typography settings.
	// Make sure to use End when finished with the text.
	Begin(typography Typography)

	// Line draws a text line at the specified position.
	Line(value string, position Position)

	// End marks the end of the text and pushes all collected data for
	// drawing.
	End()
}

// Fill configures how a solid shape is to be drawn.
type Fill struct {

	// Rule specifies the mechanism through which it is determined
	// which point is part of the shape in an overlapping or concave
	// polygon.
	Rule FillRule

	// Color specifies the color to use to fill the shape.
	Color Color

	// Image specifies an optional image to be used for filling
	// the shape.
	Image Image

	// ImageOffset determines the offset of the origin of the
	// image relative to the current translation context.
	ImageOffset Position

	// ImageSize determines the size of the drawn image. In
	// essence, this size performs scaling.
	ImageSize Size
}

// FillRule represents the mechanism through which it is determined
// which point is part of the shape in an overlapping or concave
// polygon.
type FillRule int

const (
	// FillRuleSimple is the fastest approach and should be used
	// with non-overlapping concave shapes.
	FillRuleSimple FillRule = iota

	// FillRuleNonZero will fill areas that are covered by the
	// shape, regardless if it overlaps.
	FillRuleNonZero

	// FillRuleEvenOdd will fill areas that are covered by the
	// shape and it does not overlap or overlaps an odd number
	// of times.
	FillRuleEvenOdd
)

// Stroke configures how a contour is to be drawn.
type Stroke struct {

	// Size determines the size of the contour.
	Size int

	// Color specifies the color of the contour.
	Color Color
}

// Typography configures how text is to be drawn.
type Typography struct {

	// Font specifies the font to be used.
	Font Font

	// Size specifies the font size.
	Size int

	// Color indicates the color of the text.
	Color Color
}

// RectRoundness is used to configure the roundness of
// a round rectangle through corner radiuses.
type RectRoundness struct {

	// TopLeftRadius specifies the radius of the top-left corner.
	TopLeftRadius int

	// TopRightRadius specifies the radius of the top-right corner.
	TopRightRadius int

	// BottomLeftRadius specifies the radius of the bottom-left corner.
	BottomLeftRadius int

	// BottomRightRadius specifies the radius of the bottom-right corner.
	BottomRightRadius int
}
