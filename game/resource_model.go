package game

import (
	"fmt"
	"reflect"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/storage/chunked"
)

func newModelResourceLoader() ResourceLoader[any] {
	return GenericResourceLoader(&modelResourceLoader{
		resourceType: reflect.TypeOf(gog.Zero[*ModelTemplate]()),
	})
}

type modelResourceLoader struct {
	resourceType reflect.Type
}

func (l *modelResourceLoader) ApplicableType() reflect.Type {
	return l.resourceType
}

func (l *modelResourceLoader) LoadResource(loader *AssetLoader, asset *chunked.Asset) (*ModelTemplate, error) {
	var dtoModel dto.Model
	if err := asset.Read(&dtoModel); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	model, err := LoadModelTemplate(loader, dtoModel)
	if err != nil {
		return nil, fmt.Errorf("failed to load model template: %w", err)
	}
	return model, nil
}

func (l *modelResourceLoader) UnloadResource(loader *AssetLoader, resource *ModelTemplate) error {
	if err := UnloadModelTemplate(loader, resource); err != nil {
		return fmt.Errorf("failed to unload model template: %w", err)
	}
	return nil
}
