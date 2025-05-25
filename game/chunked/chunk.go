package chunked

import (
	"reflect"

	"github.com/google/uuid"
)

var (
	chunkType         = reflect.TypeFor[Chunk]()
	chunkProviderType = reflect.TypeFor[ChunkProvider]()
)

type Chunk interface {
	ChunkID() uuid.UUID
}

type ChunkProvider interface {
	Chunks() []Chunk
}

type ChunkList []Chunk

func (l ChunkList) Chunks() []Chunk {
	return l
}

type chunkHeader struct {
	ChunkID   uuid.UUID
	ChunkSize uint32
}

type eofChunk struct{}

func (c eofChunk) ChunkID() uuid.UUID {
	return uuid.Nil
}

// type UnknownChunk struct {
// 	ID   uuid.UUID
// 	Data []byte
// }

// func (c UnknownChunk) ChunkID() uuid.UUID {
// 	return c.ID
// }
