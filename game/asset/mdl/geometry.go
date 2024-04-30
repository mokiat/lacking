package mdl

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset"
)

type Topology = asset.Topology

const (
	TopologyPoints        Topology = asset.TopologyPoints
	TopologyLineList      Topology = asset.TopologyLineList
	TopologyLineStrip     Topology = asset.TopologyLineStrip
	TopologyTriangleList  Topology = asset.TopologyTriangleList
	TopologyTriangleStrip Topology = asset.TopologyTriangleStrip
)

type VertexFormat uint8

const (
	VertexFormatCoord VertexFormat = 1 << iota
	VertexFormatNormal
	VertexFormatTangent
	VertexFormatTexCoord
	VertexFormatColor
	VertexFormatWeights
	VertexFormatJoints
)

type Geometry struct {
	name         string
	vertexFormat VertexFormat
	fragments    []*Fragment
}

func (g *Geometry) Name() string {
	return g.name
}

func (g *Geometry) SetName(name string) {
	g.name = name
}

func (g *Geometry) Format() VertexFormat {
	return g.vertexFormat
}

func (g *Geometry) SetFormat(format VertexFormat) {
	g.vertexFormat = format
}

func (g *Geometry) AddFragment(fragment *Fragment) {
	g.fragments = append(g.fragments, fragment)
}

type Fragment struct {
	name     string
	topology Topology
	vertices []Vertex
	indices  []int
}

func (f *Fragment) Name() string {
	return f.name
}

func (f *Fragment) SetName(name string) {
	f.name = name
}

func (f *Fragment) Topology() Topology {
	return f.topology
}

func (f *Fragment) SetTopology(topology Topology) {
	f.topology = topology
}

func (f *Fragment) VertexOffset() int {
	return len(f.vertices)
}

func (f *Fragment) AddVertex(vertex Vertex) {
	f.vertices = append(f.vertices, vertex)
}

func (f *Fragment) AddIndex(index int) {
	f.indices = append(f.indices, index)
}

type Vertex struct {
	Coord    dprec.Vec3
	Normal   dprec.Vec3
	Tangent  dprec.Vec3
	TexCoord dprec.Vec2
	Color    dprec.Vec4
	Weights  [4]float32
	Joints   [4]uint32
}

func (v Vertex) Translate(offset dprec.Vec3) Vertex {
	v.Coord = dprec.Vec3Sum(v.Coord, offset)
	return v
}

func (v Vertex) Rotate(rotation dprec.Quat) Vertex {
	v.Coord = dprec.QuatVec3Rotation(rotation, v.Coord)
	v.Normal = dprec.QuatVec3Rotation(rotation, v.Normal)
	v.Tangent = dprec.QuatVec3Rotation(rotation, v.Tangent)
	return v
}

func (v Vertex) Scale(factor dprec.Vec3) Vertex {
	v.Coord = dprec.Vec3{
		X: v.Coord.X * factor.X,
		Y: v.Coord.Y * factor.Y,
		Z: v.Coord.Z * factor.Z,
	}
	return v
}
