package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type BackgroundSource interface {
	AllSkyPlacements() []mdl.Placed[*mdl.Sky]
}

func NewBackgroundConverter() *BackgroundConverter {
	return &BackgroundConverter{}
}

type BackgroundConverter struct{}

func (c *BackgroundConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(BackgroundSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateBackgroundChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.BackgroundChunkID, chunk))
	return nil
}

func (c *BackgroundConverter) CreateBackgroundChunk(src BackgroundSource) (*dto.BackgroundChunk, error) {
	allSkyPlacements := src.AllSkyPlacements()
	dtoSkies := make([]dto.Sky, len(allSkyPlacements))
	for i, placement := range allSkyPlacements {
		sky := placement.Value
		dtoSkies[i] = dto.Sky{
			ID:         sky.ID(),
			NodeID:     placement.Node.ID(),
			MaterialID: sky.Material().ID(),
		}
	}
	return &dto.BackgroundChunk{
		Skies: dtoSkies,
	}, nil
}
