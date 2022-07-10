package game

import (
	"github.com/mokiat/lacking/game/graphics"
	"golang.org/x/exp/maps"
)

type Resource interface {
	Delete()
}

// TODO: Use the same nested concept for this ResourceSet as is done
// in the ui framework for the Context. If something needs to be loaded for
// a short duration, a dedicated nested ResourceSet should be created.

func NewResourceSet(parent *ResourceSet) *ResourceSet {
	return &ResourceSet{}
}

type ResourceSet struct {
	namedResources map[string]Resource
	adhocResources []Resource
}

func (s *ResourceSet) Delete() {
	for _, resource := range s.namedResources {
		resource.Delete()
	}
	maps.Clear(s.namedResources)
	for _, resource := range s.adhocResources {
		resource.Delete()
	}
	s.adhocResources = nil
}

func (r *ResourceSet) Ready() bool {
	return false
}

func (r *ResourceSet) Wait() error {
	return nil
}

type TwoDTexture struct {
	gfxTexture *graphics.TwoDTexture
}

// func (t *TwoDTexture) GraphicsTexture() *graphics.TwoDTexture {

// }

// func (t *TwoDTexture) Delete() {

// }
