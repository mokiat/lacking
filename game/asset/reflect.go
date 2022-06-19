package asset

import (
	"fmt"
	"io"
	"reflect"

	"github.com/mokiat/lacking/data/storage"
)

func NewReflectEncoder(out io.Writer) *ReflectEncoder {
	return &ReflectEncoder{
		out: storage.NewTypedWriter(out),
	}
}

type ReflectEncoder struct {
	out storage.TypedWriter
}

func Encode(value interface{}) error {
	return nil
}

func NewReflectDecoder(in io.Reader) *ReflectDecoder {
	return &ReflectDecoder{
		in: storage.NewTypedReader(in),
	}
}

type ReflectDecoder struct {
	in storage.TypedReader
}

func (d *ReflectDecoder) Decode(target interface{}) error {
	reflValue := reflect.ValueOf(target)
	if reflValue.Kind() != reflect.Pointer {
		return fmt.Errorf("target %T is not a pointer", target)
	}
	if reflValue.IsNil() {
		return fmt.Errorf("target is a nil pointer")
	}
	reflElem := reflValue.Elem()
	if reflElem.Kind() != reflect.Struct {
		return fmt.Errorf("target %T does not point to struct", target)
	}
	reflFieldCount := reflElem.NumField()
	for i := 0; i < reflFieldCount; i++ {
		reflField := reflElem.Field(i)
		if err := d.decodeField(reflField); err != nil {
			return err
		}
	}
	return nil
}

func (d *ReflectDecoder) decodeField(reflField reflect.Value) error {
	switch reflField.Kind() {
	case reflect.Uint8:
		v, err := d.in.ReadByte()
		if err != nil {
			return err
		}
		reflField.Set(reflect.ValueOf(v).Convert(reflField.Type()))
		return nil

	case reflect.Uint16:
		v, err := d.in.ReadUInt16()
		if err != nil {
			return err
		}
		reflField.Set(reflect.ValueOf(v).Convert(reflField.Type()))
		return nil

	case reflect.Struct:
		reflFieldCount := reflField.NumField()
		for i := 0; i < reflFieldCount; i++ {
			reflField := reflField.Field(i)
			if err := d.decodeField(reflField); err != nil {
				return err
			}
		}
		return nil

	case reflect.Slice:
		reflElemKind := reflField.Type().Elem().Kind()
		switch reflElemKind {
		case reflect.Uint8:
			data, err := d.in.ReadBytesBlock()
			if err != nil {
				return err
			}
			reflField.Set(reflect.ValueOf(data).Convert(reflField.Type()))
			return nil

		default:
			return fmt.Errorf("unsupported slice element kind %s", reflElemKind)
		}

	default:
		return fmt.Errorf("unsupported field kind  %s", reflField.Kind())
	}
}
