package game

import (
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertGeometryShader(assetShader asset.Shader) async.Promise[*graphics.GeometryShader] {
	promise := async.NewPromise[*graphics.GeometryShader]()
	s.gfxWorker.ScheduleVoid(func() {
		gfxEngine := s.engine.Graphics()
		shader := gfxEngine.CreateGeometryShader(graphics.ShaderInfo{
			SourceCode: assetShader.SourceCode,
		})
		promise.Deliver(shader)
	})
	return promise
}

func (s *ResourceSet) convertShadowShader(assetShader asset.Shader) async.Promise[*graphics.ShadowShader] {
	promise := async.NewPromise[*graphics.ShadowShader]()
	s.gfxWorker.ScheduleVoid(func() {
		gfxEngine := s.engine.Graphics()
		shader := gfxEngine.CreateShadowShader(graphics.ShaderInfo{
			SourceCode: assetShader.SourceCode,
		})
		promise.Deliver(shader)
	})
	return promise
}

func (s *ResourceSet) convertForwardShader(assetShader asset.Shader) async.Promise[*graphics.ForwardShader] {
	promise := async.NewPromise[*graphics.ForwardShader]()
	s.gfxWorker.ScheduleVoid(func() {
		gfxEngine := s.engine.Graphics()
		shader := gfxEngine.CreateForwardShader(graphics.ShaderInfo{
			SourceCode: assetShader.SourceCode,
		})
		promise.Deliver(shader)
	})
	return promise
}

func (s *ResourceSet) convertSkyShader(assetShader asset.Shader) async.Promise[*graphics.SkyShader] {
	promise := async.NewPromise[*graphics.SkyShader]()
	s.gfxWorker.ScheduleVoid(func() {
		gfxEngine := s.engine.Graphics()
		shader := gfxEngine.CreateSkyShader(graphics.ShaderInfo{
			SourceCode: assetShader.SourceCode,
		})
		promise.Deliver(shader)
	})
	return promise
}
