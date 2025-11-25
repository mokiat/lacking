package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/storage/chunked"
)

type Converter interface {
	Convert(target *ds.List[chunked.Chunk], asset any) error
}

func NewModelConverter() *ModelConverter {
	return &ModelConverter{
		converters: []Converter{
			NewAnimationConverter(),
			NewBackgroundConverter(),
			NewHierarchyConverter(),
			NewLightingConverter(),
			NewMeshConverter(),
			NewPhysicsConverter(),
			NewShadingConverter(),
		},
	}
}

type ModelConverter struct {
	converters []Converter
}

func (c *ModelConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	for _, converter := range c.converters {
		if err := converter.Convert(target, asset); err != nil {
			return err
		}
	}
	return nil
}
