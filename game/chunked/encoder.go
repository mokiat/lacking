package chunked

import (
	"fmt"
	"reflect"

	"github.com/mokiat/gblob"
)

type encoder struct {
	out *gblob.PackedEncoder
}

func (e encoder) Encode(source any) error {
	value := reflect.ValueOf(source)
	if err := e.encodeValue(value); err != nil {
		return fmt.Errorf("error encoding source: %w", err)
	}
	if err := e.encodeChunk(eofChunk{}); err != nil {
		return fmt.Errorf("error encoding EOF chunk: %w", err)
	}
	return nil
}

func (e encoder) encodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil // skipping nil values
		}
		value = value.Elem()
	}

	// Check if the value is a ChunkProvider.
	if value.Type().Implements(chunkProviderType) {
		provider := value.Interface().(ChunkProvider)
		for _, chunk := range provider.Chunks() {
			if err := e.encodeChunk(chunk); err != nil {
				return fmt.Errorf("error encoding chunk from provider: %w", err)
			}
		}
	}

	// Check if the value itself is a Chunk.
	if value.Type().Implements(chunkType) {
		chunk := value.Interface().(Chunk)
		if err := e.encodeChunk(chunk); err != nil {
			return fmt.Errorf("error encoding chunk: %w", err)
		}
	}

	// Check if the value is a struct.
	if value.Kind() == reflect.Struct {
		for i := range value.NumField() {
			field := value.Field(i)
			if err := e.encodeValue(field); err != nil {
				return fmt.Errorf("error encoding field %d: %w", i, err)
			}
		}
	}

	return nil
}

func (e encoder) encodeChunk(chunk Chunk) error {
	size, err := e.measureChunkSize(chunk)
	if err != nil {
		return fmt.Errorf("error measuring chunk size: %w", err)
	}

	header := chunkHeader{
		ChunkID:   chunk.ChunkID(),
		ChunkSize: size,
	}
	if err := e.out.Encode(header); err != nil {
		return fmt.Errorf("error encoding chunk header: %w", err)
	}

	if err := e.out.Encode(chunk); err != nil {
		return fmt.Errorf("error encoding chunk data: %w", err)
	}

	return nil
}

func (e encoder) measureChunkSize(chunk Chunk) (uint32, error) {
	counter := countedWriter{}
	encoder := gblob.NewLittleEndianPackedEncoder(&counter)
	if err := encoder.Encode(chunk); err != nil {
		return 0, fmt.Errorf("error encoding chunk: %w", err)
	}
	return counter.count, nil
}
