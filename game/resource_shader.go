package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertShader(assetShader asset.Shader) async.Promise[*graphics.Shader] {
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

func (s *ResourceSet) resolveShaderType(assetType asset.ShaderType) graphics.ShaderType {
	switch assetType {
	case asset.ShaderTypeGeometry:
		return graphics.ShaderTypeGeometry
	case asset.ShaderTypeShadow:
		return graphics.ShaderTypeShadow
	case asset.ShaderTypeForward:
		return graphics.ShaderTypeForward
	case asset.ShaderTypeSky:
		return graphics.ShaderTypeSky
	case asset.ShaderTypePostprocess:
		return graphics.ShaderTypePostprocess
	default:
		panic(fmt.Errorf("unsupported shader type: %d", assetType))
	}
}
