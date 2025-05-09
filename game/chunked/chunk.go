package chunked

import (
	"reflect"

	"github.com/google/uuid"
)

var (
	chunkType = reflect.TypeFor[Chunk]()
)

type Chunk interface {
	ChunkID() uuid.UUID
}

type ChunkHeader struct {
	ChunkID   uuid.UUID
	ChunkSize uint32
}

type UnknownChunk struct {
	ID   uuid.UUID
	Data []byte
}
