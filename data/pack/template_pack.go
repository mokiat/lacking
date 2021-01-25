package pack

type UITemplateProvider interface {
	Template() *UITemplate
}

type UITemplate struct {
	Name       string
	Attributes map[string]string
	Children   []UITemplate
}
