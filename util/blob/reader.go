package blob

import (
	"io"
)

type TypedReader interface {
	ReadBytes([]byte) error
	ReadBytesBlock() ([]byte, error)
	ReadByte() (byte, error)
	ReadBool() (bool, error)
	ReadUint8() (uint8, error)
	ReadInt8() (int8, error)
	ReadUint16() (uint16, error)
	ReadInt16() (int16, error)
	ReadUint32() (uint32, error)
	ReadInt32() (int32, error)
	ReadUint64() (uint64, error)
	ReadInt64() (int64, error)
	ReadFloat32() (float32, error)
	ReadFloat64() (float64, error)
	ReadString8() (string, error)
	ReadString16() (string, error)
	ReadString32() (string, error)
}

func NewTypedReader(delegate io.Reader) TypedReader {
	return &typedReader{
		delegate: delegate,
		buffer:   make([]byte, 8),
	}
}

type typedReader struct {
	delegate io.Reader
	buffer   Buffer
}

func (r typedReader) ReadBytes(data []byte) error {
	_, err := io.ReadFull(r.delegate, data)
	return err
}

func (r typedReader) ReadBytesBlock() ([]byte, error) {
	count, err := r.ReadUint64()
	if err != nil {
		return nil, err
	}

	data := make([]byte, count)
	if err := r.ReadBytes(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r typedReader) ReadByte() (byte, error) {
	if err := r.readBuffer(1); err != nil {
		return 0, err
	}
	return r.buffer.Uint8(0), nil
}

func (r typedReader) ReadBool() (bool, error) {
	b, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (r typedReader) ReadUint8() (uint8, error) {
	return r.ReadByte()
}

func (r typedReader) ReadInt8() (int8, error) {
	value, err := r.ReadByte()
	return int8(value), err
}

func (r typedReader) ReadUint16() (uint16, error) {
	if err := r.readBuffer(2); err != nil {
		return 0, err
	}
	return r.buffer.Uint16(0), nil
}

func (r typedReader) ReadInt16() (int16, error) {
	value, err := r.ReadUint16()
	return int16(value), err
}

func (r typedReader) ReadUint32() (uint32, error) {
	if err := r.readBuffer(4); err != nil {
		return 0, err
	}
	return r.buffer.Uint32(0), nil
}

func (r typedReader) ReadInt32() (int32, error) {
	value, err := r.ReadUint32()
	return int32(value), err
}

func (r typedReader) ReadUint64() (uint64, error) {
	if err := r.readBuffer(8); err != nil {
		return 0, err
	}
	return r.buffer.Uint64(0), nil
}

func (r typedReader) ReadInt64() (int64, error) {
	value, err := r.ReadUint64()
	return int64(value), err
}

func (r typedReader) ReadFloat32() (float32, error) {
	if err := r.readBuffer(4); err != nil {
		return 0.0, err
	}
	return r.buffer.Float32(0), nil
}

func (r typedReader) ReadFloat64() (float64, error) {
	if err := r.readBuffer(8); err != nil {
		return 0.0, err
	}
	return r.buffer.Float64(0), nil
}

func (r typedReader) ReadString8() (string, error) {
	length, err := r.ReadUint8()
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if err := r.ReadBytes(data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (r typedReader) ReadString16() (string, error) {
	length, err := r.ReadUint16()
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if err := r.ReadBytes(data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (r typedReader) ReadString32() (string, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if err := r.ReadBytes(data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (r typedReader) readBuffer(count int) error {
	return r.ReadBytes(r.buffer[:count])
}
