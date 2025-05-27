package backgroundconv

import (
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	AllSkyPlacements() []mdl.Placed[*mdl.Sky]
}

func CreateBackgroundChunk(src Source) (*backgrounddto.BackgroundChunk, error) {
	allSkyPlacements := src.AllSkyPlacements()
	dtoSkies := make([]backgrounddto.Sky, len(allSkyPlacements))
	for i, placement := range allSkyPlacements {
		sky := placement.Value
		dtoSkies[i] = backgrounddto.Sky{
			ID:         sky.ID(),
			NodeID:     placement.Node.ID(),
			MaterialID: sky.Material().ID(),
		}
	}
	return &backgrounddto.BackgroundChunk{
		Skies: dtoSkies,
	}, nil
}
