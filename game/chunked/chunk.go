package chunked

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/mokiat/gblob"
)

var (
	chunkType         = reflect.TypeFor[Chunk]()
	chunkProviderType = reflect.TypeFor[ChunkProvider]()
	chunkConsumerType = reflect.TypeFor[ChunkConsumer]()
)

type Chunk interface {
	ChunkID() uuid.UUID
}

type ChunkProvider interface {
	Chunks() []Chunk
}

type ChunkConsumer interface {
	AddChunk(chunk Chunk)
}

type ChunkList []Chunk

func (l ChunkList) Chunks() []Chunk {
	return l
}

type BaseChunkHolder struct {
	Items []Chunk
}

func (h *BaseChunkHolder) AddChunk(chunk Chunk) {
	h.Items = append(h.Items, chunk)
}

func (h BaseChunkHolder) Chunks() []Chunk {
	return h.Items
}

type chunkHeader struct {
	ChunkID   uuid.UUID
	ChunkSize uint32
}

type eofChunk struct{}

func (c eofChunk) ChunkID() uuid.UUID {
	return uuid.Nil
}

type RawChunk struct {
	ID   uuid.UUID
	Data []byte
}

func (c RawChunk) ChunkID() uuid.UUID {
	return c.ID
}

var _ gblob.PackedEncodable = RawChunk{}
var _ gblob.PackedDecodable = &RawChunk{}

func (c RawChunk) EncodePacked(writer gblob.TypedWriter) error {
	return writer.WriteBytes(c.Data)
}

func (c *RawChunk) DecodePacked(reader gblob.TypedReader) error {
	return reader.ReadBytes(c.Data)
}
