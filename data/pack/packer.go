package pack

import (
	"fmt"
	"runtime/debug"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/util/resource"
	"golang.org/x/sync/errgroup"
)

var ErrFocused = fmt.Errorf("focused pipelines")

func NewPacker(registry *asset.Registry) *Packer {
	return &Packer{
		registry: registry,
	}
}

type Packer struct {
	registry         *asset.Registry
	pipelines        []*Pipeline
	focusedPipelines []*Pipeline
}

func (p *Packer) Pipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), p.registry, resource.NewFileLocator("./"))
	fn(pipeline)
	p.pipelines = append(p.pipelines, pipeline)
}

func (p *Packer) FPipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), p.registry, resource.NewFileLocator("./"))
	fn(pipeline)
	p.focusedPipelines = append(p.focusedPipelines, pipeline)
}

func (p *Packer) RunSerial() error {
	pipelines, focused := p.activePipelines()
	for _, pipeline := range pipelines {
		if err := p.runPipeline(pipeline); err != nil {
			return err
		}
	}
	if focused {
		return ErrFocused
	}
	return nil
}

func (p *Packer) RunParallel() error {
	pipelines, focused := p.activePipelines()

	var group errgroup.Group
	for _, pipeline := range pipelines {
		pip := pipeline
		group.Go(func() error {
			return p.runPipeline(pip)
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	if focused {
		return ErrFocused
	}
	return nil
}

func (p *Packer) activePipelines() ([]*Pipeline, bool) {
	if len(p.focusedPipelines) > 0 {
		return p.focusedPipelines, true
	}
	return p.pipelines, false
}

func (p *Packer) runPipeline(pipeline *Pipeline) (err error) {
	defer func() {
		if problem := recover(); problem != nil {
			err = fmt.Errorf("pipeline %d paniced: %v", pipeline.id, string(debug.Stack()))
			return
		}
	}()
	if err := pipeline.execute(); err != nil {
		return fmt.Errorf("pipeline %d error: %w", pipeline.id, err)
	}
	return nil
}
