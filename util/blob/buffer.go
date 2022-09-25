package blob

import (
	"math"

	"github.com/x448/float16"
)

type Buffer []byte

func (b Buffer) Uint8(offset int) uint8 {
	return b[offset]
}

func (b Buffer) SetUint8(offset int, value uint8) {
	b[offset] = value
}

func (b Buffer) Uint16(offset int) uint16 {
	return uint16(b[offset+0])<<0 | uint16(b[offset+1])<<8
}

func (b Buffer) SetUint16(offset int, value uint16) {
	b[offset+0] = byte(value >> 0)
	b[offset+1] = byte(value >> 8)
}

func (b Buffer) Uint32(offset int) uint32 {
	return uint32(b[offset+0])<<0 | uint32(b[offset+1])<<8 | uint32(b[offset+2])<<16 | uint32(b[offset+3])<<24
}

func (b Buffer) SetUint32(offset int, value uint32) {
	b[offset+0] = byte(value >> 0)
	b[offset+1] = byte(value >> 8)
	b[offset+2] = byte(value >> 16)
	b[offset+3] = byte(value >> 24)
}

func (b Buffer) Uint64(offset int) uint64 {
	return uint64(b[offset+0])<<0 |
		uint64(b[offset+1])<<8 |
		uint64(b[offset+2])<<16 |
		uint64(b[offset+3])<<24 |
		uint64(b[offset+4])<<32 |
		uint64(b[offset+5])<<40 |
		uint64(b[offset+6])<<48 |
		uint64(b[offset+7])<<56
}

func (b Buffer) SetUint64(offset int, value uint64) {
	b[offset+0] = byte(value >> 0)
	b[offset+1] = byte(value >> 8)
	b[offset+2] = byte(value >> 16)
	b[offset+3] = byte(value >> 24)
	b[offset+4] = byte(value >> 32)
	b[offset+5] = byte(value >> 40)
	b[offset+6] = byte(value >> 48)
	b[offset+7] = byte(value >> 56)
}

func (b Buffer) Float16(offset int) float16.Float16 {
	return float16.Frombits(b.Uint16(offset))
}

func (b Buffer) SetFloat16(offset int, value float16.Float16) {
	b.SetUint16(offset, value.Bits())
}

func (b Buffer) Float32(offset int) float32 {
	return math.Float32frombits(b.Uint32(offset))
}

func (b Buffer) SetFloat32(offset int, value float32) {
	b.SetUint32(offset, math.Float32bits(value))
}

func (b Buffer) Float64(offset int) float64 {
	return math.Float64frombits(b.Uint64(offset))
}

func (b Buffer) SetFloat64(offset int, value float64) {
	b.SetUint64(offset, math.Float64bits(value))
}
