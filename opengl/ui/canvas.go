package ui

import (
	"encoding/binary"
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/opengl"
	"github.com/mokiat/lacking/ui"
)

const (
	layerCount = 256

	positionAttribIndex = 0
	texCoordAttribIndex = 1
	colorAttribIndex    = 2
)

type SolidVertex struct {
	Position sprec.Vec3
	Color    ui.Color
}

func (v SolidVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
	plotter.PlotFloat32(v.Position.Z)
	plotter.PlotByte(v.Color.R)
	plotter.PlotByte(v.Color.G)
	plotter.PlotByte(v.Color.B)
	plotter.PlotByte(v.Color.A)
}

type SolidTriangle struct {
	Vertices [3]SolidVertex
}

type SolidRenderSet struct {
	Presentation *presentation
	Triangles    []SolidTriangle
	VertexData   []byte
	VertexBuffer *opengl.Buffer
	VertexArray  *opengl.VertexArray
}

type ImageVertex struct {
	Position sprec.Vec3
	TexCoord sprec.Vec2
}

func (v ImageVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
	plotter.PlotFloat32(v.Position.Z)
	plotter.PlotFloat32(v.TexCoord.X)
	plotter.PlotFloat32(v.TexCoord.Y)
}

type ImageTriangle struct {
	Texture  *opengl.TwoDTexture
	Vertices [3]ImageVertex
}

type ImageRenderSet struct {
	Presentation *presentation
	Triangles    []ImageTriangle
	VertexData   []byte
	VertexBuffer *opengl.Buffer
	VertexArray  *opengl.VertexArray
}

func NewCanvas() *Canvas {
	return &Canvas{
		layers: make([]stateLayer, layerCount),
	}
}

type Canvas struct {
	layers     []stateLayer
	layerDepth int

	windowSize       ui.Size
	projectionMatrix sprec.Mat4

	solidRenderSet SolidRenderSet
	imageRenderSet ImageRenderSet
}

func (c *Canvas) Init() error {
	solidPresentation, err := newSolidPresentation()
	if err != nil {
		return fmt.Errorf("failed to create solid presentation: %w", err)
	}

	solidVertexData := make([]byte, 1<<15)
	solidVertexBufferInfo := opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    solidVertexData,
	}
	solidVertexBuffer := opengl.NewBuffer()
	if err := solidVertexBuffer.Allocate(solidVertexBufferInfo); err != nil {
		return err
	}
	solidVertexArrayInfo := opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: solidVertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  3*4 + 4,
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          positionAttribIndex,
				ComponentCount: 3,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
			{
				Index:          colorAttribIndex,
				ComponentCount: 4,
				ComponentType:  gl.UNSIGNED_BYTE,
				Normalized:     true,
				OffsetBytes:    3 * 4,
				BufferBinding:  0,
			},
		},
	}
	solidVertexArray := opengl.NewVertexArray()
	if err := solidVertexArray.Allocate(solidVertexArrayInfo); err != nil {
		return err
	}

	c.solidRenderSet = SolidRenderSet{
		Presentation: solidPresentation,
		Triangles:    make([]SolidTriangle, 1024),
		VertexData:   solidVertexData,
		VertexBuffer: solidVertexBuffer,
		VertexArray:  solidVertexArray,
	}

	imagePresentation, err := newImagePresentation()
	if err != nil {
		return fmt.Errorf("failed to create image presentation: %w", err)
	}

	imageVertexData := make([]byte, 1<<15)
	imageVertexBufferInfo := opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    imageVertexData,
	}
	imageVertexBuffer := opengl.NewBuffer()
	if err := imageVertexBuffer.Allocate(imageVertexBufferInfo); err != nil {
		return err
	}
	imageVertexArrayInfo := opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: imageVertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  3*4 + 2*4,
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          positionAttribIndex,
				ComponentCount: 3,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
			{
				Index:          texCoordAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    3 * 4,
				BufferBinding:  0,
			},
		},
	}
	imageVertexArray := opengl.NewVertexArray()
	if err := imageVertexArray.Allocate(imageVertexArrayInfo); err != nil {
		return err
	}

	c.imageRenderSet = ImageRenderSet{
		Presentation: imagePresentation,
		Triangles:    make([]ImageTriangle, 1024),
		VertexData:   imageVertexData,
		VertexBuffer: imageVertexBuffer,
		VertexArray:  imageVertexArray,
	}

	return nil
}

