package ui

type Template interface {
	ID() string
	Name() string
	BoolAttribute(name string) (bool, bool)
	IntAttribute(name string) (int, bool)
	FloatAttribute(name string) (float64, bool)
	StringAttribute(name string) (string, bool)
	Children() []Template
}
