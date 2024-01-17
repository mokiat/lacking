package blob

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/sprec"
	"github.com/x448/float16"
)

// NewPlotter creates a new Plotter instance over the specified byte slice.
func NewPlotter(data []byte) *Plotter {
	return &Plotter{
		data: data,
	}
}

// Plotter is a wrapper over a byte slice that enables writing of various types
// of primitives in little-endian order.
type Plotter struct {
	data   gblob.LittleEndianBlock
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

// PlotBytes copies the specified data at the current offset and
// advances the offset with the length of the data.
func (p *Plotter) PlotBytes(data []byte) {
	copy(p.data[p.offset:], data)
	p.offset += len(data)
}

// PlotUint8 sets a single byte at the current offset
// and advances the offset with 1 byte.
func (p *Plotter) PlotUint8(value byte) {
	p.data.SetUint8(p.offset, value)
	p.offset++
}

// PlotUint16 sets a single uint16 value at the current offset and
// advances the offset with 2 bytes.
func (p *Plotter) PlotUint16(value uint16) {
	p.data.SetUint16(p.offset, value)
	p.offset += 2
}

// PlotUint32 sets a single uint32 value at the current offset and
// advances the offset with 4 bytes.
func (p *Plotter) PlotUint32(value uint32) {
	p.data.SetUint32(p.offset, value)
	p.offset += 4
}

// PlotUint64 sets a single uint64 value at the current offset and
// advances the offset with 8 bytes.
func (p *Plotter) PlotUint64(value uint64) {
	p.data.SetUint64(p.offset, value)
	p.offset += 8
}

// PlotFloat16 sets a single float16 value at the current offset and
// advances the offset with 2 bytes.
func (p *Plotter) PlotFloat16(value float16.Float16) {
	p.data.SetUint16(p.offset, value.Bits())
	p.offset += 2
}

// PlotFloat32 sets a single float32 value at the current offset and
// advances the offset with 4 bytes.
func (p *Plotter) PlotFloat32(value float32) {
	p.data.SetFloat32(p.offset, value)
	p.offset += 4
}

// PlotFloat64 sets a single float64 value at the current offset and
// advances the offset with 8 bytes.
func (p *Plotter) PlotFloat64(value float64) {
	p.data.SetFloat64(p.offset, value)
	p.offset += 8
}

// PlotSPVec2 sets a sprec.Vec2 value at the current offset and
// advances the offset with 8 bytes.
func (p *Plotter) PlotSPVec2(value sprec.Vec2) {
	p.PlotFloat32(value.X)
	p.PlotFloat32(value.Y)
}

// PlotSPVec3 sets a sprec.Vec3 value at the current offset and
// advances the offset with 12 bytes.
func (p *Plotter) PlotSPVec3(value sprec.Vec3) {
	p.PlotFloat32(value.X)
	p.PlotFloat32(value.Y)
	p.PlotFloat32(value.Z)
}

// PlotSPVec4 sets a sprec.Vec4 value at the current offset and
// advances the offset with 16 bytes.
func (p *Plotter) PlotSPVec4(value sprec.Vec4) {
	p.PlotFloat32(value.X)
	p.PlotFloat32(value.Y)
	p.PlotFloat32(value.Z)
	p.PlotFloat32(value.W)
}

// PlotSPMat4 sets a sprec.Mat4 value at the current offset and
// advances the offset with 64 bytes.
func (p *Plotter) PlotSPMat4(value sprec.Mat4) {
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
