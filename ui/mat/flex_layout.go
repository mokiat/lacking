package mat

// import "github.com/mokiat/lacking/ui"

// type FlexDirection int

// const (
// 	FlexDirectionHorizontal FlexDirection = iota
// 	FlexDirectionVertical
// )

// // FlexLayoutSettings contains optional configurations for the
// // FlexLayout.
// type FlexLayoutSettings struct {
// 	Direction FlexDirection
// }

// // NewFlexLayout creates a new FlexLayout instance.
// func NewFlexLayout(settings FlexLayoutSettings) *FlexLayout {
// 	return &FlexLayout{
// 		direction: settings.Direction,
// 	}
// }

// var _ ui.Layout = (*FlexLayout)(nil)

// // FlexLayout is an implementation of Layout that positions and
// // resizes elements according to a flexible, directional layout.
// type FlexLayout struct {
// 	direction FlexDirection
// }

// // Apply applies this layout to the specified Element.
// func (l *FlexLayout) Apply(element *ui.Element) {
// 	switch l.direction {
// 	case FlexDirectionHorizontal:
// 		l.applyHorizontal(element)
// 	case FlexDirectionVertical:
// 		l.applyVertical(element)
// 	}
// }

// func (l *FlexLayout) applyHorizontal(element *ui.Element) {

// 	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
// 		// TODO

// 		childElement.SetBounds(element.ContentBounds())
// 	}
// }

// func (l *FlexLayout) applyVertical(element *ui.Element) {
// 	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
// 		// TODO

// 		childElement.SetBounds(element.ContentBounds())
// 	}
// }
