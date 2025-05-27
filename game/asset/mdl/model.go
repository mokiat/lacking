package mdl

import (
	"iter"
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
)

type Model struct {
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
	for _, light := range s.AllAmbientLights() {
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
	for _, mesh := range s.AllMeshes() {
		definition := mesh.Definition()
		result = append(result, definition.Materials()...)
	}
	for _, sky := range s.AllSkies() {
		result = append(result, sky.Material())
	}
	return gog.Dedupe(result)
}

func (s *Model) AllAmbientLights() []*AmbientLight {
	var result []*AmbientLight
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *AmbientLight:
			result = append(result, source)
		}
	}
	return gog.Dedupe(result)
}

func (s *Model) AllMeshes() []*Mesh {
	var result []*Mesh
	seen := ds.NewSet[*Mesh](0)
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *Mesh:
			if !seen.Contains(source) {
				result = append(result, source)
			}
			seen.Add(source)
		}
	}
	return result
}

func (s *Model) AllPhysicsBodyMaterials() []*BodyMaterial {
	var result []*BodyMaterial
	seen := ds.NewSet[*BodyMaterial](0)
	for _, definition := range s.AllPhysicsBodyDefinitions() {
		material := definition.Material()
		if !seen.Contains(material) {
			result = append(result, material)
		}
		seen.Add(material)
	}
	return result
}

func (s *Model) AllPhysicsBodyDefinitions() []*BodyDefinition {
	var result []*BodyDefinition
	seen := ds.NewSet[*BodyDefinition](0)
	for _, placement := range s.AllPhysicsBodyPlacements() {
		body := placement.Value
		definition := body.Definition()
		if !seen.Contains(definition) {
			result = append(result, definition)
		}
		seen.Add(definition)
	}
	return result
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

func (s *Model) AllSkies() []*Sky {
	var result []*Sky
	seen := ds.NewSet[*Sky](0)
	for _, node := range s.NodesIter() {
		switch source := node.Target().(type) {
		case *Sky:
			if !seen.Contains(source) {
				result = append(result, source)
			}
			seen.Add(source)
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
