package ui

type Template interface {
	ID() string
	Name() string
	Attributes() AttributeSet
	Children() []Template
}
