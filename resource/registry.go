package resource

import (
	"sync"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/graphics"
)

type TypeName string

type Operator interface {
	Allocate(registry *Registry, name string) (interface{}, error)
	Release(registry *Registry, resource interface{}) error
}

func NewRegistry(locator Locator, gfxWorker *graphics.Worker) *Registry {
	registry := &Registry{
		catalog: make(map[TypeName]*Type),
	}
	registry.Register(ProgramTypeName, NewProgramOperator(locator, gfxWorker))
	registry.Register(TwoDTextureTypeName, NewTwoDTextureOperator(locator, gfxWorker))
	registry.Register(CubeTextureTypeName, NewCubeTextureOperator(locator, gfxWorker))
	registry.Register(MeshTypeName, NewMeshOperator(locator, gfxWorker))
	registry.Register(ModelTypeName, NewModelOperator(locator, gfxWorker))
	registry.Register(LevelTypeName, NewLevelOperator(locator, gfxWorker))
	return registry
}

type Registry struct {
	catalog map[TypeName]*Type
}

func (r *Registry) Register(typeName TypeName, operator Operator) {
	r.catalog[typeName] = &Type{
		registry:   r,
		operator:   operator,
		references: make(map[string]*Reference),
	}
}

func (r *Registry) Load(typeName TypeName, name string) async.Outcome {
	resType := r.catalog[typeName]
	return resType.Load(name)
}

func (r *Registry) Unload(typeName TypeName, name string) async.Outcome {
	resType := r.catalog[ProgramTypeName]
	return resType.Unload(name)
}

func (r *Registry) LoadProgram(name string) async.Outcome {
	return r.Load(ProgramTypeName, name)
}

func (r *Registry) UnloadProgram(program *Program) async.Outcome {
	return r.Unload(ProgramTypeName, program.Name)
}

func (r *Registry) LoadTwoDTexture(name string) async.Outcome {
	return r.Load(TwoDTextureTypeName, name)
}

func (r *Registry) UnloadTwoDTexture(texture *TwoDTexture) async.Outcome {
	return r.Unload(TwoDTextureTypeName, texture.Name)
}

func (r *Registry) LoadCubeTexture(name string) async.Outcome {
	return r.Load(CubeTextureTypeName, name)
}

func (r *Registry) UnloadCubeTexture(texture *CubeTexture) async.Outcome {
	return r.Unload(CubeTextureTypeName, texture.Name)
}

func (r *Registry) LoadMesh(name string) async.Outcome {
	return r.Load(MeshTypeName, name)
}

func (r *Registry) UnloadMesh(texture *Mesh) async.Outcome {
	return r.Unload(MeshTypeName, texture.Name)
}

func (r *Registry) LoadModel(name string) async.Outcome {
	return r.Load(ModelTypeName, name)
}

func (r *Registry) UnloadModel(texture *Model) async.Outcome {
	return r.Unload(ModelTypeName, texture.Name)
}

func (r *Registry) LoadLevel(name string) async.Outcome {
	return r.Load(LevelTypeName, name)
}

func (r *Registry) UnloadLevel(texture *Level) async.Outcome {
	return r.Unload(LevelTypeName, texture.Name)
}

type Type struct {
	mu         sync.Mutex
	registry   *Registry
	operator   Operator
	references map[string]*Reference
}

func (t *Type) Load(name string) async.Outcome {
	t.mu.Lock()
	defer t.mu.Unlock()

	if reference, ok := t.references[name]; ok {
		reference.Count++
		return reference.LoadOperation
	}

	reference := &Reference{
		Count: 1,
		Value: nil,
	}
	reference.LoadOperation = t.loadReference(name, reference)
	return reference.LoadOperation
}

func (t *Type) Unload(name string) async.Outcome {
	t.mu.Lock()
	defer t.mu.Unlock()

	reference, ok := t.references[name]
	if !ok {
		return async.NewValueOutcome(nil)
	}

	if reference.Count--; reference.Count > 0 {
		return async.NewValueOutcome(nil)
	}

	delete(t.references, name)
	reference.UnloadOperation = t.unloadReference(name, reference)
	return reference.UnloadOperation
}

func (t *Type) loadReference(name string, reference *Reference) async.Outcome {
	output := async.NewOutcome()
	go func() {
		value, err := t.operator.Allocate(t.registry, name)
		output.Record(async.Result{
			Value: value,
			Err:   err,
		})
	}()
	return output
}

func (t *Type) unloadReference(name string, reference *Reference) async.Outcome {
	output := async.NewOutcome()
	go func() {
		reference.LoadOperation.Wait() // ensure we are not still in the middle of a load

		err := t.operator.Release(t.registry, reference.Value)
		output.Record(async.Result{
			Err: err,
		})
	}()
	return output
}

// TODO: Make private
type Reference struct {
	Count           int
	LoadOperation   async.Outcome
	UnloadOperation async.Outcome
	Value           interface{}
}