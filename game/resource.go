package game

import (
	"golang.org/x/exp/maps"
)

type Resource interface {
	Delete()
}

// TODO: Use the same nested concept for this ResourceSet as is done
// in the ui framework for the Context. If something needs to be loaded for
// a short duration, a dedicated nested ResourceSet should be created.

func newResourceSet(parent *ResourceSet, engine *Engine) *ResourceSet {
	return &ResourceSet{
		engine: engine,
	}
}

type ResourceSet struct {
	engine *Engine

	namedResources map[string]Resource
	adhocResources []Resource
}

func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

// // WARNING: DO NOT WAIT ON THE PROMISE FROM THE MAIN GOROUTINE!!!
// func (s *ResourceSet) OpenTwoDTexture(resourceSet *ResourceSet, id string) *TwoDTexture {
// 	result := &TwoDTexture{}
// 	// TODO: ioWorker.Schedule(func() {...})
// 	go func() {

// 	}()
// 	// panic("TODO")
// 	return result // TODO: Result should be usable until it loads it
// }

// func (e *Engine) OpenCubeTexture(resourceSet *ResourceSet, id string) async.Promise[*graphics.CubeTexture] {
// 	return nil // TODO
// }

// func (e *Engine) OpenModel(resourceSet *ResourceSet, id string) async.Promise[*]

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
	return false // TODO: Check that all resources are loaded
}

// func (r *ResourceSet) Wait() error {
// 	return nil
// }
