package pack

import (
	"fmt"
	"log"
	"time"

	"github.com/mokiat/lacking/data/asset"
)

type Action interface {
	Run() error
}

type Described interface {
	Describe() string
}

func newPipeline(id int, resourceLocator ResourceLocator, assetLocator AssetLocator) *Pipeline {
	return &Pipeline{
		id:              id,
		resourceLocator: resourceLocator,
		assetLocator:    assetLocator,
	}
}

type Pipeline struct {
	id              int
	resourceLocator ResourceLocator
	assetLocator    AssetLocator
	actions         []Action
}

func (p *Pipeline) OpenImageResource(uri string) *OpenImageResourceAction {
	action := &OpenImageResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveTwoDTextureAsset(uri string, image ImageProvider) *SaveTwoDTextureAssetAction {
	action := &SaveTwoDTextureAssetAction{
		locator:       p.assetLocator,
		uri:           uri,
		imageProvider: image,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveCubeTextureAsset(uri string, image CubeImageProvider, opts ...SaveCubeTextureOption) *SaveCubeTextureAction {
	action := &SaveCubeTextureAction{
		locator:       p.assetLocator,
		uri:           uri,
		imageProvider: image,
		format:        asset.TexelFormatRGBA8,
	}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildCubeSideFromEquirectangular(side CubeSide, image ImageProvider) *BuildCubeSideFromEquirectangularAction {
	action := &BuildCubeSideFromEquirectangularAction{
		side:          side,
		imageProvider: image,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildCubeImage(opts ...BuildCubeImageOption) *BuildCubeImageAction {
	action := &BuildCubeImageAction{}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) ScaleCubeImage(image CubeImageProvider, dimension int) *ScaleCubeImageAction {
	action := &ScaleCubeImageAction{
		imageProvider: image,
		dimension:     dimension,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildIrradianceCubeImage(image CubeImageProvider, opts ...BuildIrradianceCubeImageOption) *BuildIrradianceCubeImageAction {
	action := &BuildIrradianceCubeImageAction{
		imageProvider: image,
		sampleCount:   10,
	}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) OpenGLTFResource(uri string) *OpenGLTFResourceAction {
	action := &OpenGLTFResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveModelAsset(uri string, model ModelProvider) *SaveModelAssetAction {
	action := &SaveModelAssetAction{
		locator:       p.assetLocator,
		uri:           uri,
		modelProvider: model,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) OpenLevelResource(uri string) *OpenLevelResourceAction {
	action := &OpenLevelResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveLevelAsset(uri string, level LevelProvider) *SaveLevelAssetAction {
	action := &SaveLevelAssetAction{
		locator:       p.assetLocator,
		uri:           uri,
		levelProvider: level,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) scheduleAction(action Action) {
	p.actions = append(p.actions, action)
}

func (p *Pipeline) execute() error {
	for _, action := range p.actions {
		described, isDescribed := action.(Described)

		startTime := time.Now()
		if err := action.Run(); err != nil {
			if isDescribed {
				return fmt.Errorf("action %q failed: %w", described.Describe(), err)
			}
			return fmt.Errorf("an action failed: %w", err)
		}
		elapsedTime := time.Since(startTime)

		if isDescribed {
			log.Printf("pipeline %d, action %s, complete in %s", p.id, described.Describe(), elapsedTime)
		}
	}
	return nil
}
