package mdl

import (
	"maps"

	"github.com/mokiat/gomath/sprec"
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

func NewGeometry() *Geometry {
	return &Geometry{
		minDistance: -32000.0,
		maxDistance: 32000.0,
		maxCascade:  255,
	}
}

type Geometry struct {
	name         string
	metadata     Metadata
	vertexFormat VertexFormat
	vertices     []Vertex
	indices      []int
	fragments    []*Fragment
	minDistance  float64
	maxDistance  float64
	maxCascade   int
}

func (g *Geometry) Metadata() Metadata {
	return g.metadata
}

func (g *Geometry) SetMetadata(metadata Metadata) {
	g.metadata = metadata
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

func (g *Geometry) VertexOffset() int {
	return len(g.vertices)
}

func (g *Geometry) Vertices() []Vertex {
	return g.vertices
}

func (g *Geometry) Vertex(at int) Vertex {
	return g.vertices[at]
}

func (g *Geometry) AddVertex(vertex Vertex) {
	g.vertices = append(g.vertices, vertex)
}

func (g *Geometry) IndexOffset() int {
	return len(g.indices)
}

func (g *Geometry) Indices() []int {
	return g.indices
}

func (g *Geometry) Index(at int) int {
	return g.indices[at]
}

func (g *Geometry) AddIndex(index int) {
	g.indices = append(g.indices, index)
}

func (g *Geometry) Fragments() []*Fragment {
	return g.fragments
}

func (g *Geometry) AddFragment(fragment *Fragment) {
	g.fragments = append(g.fragments, fragment)
}

func (g *Geometry) MinDistance() float64 {
	return g.minDistance
}

func (g *Geometry) SetMinDistance(distance float64) {
	g.minDistance = distance
}

func (g *Geometry) MaxDistance() float64 {
	return g.maxDistance
}

func (g *Geometry) SetMaxDistance(distance float64) {
	g.maxDistance = distance
}

func (g *Geometry) MaxCascade() int {
	return g.maxCascade
}

func (g *Geometry) SetMaxCascade(cascade int) {
	g.maxCascade = min(max(cascade, 0), 255)
}

func NewFragment() *Fragment {
	return &Fragment{}
}

type Fragment struct {
	name        string
	metadata    Metadata
	topology    Topology
	indexOffset int
	indexCount  int
}

func (f *Fragment) Metadata() Metadata {
	return f.metadata
}

func (f *Fragment) SetMetadata(metadata Metadata) {
	f.metadata = metadata
}

func (f *Fragment) AppendMetadata(metadata Metadata) {
	if f.metadata == nil {
		f.metadata = make(Metadata)
	}
	maps.Copy(f.metadata, metadata)
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

func (f *Fragment) IndexOffset() int {
	return f.indexOffset
}

func (f *Fragment) SetIndexOffset(offset int) {
	f.indexOffset = offset
}

func (f *Fragment) IndexCount() int {
	return f.indexCount
}

func (f *Fragment) SetIndexCount(count int) {
	f.indexCount = count
}

type Vertex struct {
	Coord    sprec.Vec3
	Normal   sprec.Vec3
	Tangent  sprec.Vec3
	TexCoord sprec.Vec2
	Color    sprec.Vec4
	Weights  sprec.Vec4
	Joints   [4]uint8
}

func (v Vertex) Translate(offset sprec.Vec3) Vertex {
	v.Coord = sprec.Vec3Sum(v.Coord, offset)
	return v
}

func (v Vertex) Rotate(rotation sprec.Quat) Vertex {
	v.Coord = sprec.QuatVec3Rotation(rotation, v.Coord)
	v.Normal = sprec.QuatVec3Rotation(rotation, v.Normal)
	v.Tangent = sprec.QuatVec3Rotation(rotation, v.Tangent)
	return v
}

func (v Vertex) Scale(factor sprec.Vec3) Vertex {
	v.Coord = sprec.Vec3{
		X: v.Coord.X * factor.X,
		Y: v.Coord.Y * factor.Y,
		Z: v.Coord.Z * factor.Z,
	}
	return v
}
