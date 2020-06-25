package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

const ProgramTypeName = TypeName("program")

func InjectProgram(target **Program) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Program)
	}
}

type Program struct {
	Name       string
	GFXProgram *graphics.Program
}

func NewProgramOperator(locator Locator, gfxWorker *async.Worker) *ProgramOperator {
	return &ProgramOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type ProgramOperator struct {
	locator   Locator
	gfxWorker *async.Worker
}

func (o *ProgramOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "programs", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open program asset %q: %w", name, err)
	}
	defer in.Close()

	programAsset := new(asset.Program)
	if err := asset.DecodeProgram(in, programAsset); err != nil {
		return nil, fmt.Errorf("failed to decode program asset %q: %w", name, err)
	}

	program := &Program{
		Name:       name,
		GFXProgram: &graphics.Program{},
	}

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		return program.GFXProgram.Allocate(graphics.ProgramData{
			VertexShaderSourceCode:   programAsset.VertexSourceCode,
			FragmentShaderSourceCode: programAsset.FragmentSourceCode,
		})
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return nil, fmt.Errorf("failed to allocate gfx program: %w", err)
	}
	return program, nil
}

func (o *ProgramOperator) Release(registry *Registry, resource interface{}) error {
	program := resource.(*Program)

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		return program.GFXProgram.Release()
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return fmt.Errorf("failed to release gfx program: %w", err)
	}
	return nil
}
