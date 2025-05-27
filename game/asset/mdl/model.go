package mdl

import (
	"iter"
	"slices"

	"github.com/mokiat/gog"
)

func NewModel() *Model {
	return &Model{
		Object: NewObject(),
	}
}

type Model struct {
	*Object
	name string

	nodes      []*Node
	animations []*Animation
}

func (s *Model) Name() string {
	return s.name
}

func (s *Model) SetName(name string) {
	s.name = name
}

func (s *Model) Nodes() []*Node {
	return s.nodes
}

func (s *Model) AddNode(node *Node) {
	s.nodes = append(s.nodes, node)
}

func (s *Model) RemoveNode(node *Node) {
	s.nodes = slices.DeleteFunc(s.nodes, func(candidate *Node) bool {
		return candidate == node
	})
}

func (s *Model) NodesIter() iter.Seq2[int, *Node] {
	return indexedIter(func(yield func(*Node) bool) {
		s.yieldNodes(s.nodes, yield)
	})
}

func (s *Model) yieldNodes(nodes []*Node, yield func(*Node) bool) bool {
	for _, node := range nodes {
		if !yield(node) {
			return false
		}
		if !s.yieldNodes(node.Nodes(), yield) {
			return false
		}
	}
	return true
}

func (s *Model) Animations() []*Animation {
	return s.animations
}

func (s *Model) AddAnimation(animation *Animation) {
	s.animations = append(s.animations, animation)
}

func (s *Model) AllShaders() []*Shader {
	var result []*Shader
	for _, material := range s.AllMaterials() {
		for _, pass := range material.AllPasses() {
			result = append(result, pass.Shader())
		}
	}
	return gog.Dedupe(result)
}

func (s *Model) AllTextures() []*Texture {
	var result []*Texture
	for _, placement := range s.AllAmbientLightPlacements() {
		light := placement.Value
		result = append(result, light.ReflectionTexture())
		result = append(result, light.RefractionTexture())
	}
	for _, material := range s.AllMaterials() {
		for _, sampler := range material.Samplers() {
			result = append(result, sampler.Texture())
		}
	}
	return gog.Dedupe(result)
}

func (s *Model) AllMaterials() []*Material {
	var result []*Material
	for _, placement := range s.AllMeshPlacements() {
		mesh := placement.Value
		definition := mesh.Definition()
		result = append(result, definition.Materials()...)
	}
	for _, placement := range s.AllSkyPlacements() {
		sky := placement.Value
		result = append(result, sky.Material())
	}
	return gog.Dedupe(result)
}

func (s *Model) AllAmbientLightPlacements() []Placed[*AmbientLight] {
	var result []Placed[*AmbientLight]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *AmbientLight:
			result = append(result, Placed[*AmbientLight]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllPointLightPlacements() []Placed[*PointLight] {
	var result []Placed[*PointLight]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *PointLight:
			result = append(result, Placed[*PointLight]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllSpotLightPlacements() []Placed[*SpotLight] {
	var result []Placed[*SpotLight]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *SpotLight:
			result = append(result, Placed[*SpotLight]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllDirectionalLightPlacements() []Placed[*DirectionalLight] {
	var result []Placed[*DirectionalLight]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *DirectionalLight:
			result = append(result, Placed[*DirectionalLight]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllArmatures() []*Armature {
	var result []*Armature
	for _, placement := range s.AllMeshPlacements() {
		mesh := placement.Value
		if armature := mesh.Armature(); armature != nil {
			result = append(result, armature)
		}
	}
	return gog.Dedupe(result)
}

func (s *Model) AllGeometries() []*Geometry {
	var result []*Geometry
	for _, definition := range s.AllMeshDefinitions() {
		result = append(result, definition.Geometry())
	}
	return gog.Dedupe(result)
}

func (s *Model) AllMeshDefinitions() []*MeshDefinition {
	var result []*MeshDefinition
	for _, placement := range s.AllMeshPlacements() {
		mesh := placement.Value
		definition := mesh.Definition()
		result = append(result, definition)
	}
	return gog.Dedupe(result)
}

func (s *Model) AllMeshPlacements() []Placed[*Mesh] {
	var result []Placed[*Mesh]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *Mesh:
			result = append(result, Placed[*Mesh]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllPhysicsBodyMaterials() []*BodyMaterial {
	var result []*BodyMaterial
	for _, definition := range s.AllPhysicsBodyDefinitions() {
		material := definition.Material()
		result = append(result, material)
	}
	return gog.Dedupe(result)
}

func (s *Model) AllPhysicsBodyDefinitions() []*BodyDefinition {
	var result []*BodyDefinition
	for _, placement := range s.AllPhysicsBodyPlacements() {
		body := placement.Value
		definition := body.Definition()
		result = append(result, definition)
	}
	return gog.Dedupe(result)
}

func (s *Model) AllPhysicsBodyPlacements() []Placed[*Body] {
	var result []Placed[*Body]
	for _, node := range s.NodesIter() {
		switch source := node.Source().(type) {
		case *Body:
			result = append(result, Placed[*Body]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

func (s *Model) AllSkyPlacements() []Placed[*Sky] {
	var result []Placed[*Sky]
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *Sky:
			result = append(result, Placed[*Sky]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
}

// TODO: Consider moving this to gog project.
func indexedIter[T any](src iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		index := 0
		for item := range src {
			if !yield(index, item) {
				return
			}
			index++
		}
	}
}
