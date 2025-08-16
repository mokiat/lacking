package chunked

import (
	"fmt"
	"reflect"

	"github.com/mokiat/gblob"
)

type decoder struct {
	in *gblob.PackedDecoder
}

func (d decoder) Decode(target any) error {
	value := reflect.ValueOf(target)
	return d.decodeValue(value)
}

func (d decoder) decodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return nil // skipping nil pointers
	}

	chunkPlacements := make(map[string]reflect.Value)
	if value.Kind() == reflect.Pointer {
		derefValue := value.Elem()
		if derefValue.Kind() == reflect.Struct {
			d.exploreStruct(derefValue, chunkPlacements)
		}
	}

	var consumer ChunkConsumer
	if value.Type().Implements(chunkConsumerType) {
		consumer, _ = reflect.TypeAssert[ChunkConsumer](value)
	}

	for {
		var header chunkHeader
		if err := d.in.Decode(&header); err != nil {
			return fmt.Errorf("error reading chunk header: %w", err)
		}
		if header.ChunkID == "" {
			return nil // EOF chunk reached
		}

		if placement, ok := chunkPlacements[header.ChunkID]; ok {
			if (placement.Kind() == reflect.Pointer) && placement.IsNil() {
				placement.Set(reflect.New(placement.Type().Elem()))
			}
			target := placement.Interface()
			if err := d.in.Decode(target); err != nil {
				return fmt.Errorf("error decoding field: %w", err)
			}
		} else {
			if consumer != nil {
				target := RawChunk{
					ID:   header.ChunkID,
					Data: make([]byte, header.ChunkSize),
				}
				if err := target.Decode(d.in); err != nil {
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

func (d decoder) exploreStruct(value reflect.Value, chunkPlacements map[string]reflect.Value) {
	for i := range value.NumField() {
		typeField := value.Type().Field(i)
		if chunkID, ok := typeField.Tag.Lookup("chunk"); ok {
			field := value.Field(i)
			chunkPlacements[chunkID] = field
		} else if typeField.Type.Kind() == reflect.Struct {
			field := value.Field(i)
			d.exploreStruct(field, chunkPlacements)
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
