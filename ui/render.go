package ui

type RenderContext struct {
	Canvas      Canvas
	DirtyRegion Bounds
}

type Renderable interface {
	Render(element *Element, ctx RenderContext)
}

func renderElement(element *Element, ctx RenderContext) {
	dirtyRegion := ctx.DirtyRegion.Intersect(element.bounds)
	if dirtyRegion.Empty() {
		return
	}

	ctx.Canvas.Push()
	ctx.Canvas.Clip(element.bounds)
	ctx.Canvas.Translate(element.bounds.Position)
	elementCtx := RenderContext{
		Canvas:      ctx.Canvas,
		DirtyRegion: dirtyRegion.Translate(element.bounds.Position.Inverse()),
	}

	if renderable, ok := element.handler.(Renderable); ok {
		renderable.Render(element, elementCtx)
	}
	renderElementContent(element, elementCtx)

	ctx.Canvas.Pop()
}

func renderElementContent(element *Element, ctx RenderContext) {
	if element.firstChild == nil {
		return
	}
	contentBounds := element.ContentBounds()
	if contentBounds.Empty() {
		return
	}
	ctx.Canvas.Push()
	ctx.Canvas.Clip(contentBounds)
	ctx.Canvas.Translate(contentBounds.Position)
	for child := element.firstChild; child != nil; child = child.rightSibling {
		renderElement(child, RenderContext{
			Canvas:      ctx.Canvas,
			DirtyRegion: ctx.DirtyRegion.Translate(contentBounds.Position.Inverse()),
		})
	}
	ctx.Canvas.Pop()
}
