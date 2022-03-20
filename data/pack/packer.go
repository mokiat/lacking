package pack

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	gameasset "github.com/mokiat/lacking/game/asset"
)

func NewPacker(registry gameasset.Registry) *Packer {
	return &Packer{
		registry: registry,
	}
}

type Packer struct {
	registry         gameasset.Registry
	pipelines        []*Pipeline
	focusedPipelines []*Pipeline
}

func (p *Packer) Pipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), p.registry, FileResourceLocator{})
	fn(pipeline)
	p.pipelines = append(p.pipelines, pipeline)
}

func (p *Packer) FPipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), p.registry, FileResourceLocator{})
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
		return fmt.Errorf("focused pipelines")
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
		return fmt.Errorf("focused pipelines")
	}
	return nil
}

func (p *Packer) activePipelines() ([]*Pipeline, bool) {
	if len(p.focusedPipelines) > 0 {
		return p.focusedPipelines, true
	}
	return p.pipelines, false
}

func (p *Packer) runPipeline(pipeline *Pipeline) error {
	if err := pipeline.execute(); err != nil {
		return fmt.Errorf("pipeline %d error: %w", pipeline.id, err)
	}
	return nil
}
