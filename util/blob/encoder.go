package blob

import (
	"fmt"
	"io"
	"reflect"
)

func NewReflectEncoder(out io.Writer) *ReflectEncoder {
	return &ReflectEncoder{
		out: NewTypedWriter(out),
	}
}

type ReflectEncoder struct {
	out TypedWriter
}

func (e *ReflectEncoder) Encode(source interface{}) error {
	return e.encodeValue(reflect.ValueOf(source))
}

func (e *ReflectEncoder) encodeValue(value reflect.Value) error {
	switch kind := value.Kind(); kind {
	case reflect.Pointer:
		return e.encodeValue(value.Elem())
	case reflect.Int8:
		if err := e.out.WriteInt8(int8(value.Int())); err != nil {
			return err
		}
		return nil
	case reflect.Uint8:
		if err := e.out.WriteUint8(uint8(value.Uint())); err != nil {
			return err
		}
		return nil
	case reflect.Int16:
		if err := e.out.WriteInt16(int16(value.Int())); err != nil {
			return err
		}
		return nil
	case reflect.Uint16:
		if err := e.out.WriteUint16(uint16(value.Uint())); err != nil {
			return err
		}
		return nil
	case reflect.Int32:
		if err := e.out.WriteInt32(int32(value.Int())); err != nil {
			return err
		}
		return nil
	case reflect.Uint32:
		if err := e.out.WriteUint32(uint32(value.Uint())); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", value.Kind())
	}
}