func (c *Canvas) Release() error {
	if err := c.solidRenderSet.VertexArray.Release(); err != nil {
		return err
	}
	if err := c.solidRenderSet.VertexBuffer.Release(); err != nil {
		return err
	}
	if err := releasePresentation(c.solidRenderSet.Presentation); err != nil {
		return err
	}

	if err := c.imageRenderSet.VertexArray.Release(); err != nil {
		return err
	}
	if err := c.imageRenderSet.VertexBuffer.Release(); err != nil {
		return err
	}
	if err := releasePresentation(c.imageRenderSet.Presentation); err != nil {
		return err
	}
	return nil
}

func (c *Canvas) Resize(size ui.Size) {
	c.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	)
	gl.Viewport(0, 0, int32(size.Width), int32(size.Height))
}

func (c *Canvas) ResizeFramebuffer(size ui.Size) {
}

func (c *Canvas) Reset() {
	c.solidRenderSet.Triangles = c.solidRenderSet.Triangles[:0]
	c.imageRenderSet.Triangles = c.imageRenderSet.Triangles[:0]

	c.layerDepth = 0
	rootLayer := c.layerAt(c.layerDepth)
	rootLayer.Translation = ui.NewPosition(0, 0)
	rootLayer.ClipBounds = ui.Bounds{
		Position: ui.NewPosition(0, 0),
		Size:     c.windowSize,
	}
}

func (c *Canvas) Push() {
	oldLayer := c.layerAt(c.layerDepth)
	if c.layerDepth++; c.layerDepth >= len(c.layers) {
		panic("maximum state depth exceeded: too many pushes")
	}
	newLayer := c.layerAt(c.layerDepth)
	*newLayer = *oldLayer
}

func (c *Canvas) Pop() {
	if c.layerDepth--; c.layerDepth < 0 {
		panic("minimum state depth exceeded: too many pops")
	}
}

func (c *Canvas) Translate(delta ui.Position) {
	layer := c.currentLayer()
	layer.Translation = layer.Translation.Translate(delta.X, delta.Y)
}

func (c *Canvas) Clip(bounds ui.Bounds) {
	layer := c.currentLayer()
	layer.ClipBounds = bounds.Translate(layer.Translation)
}

func (c *Canvas) UseSolidColor(color ui.Color) {
	layer := c.currentLayer()
	layer.SolidColor = color
}

func (c *Canvas) DrawRectangle(position ui.Position, size ui.Size) {
	layer := c.currentLayer()
	c.solidRenderSet.Triangles = append(c.solidRenderSet.Triangles, SolidTriangle{
		Vertices: [3]SolidVertex{
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
		},
	})
	c.solidRenderSet.Triangles = append(c.solidRenderSet.Triangles, SolidTriangle{
		Vertices: [3]SolidVertex{
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				Color: layer.SolidColor,
			},
		},
	})
}

