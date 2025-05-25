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
	value := reflect.ValueOf(target)
	if (value.Kind() == reflect.Pointer) && value.IsNil() {
		value.Set(reflect.New(value.Type().Elem()))
	}

	chunkValues := make(map[uuid.UUID]reflect.Value)
	consumer := d.exploreValue(value, chunkValues)
	return d.decodeChunks(chunkValues, consumer)
}

func (d decoder) exploreValue(value reflect.Value, chunkValues map[uuid.UUID]reflect.Value) ChunkConsumer {
	var consumer ChunkConsumer

	if value.Type().Implements(chunkConsumerType) {
		if value.Kind() != reflect.Pointer || !value.IsNil() {
			consumer = value.Interface().(ChunkConsumer)
		}
	}

	if value.Type().Implements(chunkType) {
		tempValue := value
		if tempValue.Kind() == reflect.Pointer && tempValue.IsNil() {
			tempValue = reflect.New(tempValue.Type().Elem())
		}
		chunk := tempValue.Interface().(Chunk)
		chunkValues[chunk.ChunkID()] = value
	}

	if (value.Kind() == reflect.Pointer) && !value.IsNil() {
		derefValue := value.Elem()
		if derefValue.Kind() == reflect.Struct {
			for i := range derefValue.NumField() {
				field := derefValue.Field(i)
				if cons := d.exploreValue(field, chunkValues); cons != nil && consumer == nil {
					consumer = cons
				}
			}
		}
	}

	return consumer
}

func (d decoder) decodeChunks(chunkValues map[uuid.UUID]reflect.Value, consumer ChunkConsumer) error {
	for {
		var header chunkHeader
		if err := d.in.Decode(&header); err != nil {
			return fmt.Errorf("error reading chunk header: %w", err)
		}
		if header.ChunkID == uuid.Nil {
			return nil // EOF chunk reached
		}

		if value, ok := chunkValues[header.ChunkID]; ok {
			if (value.Kind() == reflect.Pointer) && value.IsNil() {
				value.Set(reflect.New(value.Type().Elem()))
			}
			target := value.Interface()
			if err := d.in.Decode(target); err != nil {
				return fmt.Errorf("error decoding field: %w", err)
			}
		} else {
			if consumer != nil {
				target := &RawChunk{
					ID:   header.ChunkID,
					Data: make([]byte, header.ChunkSize),
				}
				if err := d.in.Decode(target); err != nil {
					return fmt.Errorf("error decoding raw chunk: %w", err)
				}
				consumer.AddChunk(target)
			} else {
				if err := d.skip(int(header.ChunkSize)); err != nil {
					return fmt.Errorf("error skipping chunk: %w", err)
				}
			}
		}
	}
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
