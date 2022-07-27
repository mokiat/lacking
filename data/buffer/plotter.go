package buffer

import (
	"encoding/binary"
	"math"

	"github.com/mokiat/gomath/sprec"
	"github.com/x448/float16"
)

// NewPlotter creates a new Plotter instance over the specified
// byte slice that will do writing using the specified byte order.
func NewPlotter(data []byte, order binary.ByteOrder) *Plotter {
	return &Plotter{
		data:  data,
		order: order,
	}
}

// Plotter is a wrapper over a byte slice that enables
// writing of various types of primitives.
type Plotter struct {
	data   []byte
	order  binary.ByteOrder
	offset int
}

// Data returns the underlying byte slice.
func (p *Plotter) Data() []byte {
	return p.data
}

// Rewind moves the write head back to the start of
// the slice.
func (p *Plotter) Rewind() {
	p.offset = 0
}

// Offset returns the location of the write head.
func (p *Plotter) Offset() int {
	return p.offset
}

// Seek changes the location of the write head.
func (p *Plotter) Seek(offset int) {
	p.offset = offset
}

// Skip moves the offset by the specified amount.
func (p *Plotter) Skip(offset int) {
	p.offset += offset
}

// PlotByte sets a single byte at the current offset
// and advances the offset with one byte.
func (p *Plotter) PlotByte(value byte) {
	p.data[p.offset] = value
	p.offset++
}

// PlotUint16 sets a single uint16 value at the current offset and
// advances the offset with two bytes.
func (p *Plotter) PlotUint16(value uint16) {
	p.order.PutUint16(p.data[p.offset:], value)
	p.offset += 2
}

// PlotUint32 sets a single uint32 value at the current offset and
// advances the offset with four bytes.
func (p *Plotter) PlotUint32(value uint32) {
	p.order.PutUint32(p.data[p.offset:], value)
	p.offset += 4
}

// PlotFloat16 sets a single float16 value at the current offset and
// advances the offset with two bytes.
func (p *Plotter) PlotFloat16(value float16.Float16) {
	p.order.PutUint16(p.data[p.offset:], value.Bits())
	p.offset += 2
}

// PlotFloat32 sets a single float32 value at the current offset and
// advances the offset with four bytes.
func (p *Plotter) PlotFloat32(value float32) {
	p.order.PutUint32(p.data[p.offset:], math.Float32bits(value))
	p.offset += 4
}

// PlotVec4 sets a sprec.Vec4 value at the current offset and
// advances the offset with 16 bytes.
func (p *Plotter) PlotVec4(value sprec.Vec4) {
	p.PlotFloat32(value.X)
	p.PlotFloat32(value.Y)
	p.PlotFloat32(value.Z)
	p.PlotFloat32(value.W)
}

// PlotMat4 sets a sprec.Mat4 value at the current offset and
// advances the offset with 64 bytes.
func (p *Plotter) PlotMat4(value sprec.Mat4) {
	p.PlotFloat32(value.M11)
	p.PlotFloat32(value.M21)
	p.PlotFloat32(value.M31)
	p.PlotFloat32(value.M41)

	p.PlotFloat32(value.M12)
	p.PlotFloat32(value.M22)
	p.PlotFloat32(value.M32)
	p.PlotFloat32(value.M42)

	p.PlotFloat32(value.M13)
	p.PlotFloat32(value.M23)
	p.PlotFloat32(value.M33)
	p.PlotFloat32(value.M43)

	p.PlotFloat32(value.M14)
	p.PlotFloat32(value.M24)
	p.PlotFloat32(value.M34)
	p.PlotFloat32(value.M44)
}
