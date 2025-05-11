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
	return e.encodeValue(reflect.ValueOf(source))
}

func (e encoder) encodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct or pointer to struct but got %s", value.Kind())
	}

	fieldCount := value.NumField()
	if err := e.out.Encode(uint16(fieldCount)); err != nil {
		return fmt.Errorf("error writing chunk count: %w", err)
	}

	for i := range fieldCount {
		field := value.Field(i)
		if err := e.encodeField(field); err != nil {
			return fmt.Errorf("error encoding field %d: %w", i, err)
		}
	}

	// TODO: Handle preservation of unknown chunks through an interface API.

	return nil
}

func (e encoder) encodeField(field reflect.Value) error {
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return e.encodeChunk(nilChunk{})
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct or pointer to struct but got %s", field.Kind())
	}

	if !field.Type().Implements(chunkType) {
		return fmt.Errorf("expected chunk type but got %s", field.Type())
	}
	chunk := field.Interface().(Chunk)

	return e.encodeChunk(chunk)
}

func (e encoder) encodeChunk(chunk Chunk) error {
	size, err := e.measureChunkSize(chunk)
	if err != nil {
		return fmt.Errorf("error measuring chunk size: %w", err)
	}

	header := ChunkHeader{
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
