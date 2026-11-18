package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/storage/chunked"
)

type ModelConverter struct {
	converters []dsl.Converter
}

var _ dsl.Converter = (*ModelConverter)(nil)

func NewModelConverter() *ModelConverter {
	return &ModelConverter{
		converters: []dsl.Converter{
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

func (c *ModelConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	for _, converter := range c.converters {
		if err := converter.Convert(target, asset); err != nil {
			return err
		}
	}
	return nil
}
