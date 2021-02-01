package pack

type UITemplateProvider interface {
	Template(ctx *Context) (*UITemplate, error)
}

type UITemplate struct {
	Name       string
	Attributes map[string]string
	Children   []UITemplate
}
