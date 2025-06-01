package game

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"golang.org/x/sync/errgroup"
)

type ArmatureTemplate struct {
	Definition   *graphics.ArmatureDefinition
	NodeBindings []uint32
}

func (l *AssetLoader) ResolveArmatureTemplate(assetArmature dto.Armature) (Identifiable[ArmatureTemplate], error) {
	info := graphics.ArmatureDefinitionInfo{
		InverseBindMatrices: make([]sprec.Mat4, len(assetArmature.Joints)),
	}
	for i, assetJoint := range assetArmature.Joints {
		info.InverseBindMatrices[i] = assetJoint.InverseBindMatrix
	}
	var armatureDefinition *graphics.ArmatureDefinition
	allocateDefinition := func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		armatureDefinition = gfxEngine.CreateArmatureDefinition(info)
		return nil
	}
	if err := l.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[ArmatureTemplate]{}, err
	}

	nodeBindings := make([]uint32, len(assetArmature.Joints))
	for i, assetJoint := range assetArmature.Joints {
		nodeBindings[i] = assetJoint.NodeID
	}

	return Identifiable[ArmatureTemplate]{
		ID: assetArmature.ID,
		Value: ArmatureTemplate{
			Definition:   armatureDefinition,
			NodeBindings: nodeBindings,
		},
	}, nil
}

func (l *AssetLoader) ResolveArmatureTemplates(assetArmatures []dto.Armature) (IdentifiableList[ArmatureTemplate], error) {
	templates := make(IdentifiableList[ArmatureTemplate], len(assetArmatures))
	var group errgroup.Group
	for i, assetArmature := range assetArmatures {
		group.Go(func() error {
			template, err := l.ResolveArmatureTemplate(assetArmature)
			templates[i] = template
			return err
		})
	}
	return templates, group.Wait()
}

func (s *Scene) InstantiateArmatureTemplate(template ArmatureTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.Armature {
	armature := s.gfxScene.CreateArmature(graphics.ArmatureInfo{
		Definition: template.Definition,
	})
	for j, nodeID := range template.NodeBindings {
		if jointNode, ok := nodes.FindByID(nodeID); ok {
			jointNode.SetTarget(BoneNodeTarget{
				Armature:  armature,
				BoneIndex: j,
			})
		}
	}
	return armature
}
