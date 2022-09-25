package blob

import (
	"fmt"
	"io"
)

type TypedWriter interface {
	WriteBytes([]byte) error
	WriteByteBlock(data []byte) error
	WriteByte(byte) error
	WriteBool(bool) error
	WriteInt8(int8) error
	WriteUint8(uint8) error
	WriteInt16(int16) error
	WriteUint16(uint16) error
	WriteInt32(int32) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error
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
	buffer   Buffer
}

func (w typedWriter) WriteBytes(data []byte) error {
	if _, err := w.delegate.Write(data); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteByteBlock(data []byte) error {
	if err := w.WriteUint64(uint64(len(data))); err != nil {
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
		w.buffer.SetUint8(0, 1)
	} else {
		w.buffer.SetUint8(0, 0)
	}
	return w.writeBuffer(1)
}

func (w typedWriter) WriteInt8(value int8) error {
	return w.WriteByte(byte(value))
}

func (w typedWriter) WriteUint8(value uint8) error {
	return w.WriteByte(byte(value))
}

func (w typedWriter) WriteUint16(value uint16) error {
	w.buffer.SetUint16(0, value)
	return w.writeBuffer(2)
}

func (w typedWriter) WriteInt16(value int16) error {
	return w.WriteUint16(uint16(value))
}

func (w typedWriter) WriteUint32(value uint32) error {
	w.buffer.SetUint32(0, value)
	return w.writeBuffer(4)
}

func (w typedWriter) WriteInt32(value int32) error {
	return w.WriteUint32(uint32(value))
}

func (w typedWriter) WriteUint64(value uint64) error {
	w.buffer.SetUint64(0, value)
	return w.writeBuffer(8)
}

func (w typedWriter) WriteInt64(value int64) error {
	return w.WriteUint64(uint64(value))
}

func (w typedWriter) WriteFloat32(value float32) error {
	w.buffer.SetFloat32(0, value)
	return w.writeBuffer(4)
}

func (w typedWriter) WriteFloat64(value float64) error {
	w.buffer.SetFloat64(0, value)
	return w.writeBuffer(8)
}

func (w typedWriter) WriteString8(value string) error {
	length := len(value)
	if length >= 0xFF {
		return fmt.Errorf("cannot fit string of length %d in 8 bits", length)
	}
	if err := w.WriteUint8(uint8(length)); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(value)); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteString16(value string) error {
	length := len(value)
	if length >= 0xFFFF {
		return fmt.Errorf("cannot fit string of length %d in 16 bits", length)
	}
	if err := w.WriteUint16(uint16(length)); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(value)); err != nil {
		return err
	}
	return nil
}

func (w typedWriter) WriteString32(value string) error {
	length := len(value)
	if length >= 0xFFFFFFFF {
		return fmt.Errorf("cannot fit string of length %d in 32 bits", length)
	}
	if err := w.WriteUint32(uint32(length)); err != nil {
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
