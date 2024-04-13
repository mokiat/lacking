package pack

import (
	"fmt"
	"time"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/util/resource"
)

type Action interface {
	Run() error
}

type Described interface {
	Describe() string
}

func newPipeline(id int, registry asset.Registry, resourceLocator resource.ReadLocator) *Pipeline {
	return &Pipeline{
		id:              id,
		registry:        registry,
		resourceLocator: resourceLocator,
	}
}

type Pipeline struct {
	id              int
	registry        asset.Registry
	resourceLocator resource.ReadLocator
	actions         []Action
}

func (p *Pipeline) OpenGLTFResource(uri string) *OpenGLTFResourceAction {
	action := &OpenGLTFResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveModelAsset(resource asset.Resource, model ModelProvider, opts ...SaveModelAssetOption) *SaveModelAssetAction {
	action := &SaveModelAssetAction{
		resource:      resource,
		modelProvider: model,
	}
	for _, opt := range opts {
		opt(action)
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

func (p *Pipeline) SaveLevelAsset(resource asset.Resource, level LevelProvider) *SaveLevelAssetAction {
	action := &SaveLevelAssetAction{
		resource:      resource,
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
			logger.Info("[ Pipeline %d ] %s - %s", p.id, described.Describe(), elapsedTime)
		}
	}
	return nil
}
