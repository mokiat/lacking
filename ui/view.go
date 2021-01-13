package ui

type View interface {
	Element() *Element
	OpenOverlay()
}
