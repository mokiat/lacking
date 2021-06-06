package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/game/graphics"
)

func NewEngine() *Engine {
	return &Engine{
		renderer: newRenderer(),
	}
}

var _ graphics.Engine = (*Engine)(nil)

type Engine struct {
	renderer *Renderer
}

func (e *Engine) Create() {
	e.renderer.Allocate()
}

func (e *Engine) CreateScene() graphics.Scene {
	return newScene(e.renderer)
}

func (e *Engine) CreateCubeTexture(definition graphics.CubeTextureDefinition) graphics.CubeTexture {
	allocateInfo := opengl.CubeTextureAllocateInfo{
		Dimension:      int32(definition.Dimension),
		FrontSideData:  definition.FrontSideData,
		BackSideData:   definition.BackSideData,
		LeftSideData:   definition.LeftSideData,
		RightSideData:  definition.RightSideData,
		TopSideData:    definition.TopSideData,
		BottomSideData: definition.BottomSideData,
	}
	switch definition.WrapS {
	case graphics.WrapClampToEdge:
		allocateInfo.WrapS = gl.CLAMP_TO_EDGE
	case graphics.WrapRepeat:
		allocateInfo.WrapS = gl.REPEAT
	default:
		panic(fmt.Errorf("unknown wrap mode: %d", definition.WrapS))
	}
	switch definition.WrapT {
	case graphics.WrapClampToEdge:
		allocateInfo.WrapT = gl.CLAMP_TO_EDGE
	case graphics.WrapRepeat:
		allocateInfo.WrapT = gl.REPEAT
	default:
		panic(fmt.Errorf("unknown wrap mode: %d", definition.WrapT))
	}
	switch definition.MinFilter {
	case graphics.FilterNearest:
		allocateInfo.MinFilter = gl.NEAREST
	case graphics.FilterLinear:
		allocateInfo.MinFilter = gl.LINEAR
	case graphics.FilterNearestMipmapNearest:
		allocateInfo.MinFilter = gl.NEAREST_MIPMAP_NEAREST
	case graphics.FilterNearestMipmapLinear:
		allocateInfo.MinFilter = gl.NEAREST_MIPMAP_LINEAR
	case graphics.FilterLinearMipmapNearest:
		allocateInfo.MinFilter = gl.LINEAR_MIPMAP_NEAREST
	case graphics.FilterLinearMipmapLinear:
		allocateInfo.MinFilter = gl.LINEAR_MIPMAP_LINEAR
	default:
		panic(fmt.Errorf("unknown filter mode: %d", definition.MinFilter))
	}
	switch definition.MagFilter {
	case graphics.FilterNearest:
		allocateInfo.MagFilter = gl.NEAREST
	case graphics.FilterLinear:
		allocateInfo.MagFilter = gl.LINEAR
	default:
		panic(fmt.Errorf("unknown filter mode: %d", definition.MagFilter))
	}
	switch definition.DataFormat {
	case graphics.DataFormatRGBA8:
		allocateInfo.DataComponentType = gl.UNSIGNED_BYTE
		allocateInfo.DataFormat = gl.RGBA
	case graphics.DataFormatRGBA32F:
		allocateInfo.DataComponentType = gl.FLOAT
		allocateInfo.DataFormat = gl.RGBA
	default:
		panic(fmt.Errorf("unknown data format: %d", definition.DataFormat))
	}
	switch definition.InternalFormat {
	case graphics.InternalFormatRGBA8:
		allocateInfo.InternalFormat = gl.SRGB8_ALPHA8
	case graphics.InternalFormatRGBA32F:
		allocateInfo.InternalFormat = gl.RGBA32F
	default:
		panic(fmt.Errorf("unknown internal format: %d", definition.InternalFormat))
	}
	result := newCubeTexture()
	result.CubeTexture.Allocate(allocateInfo)
	return result
}

func (e *Engine) Destroy() {
	e.renderer.Release()
}
