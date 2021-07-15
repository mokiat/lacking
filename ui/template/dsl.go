package template

var dslCtx = &dslContext{
	shouldReconcile: false,
}

type dslContext struct {
	parent          *dslContext
	shouldReconcile bool
	instance        Instance
}
