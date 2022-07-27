package buffer

import (
	"encoding/binary"
	"math"

	"github.com/mokiat/gomath/sprec"
)

// NewScanner creates a new Scanner instance over the specified
// byte slice that will do reading using the specified byte order.
func NewScanner(data []byte, order binary.ByteOrder) *Scanner {
	return &Scanner{
		data:  data,
		order: order,
	}
}

// Scanner is a wrapper over a byte slice that enables
// writing of various types of primitives.
type Scanner struct {
	data   []byte
	order  binary.ByteOrder
	offset int
}

// Data returns the underlying byte slice.
func (s *Scanner) Data() []byte {
	return s.data
}

// Rewind moves the write head back to the start of
// the slice.
func (s *Scanner) Rewind() {
	s.offset = 0
}

// Offset returns the location of the write head.
func (s *Scanner) Offset() int {
	return s.offset
}

// Seek changes the location of the write head.
func (s *Scanner) Seek(offset int) {
	s.offset = offset
}

// Skip moves the offset by the specified amount.
func (s *Scanner) Skip(offset int) {
	s.offset += offset
}

// ScanByte reads a single byte from the specified offset and then
// advances the offset.
func (s *Scanner) ScanByte() byte {
	result := s.data[s.offset]
	s.offset++
	return result
}

// ScanFloat32 reads a single float32 value from the current offset and
// advances the offset with four bytes.
func (s *Scanner) ScanFloat32() float32 {
	result := math.Float32frombits(s.order.Uint32(s.data[s.offset:]))
	s.offset += 4
	return result
}

// ScanMat4 reads a sprec.Mat4 value from the current offset and
// advances the offset with 64 bytes.
func (s *Scanner) ScanMat4() sprec.Mat4 {
	result := sprec.Mat4{
		M11: s.ScanFloat32(),
		M21: s.ScanFloat32(),
		M31: s.ScanFloat32(),
		M41: s.ScanFloat32(),

		M12: s.ScanFloat32(),
		M22: s.ScanFloat32(),
		M32: s.ScanFloat32(),
		M42: s.ScanFloat32(),

		M13: s.ScanFloat32(),
		M23: s.ScanFloat32(),
		M33: s.ScanFloat32(),
		M43: s.ScanFloat32(),

		M14: s.ScanFloat32(),
		M24: s.ScanFloat32(),
		M34: s.ScanFloat32(),
		M44: s.ScanFloat32(),
	}
	s.offset += 64
	return result
}
