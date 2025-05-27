package meshconv

import (
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/storage/chunked"
)

func init() {
	dsl.RegisterConverter(meshdto.MeshChunkID, &Converter{})
}

type Converter struct{}

func (c *Converter) CanConvert(asset dsl.Resource) bool {
	_, ok := asset.(Source)
	return ok
}

func (c *Converter) Convert(asset dsl.Resource) (chunked.Chunk, error) {
	src := asset.(Source)
	chunk, err := CreateMeshChunk(src)
	if err != nil {
		return nil, err
	}
	return chunked.FromValue(meshdto.MeshChunkID, chunk), nil
}
