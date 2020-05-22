package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/gomath/sprec"
)

func NewRenderer() *Renderer {
	freePipeline := make(chan *Pipeline, 1)
	freePipeline <- newPipeline()

	queuedPipeline := make(chan *Pipeline, 1)
	queuedPipeline <- newPipeline()

	return &Renderer{
		freePipeline:   freePipeline,
		queuedPipeline: queuedPipeline,
		activePipeline: <-queuedPipeline,
		matrixCache:    make([]float32, 16),
	}
}

type Renderer struct {
	freePipeline   chan *Pipeline
	queuedPipeline chan *Pipeline
	activePipeline *Pipeline
	matrixCache    []float32
}

func (r *Renderer) BeginPipeline() *Pipeline {
	var pipeline *Pipeline
	select {
	case pipeline = <-r.freePipeline:
	default:
		select {
		case pipeline = <-r.queuedPipeline:
		default:
			pipeline = <-r.freePipeline
		}
	}
	pipeline.rewind()
	return pipeline
}

func (r *Renderer) EndPipeline(pipeline *Pipeline) {
	select {
	case r.queuedPipeline <- pipeline:
	default:
		select {
		case r.freePipeline <- pipeline:
		default:
			r.queuedPipeline <- pipeline
		}
	}
}

func (r *Renderer) Render() {
	select {
	case pipeline := <-r.queuedPipeline:
		r.activePipeline.rewind()
		r.freePipeline <- r.activePipeline
		r.activePipeline = pipeline
	default:
	}
	r.renderPipeline(r.activePipeline)
}

func (r *Renderer) renderPipeline(pipeline *Pipeline) {
	for _, action := range pipeline.preRenderActionsView() {
		r.processAction(action)
	}
	for _, sequence := range pipeline.sequencesView() {
		r.renderSequence(sequence)
	}
	for _, action := range pipeline.postRenderActionsView() {
		r.processAction(action)
	}
}

func (r *Renderer) processAction(action func()) {
	action()
}

func (r *Renderer) renderSequence(sequence Sequence) {
	gl.Enable(gl.CULL_FACE)

	if framebuffer := sequence.SourceFramebuffer; framebuffer != nil {
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer.ID)
	}
	if framebuffer := sequence.TargetFramebuffer; framebuffer != nil {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, framebuffer.ID)
		gl.Viewport(0, 0, framebuffer.Width, framebuffer.Height)
	} else {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}
	blitFlags := uint32(0)
	if sequence.BlitFramebufferColor {
		blitFlags |= gl.COLOR_BUFFER_BIT
	}
	if sequence.BlitFramebufferDepth {
		blitFlags |= gl.DEPTH_BUFFER_BIT
	}
	if blitFlags != 0 {
		sampleMode := uint32(gl.NEAREST)
		if sequence.BlitFramebufferSmooth {
			sampleMode = uint32(gl.LINEAR)
		}
		gl.BlitFramebuffer(
			0, 0, sequence.SourceFramebuffer.Width, sequence.SourceFramebuffer.Height,
			0, 0, sequence.TargetFramebuffer.Width, sequence.TargetFramebuffer.Height,
			blitFlags,
			sampleMode,
		)
	}

	clearFlags := uint32(0)
	if sequence.ClearColor {
		clearFlags |= gl.COLOR_BUFFER_BIT
	}
	if sequence.ClearDepth {
		clearFlags |= gl.DEPTH_BUFFER_BIT
	}
	if clearFlags != 0 {
		color := sequence.BackgroundColor
		gl.ClearColor(color.X, color.Y, color.Z, color.W)
		gl.Clear(clearFlags)
	}

	if sequence.TestDepth {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	gl.DepthMask(sequence.WriteDepth)

	switch sequence.DepthFunc {
	case DepthFuncLess:
		gl.DepthFunc(gl.LESS)
	case DepthFuncLessOrEqual:
		gl.DepthFunc(gl.LEQUAL)
	}

	for _, item := range sequence.itemsView() {
		r.renderItem(sequence, item)
	}

	gl.DepthMask(true) // TODO: Remove, once old renderer is scrapped
}

func (r *Renderer) renderItem(sequence Sequence, item Item) {
	gl.UseProgram(item.Program.ID)

	textureIndex := uint32(0)
	if item.Program.DiffuseTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, item.DiffuseTexture.ID)
		gl.Uniform1i(item.Program.DiffuseTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.SkyboxTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, item.SkyboxTexture.ID)
		gl.Uniform1i(item.Program.SkyboxTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.FBAlbedoTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, sequence.SourceFramebuffer.AlbedoTextureID)
		gl.Uniform1i(item.Program.FBAlbedoTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.FBNormalTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, sequence.SourceFramebuffer.NormalTextureID)
		gl.Uniform1i(item.Program.FBNormalTextureLocation, int32(textureIndex))
		textureIndex++
	}

	if item.Program.DiffuseColorLocation != -1 {
		gl.Uniform4f(item.Program.DiffuseColorLocation, item.DiffuseColor.X, item.DiffuseColor.Y, item.DiffuseColor.Z, item.DiffuseColor.W)
	}
	if item.Program.ProjectionMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ProjectionMatrixLocation, 1, false, r.matrixToArray(sequence.ProjectionMatrix))
	}
	if item.Program.ViewMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ViewMatrixLocation, 1, false, r.matrixToArray(sequence.ViewMatrix))
	}
	if item.Program.ModelMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ModelMatrixLocation, 1, false, r.matrixToArray(item.ModelMatrix))
	}

	gl.BindVertexArray(item.VertexArray.ID)
	switch item.Primitive {
	case RenderPrimitiveTriangles:
		gl.DrawElements(gl.TRIANGLES, item.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(0))
	case RenderPrimitiveLines:
		gl.LineWidth(2)
		gl.DrawElements(gl.LINES, item.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(0))
	}
}

// NOTE: Use this method only as short-lived function argument
// subsequent calls will reuse the same float32 array
func (r *Renderer) matrixToArray(matrix sprec.Mat4) *float32 {
	r.matrixCache[0] = matrix.M11
	r.matrixCache[1] = matrix.M21
	r.matrixCache[2] = matrix.M31
	r.matrixCache[3] = matrix.M41

	r.matrixCache[4] = matrix.M12
	r.matrixCache[5] = matrix.M22
	r.matrixCache[6] = matrix.M32
	r.matrixCache[7] = matrix.M42

	r.matrixCache[8] = matrix.M13
	r.matrixCache[9] = matrix.M23
	r.matrixCache[10] = matrix.M33
	r.matrixCache[11] = matrix.M43

	r.matrixCache[12] = matrix.M14
	r.matrixCache[13] = matrix.M24
	r.matrixCache[14] = matrix.M34
	r.matrixCache[15] = matrix.M44
	return &r.matrixCache[0]
}
