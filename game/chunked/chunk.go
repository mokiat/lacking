package chunked

import (
	"reflect"

	"github.com/mokiat/gblob"
)

type Chunk interface {
	ChunkID() string
	Encode(out *gblob.PackedEncoder) error
	Decode(in *gblob.PackedDecoder) error
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

type ChunkHolder struct {
	Items []Chunk
}

func (h *ChunkHolder) AddChunk(chunk Chunk) {
	h.Items = append(h.Items, chunk)
}

func (h ChunkHolder) Chunks() []Chunk {
	return h.Items
}

func FromValue[T any](id string, value T) *ValueChunk[T] {
	return &ValueChunk[T]{
		ID:    id,
		Value: value,
	}
}

var _ Chunk = &ValueChunk[any]{}

type ValueChunk[T any] struct {
	ID    string
	Value T
}

func (c *ValueChunk[T]) ChunkID() string {
	return c.ID
}

func (c *ValueChunk[T]) Encode(out *gblob.PackedEncoder) error {
	return out.Encode(c.Value)
}

func (c *ValueChunk[T]) Decode(in *gblob.PackedDecoder) error {
	return in.Decode(&c.Value)
}

type RawChunk struct {
	ID   string
	Data RawData
}

func (c RawChunk) ChunkID() string {
	return c.ID
}

func (c RawChunk) Encode(out *gblob.PackedEncoder) error {
	return out.Encode(c.Data)
}

func (c RawChunk) Decode(in *gblob.PackedDecoder) error {
	return in.Decode(&c.Data)
}

type RawData []byte

var _ gblob.PackedEncodable = RawData{}
var _ gblob.PackedDecodable = RawData{}

func (c RawData) EncodePacked(writer gblob.TypedWriter) error {
	return writer.WriteBytes(c)
}

func (c RawData) DecodePacked(reader gblob.TypedReader) error {
	return reader.ReadBytes(c)
}

var (
	chunkProviderType = reflect.TypeFor[ChunkProvider]()
	chunkConsumerType = reflect.TypeFor[ChunkConsumer]()
)

type chunkHeader struct {
	ChunkID   string
	ChunkSize uint32
}
