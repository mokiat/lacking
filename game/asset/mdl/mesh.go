package mdl

import (
	"github.com/mokiat/gog"
	"golang.org/x/exp/maps"
)

type MeshDefinition struct {
	name             string
	geometry         *Geometry
	materialBindings map[string]*Material
}

func (m *MeshDefinition) Name() string {
	return m.name
}

func (m *MeshDefinition) SetName(name string) {
	m.name = name
}

func (m *MeshDefinition) Geometry() *Geometry {
	return m.geometry
}

func (m *MeshDefinition) SetGeometry(geometry *Geometry) {
	m.geometry = geometry
}

func (m *MeshDefinition) Materials() []*Material {
	return gog.Dedupe(maps.Values(m.materialBindings))
}

func (m *MeshDefinition) BindMaterial(name string, material *Material) {
	if m.materialBindings == nil {
		m.materialBindings = make(map[string]*Material)
	}
	m.materialBindings[name] = material
}

func NewMesh() *Mesh {
	return &Mesh{}
}

type Mesh struct {
	definition *MeshDefinition
	armature   *Armature
}

func (m *Mesh) Definition() *MeshDefinition {
	return m.definition
}

func (m *Mesh) SetDefinition(definition *MeshDefinition) {
	m.definition = definition
}

func (m *Mesh) Armature() *Armature {
	return m.armature
}

func (m *Mesh) SetArmature(armature *Armature) {
	m.armature = armature
}
