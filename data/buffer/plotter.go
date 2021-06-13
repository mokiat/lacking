package buffer

import (
	"encoding/binary"
	"math"
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

// PlotFloat32 sets a single float32 value at the current offset and
// advances the offset with four bytes.
func (p *Plotter) PlotFloat32(value float32) {
	p.order.PutUint32(p.data[p.offset:], math.Float32bits(value))
	p.offset += 4
}
