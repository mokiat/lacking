package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"golang.org/x/sync/errgroup"
)

// SkyTemplate represents a sky template that can be instantiated in a scene.
type SkyTemplate struct {
	NodeID     uint32
	Definition *graphics.SkyDefinition
}

// LoadSkyTemplate resolves a sky template from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadSkyTemplate(loader *AssetLoader, assetSky dto.Sky, materials IdentifiableList[*graphics.Material]) (Identifiable[SkyTemplate], error) {
	material, ok := materials.FindByID(assetSky.MaterialID)
	if !ok {
		return Identifiable[SkyTemplate]{}, fmt.Errorf("sky material with ID %d not found", assetSky.MaterialID)
	}

	skyDefinitionInfo := graphics.SkyDefinitionInfo{
		Material: material,
	}

	var skyDefinition *graphics.SkyDefinition
	allocateDefinition := func() error {
		gfxEngine := loader.Engine().Graphics()
		skyDefinition = gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
		return nil
	}
	if err := loader.ScheduleMain(allocateDefinition).Wait(); err != nil {
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

// LoadSkyTemplates resolves a list of sky templates from the given asset skies.
//
// This is a blocking operation and should be called from a worker thread.
func LoadSkyTemplates(loader *AssetLoader, assetSkies []dto.Sky, materials IdentifiableList[*graphics.Material]) (IdentifiableList[SkyTemplate], error) {
	var group errgroup.Group
	templates := make(IdentifiableList[SkyTemplate], len(assetSkies))
	for i, assetSky := range assetSkies {
		group.Go(func() error {
			template, err := LoadSkyTemplate(loader, assetSky, materials)
			templates[i] = template
			return err
		})
	}
	return templates, group.Wait()
}

// UnloadSkyTemplate unloads a sky template from the engine.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadSkyTemplate(loader *AssetLoader, idSky Identifiable[SkyTemplate]) error {
	sky := idSky.Value
	return loader.ScheduleMain(func() error {
		sky.Definition.Delete()
		return nil
	}).Wait()
}

// UnloadSkyTemplates unloads a list of sky templates from the engine.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadSkyTemplates(loader *AssetLoader, idSkies IdentifiableList[SkyTemplate]) error {
	var group errgroup.Group
	for _, idSky := range idSkies {
		group.Go(func() error {
			return UnloadSkyTemplate(loader, idSky)
		})
	}
	return group.Wait()
}

// InstantiateSkyTemplate instantiates a sky template in the given scene.
//
// This operation needs to be called from the main thread.
func InstantiateSkyTemplate(scene *Scene, template SkyTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.Sky {
	node := nodes.GetByID(template.NodeID)
	return scene.PlaceSky(node, SkyInfo{
		Definition: template.Definition,
	})
}
