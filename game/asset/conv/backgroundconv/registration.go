package backgroundconv

import (
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/storage/chunked"
)

func init() {
	dsl.RegisterConverter(backgrounddto.BackgroundChunkID, &Converter{})
}

type Converter struct{}

func (c *Converter) CanConvert(asset dsl.Resource) bool {
	_, ok := asset.(Source)
	return ok
}

func (c *Converter) Convert(asset dsl.Resource) (chunked.Chunk, error) {
	src := asset.(Source)
	chunk, err := CreateBackgroundChunk(src)
	if err != nil {
		return nil, err
	}
	return chunked.FromValue(backgrounddto.BackgroundChunkID, chunk), nil
}
