package buffer

import (
	"encoding/binary"
	"math"
)

func NewPlotter(data []byte, order binary.ByteOrder) *Plotter {
	return &Plotter{
		data:  data,
		order: order,
	}
}

type Plotter struct {
	data   []byte
	order  binary.ByteOrder
	offset int
}

func (p *Plotter) Data() []byte {
	return p.data
}

func (p *Plotter) Rewind() {
	p.offset = 0
}

func (p *Plotter) Offset() int {
	return p.offset
}

func (p *Plotter) Seek(offset int) {
	p.offset = offset
}

func (p *Plotter) PlotByte(value byte) {
	p.data[p.offset] = value
	p.offset++
}

func (p *Plotter) PlotUint16(value uint16) {
	p.order.PutUint16(p.data[p.offset:], value)
	p.offset += 2
}

func (p *Plotter) PlotUint32(value uint32) {
	p.order.PutUint32(p.data[p.offset:], value)
	p.offset += 4
}

func (p *Plotter) PlotFloat32(value float32) {
	p.order.PutUint32(p.data[p.offset:], math.Float32bits(value))
	p.offset += 4
}
