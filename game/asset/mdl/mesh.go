package mdl

import (
	"maps"
	"slices"

	"github.com/mokiat/gog"
)

func NewMeshDefinition() *MeshDefinition {
	return &MeshDefinition{
		Object:           NewObject(),
		materialBindings: make(map[string]*Material),
	}
}

type MeshDefinition struct {
	*Object
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
	return gog.Dedupe(slices.Collect(maps.Values(m.materialBindings)))
}

func (m *MeshDefinition) BindMaterial(name string, material *Material) {
	m.materialBindings[name] = material
}

func (m *MeshDefinition) MaterialBindings() map[string]*Material {
	return m.materialBindings
}

func NewMesh() *Mesh {
	return &Mesh{
		Object: NewObject(),
	}
}

type Mesh struct {
	*Object
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
