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
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	if framebuffer := sequence.SourceFramebuffer; framebuffer != nil {
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer.ID)
	}
	if framebuffer := sequence.TargetFramebuffer; framebuffer != nil {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, framebuffer.ID)
		gl.Viewport(0, 0, framebuffer.Width, framebuffer.Height)
	} else {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	if sequence.TestDepth {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	gl.DepthMask(sequence.WriteDepth)

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

	switch sequence.DepthFunc {
	case DepthFuncLess:
		gl.DepthFunc(gl.LESS)
	case DepthFuncLessOrEqual:
		gl.DepthFunc(gl.LEQUAL)
	}

	for _, item := range sequence.itemsView() {
		r.renderItem(sequence, item)
	}
}

func (r *Renderer) renderItem(sequence Sequence, item Item) {
	if item.BackfaceCulling {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}

	gl.UseProgram(item.Program.ID)

	textureIndex := uint32(0)
	if item.Program.MetalnessTwoDTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, item.MetalnessTwoDTexture.ID)
		gl.Uniform1i(item.Program.MetalnessTwoDTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.RoughnessTwoDTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, item.RoughnessTwoDTexture.ID)
		gl.Uniform1i(item.Program.RoughnessTwoDTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.AlbedoTwoDTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, item.AlbedoTwoDTexture.ID)
		gl.Uniform1i(item.Program.AlbedoTwoDTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.AlbedoCubeTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, item.AlbedoCubeTexture.ID)
		gl.Uniform1i(item.Program.AlbedoCubeTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.NormalTwoDTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, item.NormalTwoDTexture.ID)
		gl.Uniform1i(item.Program.NormalTwoDTextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.FBColor0TextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, sequence.SourceFramebuffer.AlbedoTextureID)
		gl.Uniform1i(item.Program.FBColor0TextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.FBColor1TextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, sequence.SourceFramebuffer.NormalTextureID)
		gl.Uniform1i(item.Program.FBColor1TextureLocation, int32(textureIndex))
		textureIndex++
	}
	if item.Program.FBDepthTextureLocation != -1 {
		gl.ActiveTexture(gl.TEXTURE0 + textureIndex)
		gl.BindTexture(gl.TEXTURE_2D, sequence.SourceFramebuffer.DepthTextureID)
		gl.Uniform1i(item.Program.FBDepthTextureLocation, int32(textureIndex))
		textureIndex++
	}

	if item.Program.MetalnessLocation != -1 {
		gl.Uniform1f(item.Program.MetalnessLocation, item.Metalness)
	}
	if item.Program.RoughnessLocation != -1 {
		gl.Uniform1f(item.Program.RoughnessLocation, item.Roughness)
	}
	if item.Program.AlbedoColorLocation != -1 {
		gl.Uniform4f(item.Program.AlbedoColorLocation, item.AlbedoColor.X, item.AlbedoColor.Y, item.AlbedoColor.Z, item.AlbedoColor.W)
	}
	if item.Program.NormalScaleLocation != -1 {
		gl.Uniform1f(item.Program.NormalScaleLocation, item.NormalScale)
	}

	if item.Program.ProjectionMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ProjectionMatrixLocation, 1, false, r.matrixToArray(sequence.ProjectionMatrix))
	}
	if item.Program.ModelMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ModelMatrixLocation, 1, false, r.matrixToArray(item.ModelMatrix))
	}
	if item.Program.ViewMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.ViewMatrixLocation, 1, false, r.matrixToArray(sequence.ViewMatrix))
	}
	if item.Program.CameraMatrixLocation != -1 {
		gl.UniformMatrix4fv(item.Program.CameraMatrixLocation, 1, false, r.matrixToArray(sequence.CameraMatrix))
	}

	gl.BindVertexArray(item.VertexArray.ID)
	gl.LineWidth(2)
	gl.DrawElements(item.glPrimitive(), item.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(item.IndexOffset))
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
