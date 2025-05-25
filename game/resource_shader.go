package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertShader(assetShader shadingdto.Shader) async.Promise[*graphics.Shader] {
	promise := async.NewPromise[*graphics.Shader]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		shader := gfxEngine.CreateShader(graphics.ShaderInfo{
			ShaderType: s.resolveShaderType(assetShader.ShaderType),
			SourceCode: assetShader.SourceCode,
		})
		promise.Deliver(shader)
	})
	return promise
}

func (s *ResourceSet) resolveShaderType(assetType shadingdto.ShaderType) graphics.ShaderType {
	switch assetType {
	case shadingdto.ShaderTypeGeometry:
		return graphics.ShaderTypeGeometry
	case shadingdto.ShaderTypeShadow:
		return graphics.ShaderTypeShadow
	case shadingdto.ShaderTypeForward:
		return graphics.ShaderTypeForward
	case shadingdto.ShaderTypeSky:
		return graphics.ShaderTypeSky
	case shadingdto.ShaderTypePostprocess:
		return graphics.ShaderTypePostprocess
	default:
		panic(fmt.Errorf("unsupported shader type: %d", assetType))
	}
}
