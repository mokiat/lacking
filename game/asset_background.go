package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"golang.org/x/sync/errgroup"
)

type SkyTemplate struct {
	NodeID     uint32
	Definition *graphics.SkyDefinition
}

func (l *AssetLoader) ResolveSkyTemplate(assetSky dto.Sky, materials IdentifiableList[*graphics.Material]) (Identifiable[SkyTemplate], error) {
	material, ok := materials.FindByID(assetSky.MaterialID)
	if !ok {
		return Identifiable[SkyTemplate]{}, fmt.Errorf("sky material with ID %d not found", assetSky.MaterialID)
	}

	skyDefinitionInfo := graphics.SkyDefinitionInfo{
		Material: material,
	}

	var skyDefinition *graphics.SkyDefinition
	allocateDefinition := func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		skyDefinition = gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[SkyTemplate]{}, err
	}

	return Identifiable[SkyTemplate]{
		ID: assetSky.ID,
		Value: SkyTemplate{
			NodeID:     assetSky.NodeID,
			Definition: skyDefinition,
		},
	}, nil
}

func (l *AssetLoader) ResolveSkyTemplates(assetSkies []dto.Sky, materials IdentifiableList[*graphics.Material]) (IdentifiableList[SkyTemplate], error) {
	var group errgroup.Group
	templates := make(IdentifiableList[SkyTemplate], len(assetSkies))
	for i, assetSky := range assetSkies {
		group.Go(func() error {
			template, err := l.ResolveSkyTemplate(assetSky, materials)
			templates[i] = template
			return err
		})
	}
	return templates, group.Wait()
}

func (s *Scene) InstantiateSkyTemplate(template SkyTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.Sky {
	node := nodes.GetByID(template.NodeID)
	return s.PlaceSky(node, SkyInfo{
		Definition: template.Definition,
	})
}
