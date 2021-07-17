package template

var dslCtx = &dslContext{}

type dslContext struct {
	parent   *dslContext
	instance Instance
}
