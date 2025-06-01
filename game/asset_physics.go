package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/physics"
	"golang.org/x/sync/errgroup"
)

func (l *AssetLoader) ResolvePhysicsMaterial(assetMaterial dto.BodyMaterial) (Identifiable[*physics.Material], error) {
	materialInfo := physics.MaterialInfo{
		FrictionCoefficient:    assetMaterial.FrictionCoefficient,
		RestitutionCoefficient: assetMaterial.RestitutionCoefficient,
	}

	var material *physics.Material
	allocateMaterial := func(engine *Engine) error {
		physicsEngine := engine.Physics()
		material = physicsEngine.CreateMaterial(materialInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateMaterial).Wait(); err != nil {
		return Identifiable[*physics.Material]{}, err
	}

	return Identifiable[*physics.Material]{
		ID:    assetMaterial.ID,
		Value: material,
	}, nil
}

func (l *AssetLoader) ResolvePhysicsMaterials(assetMaterials []dto.BodyMaterial) (IdentifiableList[*physics.Material], error) {
	materials := make(IdentifiableList[*physics.Material], len(assetMaterials))
	var group errgroup.Group
	for i, assetMaterial := range assetMaterials {
		group.Go(func() error {
			material, err := l.ResolvePhysicsMaterial(assetMaterial)
			materials[i] = material
			return err
		})
	}
	return materials, group.Wait()
}