func (c *Canvas) DrawImage(img ui.Image, position ui.Position, size ui.Size) {
	openglImg := img.(*Image)
	layer := c.currentLayer()
	c.imageRenderSet.Triangles = append(c.imageRenderSet.Triangles, ImageTriangle{
		Texture: openglImg.texture,
		Vertices: [3]ImageVertex{
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(0.0, 1.0),
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(0.0, 0.0),
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(1.0, 0.0),
			},
		},
	})
	c.imageRenderSet.Triangles = append(c.imageRenderSet.Triangles, ImageTriangle{
		Texture: openglImg.texture,
		Vertices: [3]ImageVertex{
			{
				Position: sprec.NewVec3(
					float32(position.X+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(0.0, 1.0),
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+size.Height+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(1.0, 0.0),
			},
			{
				Position: sprec.NewVec3(
					float32(position.X+size.Width+layer.Translation.X),
					float32(position.Y+layer.Translation.Y),
					float32(-0.1),
				),
				TexCoord: sprec.NewVec2(1.0, 1.0),
			},
		},
	})
}

func (c *Canvas) CreateImage(data asset.TwoDTexture) (ui.Image, error) {
	info := opengl.TwoDTextureAllocateInfo{
		Width:  int32(data.Width),
		Height: int32(data.Height),
		Data:   data.Data,
	}

	texture := opengl.NewTwoDTexture()
	if err := texture.Allocate(info); err != nil {
		return nil, fmt.Errorf("failed to allocate image: %w", err)
	}

	return &Image{
		texture: texture,
		size:    ui.NewSize(int(data.Width), int(data.Height)),
	}, nil
}

func (c *Canvas) Flush() {
	{
		if len(c.solidRenderSet.Triangles) > 0 {
			plotter := buffer.NewPlotter(c.solidRenderSet.VertexData, binary.LittleEndian)
			for _, triangle := range c.solidRenderSet.Triangles {
				triangle.Vertices[0].Serialize(plotter)
				triangle.Vertices[1].Serialize(plotter)
				triangle.Vertices[2].Serialize(plotter)
			}
			updateInfo := opengl.BufferUpdateInfo{
				Data:        c.solidRenderSet.VertexData[:plotter.Offset()],
				OffsetBytes: 0,
			}
			if err := c.solidRenderSet.VertexBuffer.Update(updateInfo); err != nil {
				panic(err)
			}
		}
	}
	{
		if len(c.imageRenderSet.Triangles) > 0 {
			plotter := buffer.NewPlotter(c.imageRenderSet.VertexData, binary.LittleEndian)
			for _, triangle := range c.imageRenderSet.Triangles {
				triangle.Vertices[0].Serialize(plotter)
				triangle.Vertices[1].Serialize(plotter)
				triangle.Vertices[2].Serialize(plotter)
			}
			updateInfo := opengl.BufferUpdateInfo{
				Data:        c.imageRenderSet.VertexData[:plotter.Offset()],
				OffsetBytes: 0,
			}
			if err := c.imageRenderSet.VertexBuffer.Update(updateInfo); err != nil {
				panic(err)
			}
		}
	}

	gl.ClearColor(1.0, 0.5, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.CULL_FACE)

	{
		if len(c.solidRenderSet.Triangles) > 0 {
			pres := c.solidRenderSet.Presentation
			gl.UseProgram(pres.program.ID())
			gl.UniformMatrix4fv(pres.projectionMatrixLocation, 1, false, matrixToArray(c.projectionMatrix))
			gl.BindVertexArray(c.solidRenderSet.VertexArray.ID())
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.solidRenderSet.Triangles)*3))
		}
	}

	{
		if len(c.imageRenderSet.Triangles) > 0 {
			pres := c.imageRenderSet.Presentation
			gl.UseProgram(pres.program.ID())
			gl.UniformMatrix4fv(pres.projectionMatrixLocation, 1, false, matrixToArray(c.projectionMatrix))
			gl.BindVertexArray(c.imageRenderSet.VertexArray.ID())
			for i, triangle := range c.imageRenderSet.Triangles {
				gl.BindTextureUnit(0, triangle.Texture.ID())
				gl.Uniform1i(pres.textureLocation, 0)
				gl.DrawArrays(gl.TRIANGLES, int32(i*3), int32(3))
			}
		}
	}
}

func (c *Canvas) layerAt(depth int) *stateLayer {
	return &c.layers[depth]
}

func (c *Canvas) currentLayer() *stateLayer {
	return c.layerAt(c.layerDepth)
}

type stateLayer struct {
	Translation ui.Position
	ClipBounds  ui.Bounds
	SolidColor  ui.Color
}

// NOTE: Use this method only as short-lived function argument
// subsequent calls will reuse the same float32 array
func matrixToArray(matrix sprec.Mat4) *float32 {
	var matrixCache [16]float32
	matrixCache[0] = matrix.M11
	matrixCache[1] = matrix.M21
	matrixCache[2] = matrix.M31
	matrixCache[3] = matrix.M41

	matrixCache[4] = matrix.M12
	matrixCache[5] = matrix.M22
	matrixCache[6] = matrix.M32
	matrixCache[7] = matrix.M42

	matrixCache[8] = matrix.M13
	matrixCache[9] = matrix.M23
	matrixCache[10] = matrix.M33
	matrixCache[11] = matrix.M43

	matrixCache[12] = matrix.M14
	matrixCache[13] = matrix.M24
	matrixCache[14] = matrix.M34
	matrixCache[15] = matrix.M44
	return &matrixCache[0]
}
