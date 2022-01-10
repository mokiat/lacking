package storage

import (
	"io"
	"math"
)

type TypedWriter interface {
	WriteBytes([]byte) error
	WriteByteBlock(data []byte) error
	WriteByte(byte) error
	WriteBool(bool) error
	WriteInt8(int8) error
	WriteUInt8(uint8) error
	WriteInt16(int16) error
	WriteUInt16(uint16) error
	WriteInt32(int32) error
	WriteUInt32(uint32) error
	WriteUInt64(uint64) error
	WriteInt64(int64) error
	WriteFloat32(float32) error
	WriteFloat64(float64) error
	WriteString8(string) error
	WriteString16(string) error
	WriteString32(string) error
}

func NewTypedWriter(delegate io.Writer) TypedWriter {
	return typedWriter{
		delegate: delegate,
		buffer:   make([]byte, 8),
	}
}

type typedWriter struct {
	delegate io.Writer
	buffer   []byte
}

func (w typedWriter) WriteBytes(data []byte) error {
	if _, err := w.delegate.Write(data); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteByteBlock(data []byte) error {
	if err := w.WriteUInt64(uint64(len(data))); err != nil {
		return err
	}
	if _, err := w.delegate.Write(data); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteByte(value byte) error {
	w.buffer[0] = value
	return w.writeBuffer(1)
}

func (w typedWriter) WriteBool(value bool) error {
	if value {
		w.buffer[0] = 1
	} else {
		w.buffer[0] = 0
	}
	return w.writeBuffer(1)
}

func (w typedWriter) WriteInt8(value int8) error {
	return w.WriteByte(byte(value))
}

func (w typedWriter) WriteUInt8(value uint8) error {
	return w.WriteByte(byte(value))
}

func (w typedWriter) WriteUInt16(value uint16) error {
	w.buffer[0] = byte(value)
	w.buffer[1] = byte(value >> 8)
	return w.writeBuffer(2)
}

func (w typedWriter) WriteInt16(value int16) error {
	return w.WriteUInt16(uint16(value))
}

func (w typedWriter) WriteUInt32(value uint32) error {
	w.buffer[0] = byte(value)
	w.buffer[1] = byte(value >> 8)
	w.buffer[2] = byte(value >> 16)
	w.buffer[3] = byte(value >> 24)
	return w.writeBuffer(4)
}

func (w typedWriter) WriteInt32(value int32) error {
	return w.WriteUInt32(uint32(value))
}

func (w typedWriter) WriteUInt64(value uint64) error {
	w.buffer[0] = byte(value)
	w.buffer[1] = byte(value >> 8)
	w.buffer[2] = byte(value >> 16)
	w.buffer[3] = byte(value >> 24)
	w.buffer[4] = byte(value >> 32)
	w.buffer[5] = byte(value >> 40)
	w.buffer[6] = byte(value >> 48)
	w.buffer[7] = byte(value >> 56)
	return w.writeBuffer(8)
}

func (w typedWriter) WriteInt64(value int64) error {
	return w.WriteUInt64(uint64(value))
}

func (w typedWriter) WriteFloat32(value float32) error {
	bits := math.Float32bits(value)
	return w.WriteUInt32(bits)
}

func (w typedWriter) WriteFloat64(value float64) error {
	bits := math.Float64bits(value)
	return w.WriteUInt64(bits)
}

func (w typedWriter) WriteString8(value string) error {
	length := len(value)
	if length >= 256 {
		panic("Cannot fit string length in 8 bits")
	}
	if err := w.WriteUInt8(uint8(length)); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(value)); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteString16(value string) error {
	length := len(value)
	if length >= 256*256 {
		panic("Cannot fit string length in 16 bits")
	}
	if err := w.WriteUInt16(uint16(length)); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(value)); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteString32(value string) error {
	length := len(value)
	if err := w.WriteUInt32(uint32(length)); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(value)); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) writeBuffer(count int) error {
	return w.WriteBytes(w.buffer[:count])
}
