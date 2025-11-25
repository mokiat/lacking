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
	return e.encodeValue(value)
}

func (e encoder) encodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil // skipping nil values
		}
		value = value.Elem()
	}

	// Check if the value is a struct.
	if value.Kind() == reflect.Struct {
		if err := e.encodeStruct(value); err != nil {
			return fmt.Errorf("error encoding struct: %w", err)
		}
	}

	// Check if the value is a ChunkProvider.
	if value.Type().Implements(chunkProviderType) {
		provider, _ := reflect.TypeAssert[ChunkProvider](value)
		for _, chunk := range provider.Chunks() {
			if err := e.encodeChunk(chunk); err != nil {
				return fmt.Errorf("error encoding chunk: %w", err)
			}
		}
	}

	if err := e.encodeEOF(); err != nil {
		return fmt.Errorf("error encoding EOF: %w", err)
	}

	return nil
}

func (e encoder) encodeStruct(value reflect.Value) error {
	if (value.Kind() == reflect.Pointer) && value.IsNil() {
		return nil // skipping nil values
	}
	for i := range value.NumField() {
		typeField := value.Type().Field(i)
		if chunkID, ok := typeField.Tag.Lookup("chunk"); ok {
			field := value.Field(i)
			if err := e.encodeField(field, chunkID); err != nil {
				return fmt.Errorf("error encoding field: %w", err)
			}
		} else if typeField.Type.Kind() == reflect.Struct {
			field := value.Field(i)
			if err := e.encodeStruct(field); err != nil {
				return fmt.Errorf("error encoding nested struct: %w", err)
			}
		}
	}
	return nil
}

func (e encoder) encodeField(field reflect.Value, chunkID string) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		return nil // skipping nil pointer fields
	}
	content := field.Interface()

	size, err := e.measureContentSize(content)
	if err != nil {
		return fmt.Errorf("error measuring chunk size: %w", err)
	}

	header := chunkHeader{
		ChunkID:   chunkID,
		ChunkSize: size,
	}
	if err := e.out.Encode(header); err != nil {
		return fmt.Errorf("error encoding chunk header: %w", err)
	}

	if err := e.out.Encode(content); err != nil {
		return fmt.Errorf("error encoding chunk data: %w", err)
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

	if err := chunk.Encode(e.out); err != nil {
		return fmt.Errorf("error encoding chunk data: %w", err)
	}

	return nil
}

func (e encoder) encodeEOF() error {
	header := chunkHeader{
		ChunkID:   "",
		ChunkSize: 0,
	}
	if err := e.out.Encode(header); err != nil {
		return fmt.Errorf("error encoding chunk header: %w", err)
	}
	return nil
}

func (e encoder) measureContentSize(chunk any) (uint32, error) {
	counter := countedWriter{}
	encoder := gblob.NewLittleEndianPackedEncoder(&counter)
	if err := encoder.Encode(chunk); err != nil {
		return 0, fmt.Errorf("error encoding chunk: %w", err)
	}
	return counter.count, nil
}

func (e encoder) measureChunkSize(chunk Chunk) (uint32, error) {
	counter := countedWriter{}
	encoder := gblob.NewLittleEndianPackedEncoder(&counter)
	if err := chunk.Encode(encoder); err != nil {
		return 0, fmt.Errorf("error encoding chunk: %w", err)
	}
	return counter.count, nil
}
