package data

import "math"

type Buffer []byte

func (b Buffer) SetFloat32(offset int, value float32) {
	bits := math.Float32bits(value)
	b[offset+0] = byte(bits >> 0)
	b[offset+1] = byte(bits >> 8)
	b[offset+2] = byte(bits >> 16)
	b[offset+3] = byte(bits >> 24)
}

func (b Buffer) Float32(offset int) float32 {
	value := uint32(b[offset+0])<<0 |
		uint32(b[offset+1])<<8 |
		uint32(b[offset+2])<<16 |
		uint32(b[offset+3])<<24
	return math.Float32frombits(value)
}

func (b Buffer) SetUInt16(offset int, value uint16) {
	b[offset+0] = byte(value >> 0)
	b[offset+1] = byte(value >> 8)
}

func (b Buffer) UInt16(offset int) uint16 {
	return uint16(b[offset+0])<<0 | uint16(b[offset+1])<<8
}
