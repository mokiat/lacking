package chunked

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/mokiat/gblob"
)

type decoder struct {
	in *gblob.PackedDecoder
}

func (d decoder) Decode(target any) error {
	return d.decodeValue(reflect.ValueOf(target))
}

func (d decoder) decodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	} else {
		return fmt.Errorf("expected pointer to struct but got %s", value.Kind())
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct but got %s", value.Kind())
	}

	chunkIndices := make(map[uuid.UUID]int)
	for i := range value.NumField() {
		if id, ok := d.decodeChunkID(value.Field(i)); ok {
			chunkIndices[id] = i
		}
	}

	var chunkCount uint16
	if err := d.in.Decode(&chunkCount); err != nil {
		return fmt.Errorf("error reading chunk count: %w", err)
	}

	for range chunkCount {
		var header ChunkHeader
		if err := d.in.Decode(&header); err != nil {
			return fmt.Errorf("error reading chunk header: %w", err)
		}

		fieldIndex, ok := chunkIndices[header.ChunkID]
		if !ok {
			// TODO: Handle preservation of unknown chunks through an interface API.
			if err := d.skip(int(header.ChunkSize)); err != nil {
				return fmt.Errorf("error skipping chunk: %w", err)
			}
			continue
		}

		if err := d.decodeField(value.Field(fieldIndex)); err != nil {
			return fmt.Errorf("error decoding chunk: %w", err)
		}
	}

	return nil
}

func (d decoder) decodeChunkID(field reflect.Value) (uuid.UUID, bool) {
	if !field.Type().Implements(chunkType) {
		return uuid.Nil, false
	}
	if field.Kind() == reflect.Pointer && field.IsNil() {
		field = reflect.New(field.Type().Elem())
	}
	chunk := field.Interface().(Chunk)
	return chunk.ChunkID(), true
}

func (d decoder) skip(size int) error {
	destination := skipReader{
		count: size,
	}
	if err := d.in.Decode(&destination); err != nil {
		return err
	}
	return nil
}

func (d decoder) decodeField(field reflect.Value) error {
	if (field.Kind() == reflect.Pointer) && field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	target := field.Interface()
	if err := d.in.Decode(target); err != nil {
		return fmt.Errorf("error decoding field: %w", err)
	}
	return nil
}
