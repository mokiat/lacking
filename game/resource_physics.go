package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/physics"
)

func (s *ResourceSet) convertBody(bodyDefinitions IdentifiableList[*physics.BodyDefinition], assetBody dto.Body) bodyInstance {
	bodyDefinition := bodyDefinitions.GetByID(assetBody.BodyDefinitionID)
	return bodyInstance{
		NodeID:     assetBody.NodeID,
		Definition: bodyDefinition,
	}
}
