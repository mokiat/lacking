package ui

var registry map[string]Builder

func init() {
	registry = make(map[string]Builder)
}

func Register(name string, builder Builder) {
	registry[name] = builder
}
