package storage

import (
	"io"
	"math"
)

type TypedReader interface {
	ReadBytes([]byte) error
	ReadBytesBlock() ([]byte, error)
	ReadByte() (byte, error)
	ReadBool() (bool, error)
	ReadUInt8() (uint8, error)
	ReadInt8() (int8, error)
	ReadUInt16() (uint16, error)
	ReadInt16() (int16, error)
	ReadUInt32() (uint32, error)
	ReadInt32() (int32, error)
	ReadUInt64() (uint64, error)
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
	buffer   []byte
}

func (r typedReader) ReadBytes(data []byte) error {
	_, err := io.ReadFull(r.delegate, data)
	return err
}

func (r typedReader) ReadBytesBlock() ([]byte, error) {
	count, err := r.ReadUInt64()
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
	return r.buffer[0], nil
}

func (r typedReader) ReadBool() (bool, error) {
	if err := r.readBuffer(1); err != nil {
		return false, err
	}
	return r.buffer[0] != 0, nil
}

func (r typedReader) ReadUInt8() (uint8, error) {
	return r.ReadByte()
}

func (r typedReader) ReadInt8() (int8, error) {
	value, err := r.ReadByte()
	return int8(value), err
}

func (r typedReader) ReadUInt16() (uint16, error) {
	if err := r.readBuffer(2); err != nil {
		return 0, err
	}
	value := uint16(r.buffer[0])
	value |= uint16(r.buffer[1]) << 8
	return value, nil
}

func (r typedReader) ReadInt16() (int16, error) {
	value, err := r.ReadUInt16()
	return int16(value), err
}

func (r typedReader) ReadUInt32() (uint32, error) {
	if err := r.readBuffer(4); err != nil {
		return 0, err
	}
	value := uint32(r.buffer[0])
	value |= uint32(r.buffer[1]) << 8
	value |= uint32(r.buffer[2]) << 16
	value |= uint32(r.buffer[3]) << 24
	return value, nil
}

func (r typedReader) ReadInt32() (int32, error) {
	value, err := r.ReadUInt32()
	return int32(value), err
}

func (r typedReader) ReadUInt64() (uint64, error) {
	if err := r.readBuffer(8); err != nil {
		return 0, err
	}
	value := uint64(r.buffer[0])
	value |= uint64(r.buffer[1]) << 8
	value |= uint64(r.buffer[2]) << 16
	value |= uint64(r.buffer[3]) << 24
	value |= uint64(r.buffer[4]) << 32
	value |= uint64(r.buffer[5]) << 40
	value |= uint64(r.buffer[6]) << 48
	value |= uint64(r.buffer[7]) << 56
	return value, nil
}

func (r typedReader) ReadInt64() (int64, error) {
	value, err := r.ReadUInt64()
	return int64(value), err
}

func (r typedReader) ReadFloat32() (float32, error) {
	bits, err := r.ReadUInt32()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(bits), nil
}

func (r typedReader) ReadFloat64() (float64, error) {
	bits, err := r.ReadUInt64()
	if err != nil {
		return 0.0, err
	}
	return math.Float64frombits(bits), nil
}

func (r typedReader) ReadString8() (string, error) {
	length, err := r.ReadUInt8()
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
	length, err := r.ReadUInt16()
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
	length, err := r.ReadUInt32()
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
