package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"golang.org/x/sync/errgroup"
)

func (l *AssetLoader) ResolveShader(assetShader dto.Shader) (Identifiable[*graphics.Shader], error) {
	var shader *graphics.Shader

	err := l.ScheduleMain(func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		shader = gfxEngine.CreateShader(graphics.ShaderInfo{
			ShaderType: l.resolveShaderType(assetShader.ShaderType),
			SourceCode: assetShader.SourceCode,
		})
		return nil
	}).Wait()

	return Identifiable[*graphics.Shader]{
		ID:    assetShader.ID,
		Value: shader,
	}, err
}

func (l *AssetLoader) ResolveShaders(assetShaders []dto.Shader) (IdentifiableList[*graphics.Shader], error) {
	shaders := make(IdentifiableList[*graphics.Shader], len(assetShaders))
	var group errgroup.Group
	for i, assetShader := range assetShaders {
		group.Go(func() error {
			shader, err := l.ResolveShader(assetShader)
			shaders[i] = shader
			return err
		})
	}
	return shaders, group.Wait()
}

func (l *AssetLoader) resolveShaderType(assetType dto.ShaderType) graphics.ShaderType {
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
