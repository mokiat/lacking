package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertShader(assetShader dto.Shader) async.Promise[*graphics.Shader] {
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

func (s *ResourceSet) resolveShaderType(assetType dto.ShaderType) graphics.ShaderType {
	switch assetType {
	case dto.ShaderTypeGeometry:
		return graphics.ShaderTypeGeometry
	case dto.ShaderTypeShadow:
		return graphics.ShaderTypeShadow
	case dto.ShaderTypeForward:
		return graphics.ShaderTypeForward
	case dto.ShaderTypeSky:
		return graphics.ShaderTypeSky
	case dto.ShaderTypePostprocess:
		return graphics.ShaderTypePostprocess
	default:
		panic(fmt.Errorf("unsupported shader type: %d", assetType))
	}
}
