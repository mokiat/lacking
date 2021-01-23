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

func (w *Plotter) Rewind() {
	w.offset = 0
}

func (w *Plotter) Seek(offset int) {
	w.offset = offset
}

func (w *Plotter) PlotByte(value byte) {
	w.data[w.offset] = value
	w.offset++
}

func (w *Plotter) PlotUint16(value uint16) {
	w.order.PutUint16(w.data[w.offset:], value)
	w.offset += 2
}

func (w *Plotter) PlotUint32(value uint32) {
	w.order.PutUint32(w.data[w.offset:], value)
	w.offset += 4
}

func (w *Plotter) PlotFloat32(value float32) {
	w.order.PutUint32(w.data[w.offset:], math.Float32bits(value))
	w.offset += 4
}
