package physicsconv

import (
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/storage/chunked"
)

func init() {
	dsl.RegisterConverter(physicsdto.PhysicsChunkID, &Converter{})
}

type Converter struct{}

func (c *Converter) CanConvert(asset dsl.Resource) bool {
	_, ok := asset.(Source)
	return ok
}

func (c *Converter) Convert(asset dsl.Resource) (chunked.Chunk, error) {
	src := asset.(Source)
	chunk, err := CreatePhysicsChunk(src)
	if err != nil {
		return nil, err
	}
	return chunked.FromValue(physicsdto.PhysicsChunkID, chunk), nil
}
