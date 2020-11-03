package pack

import (
	"log"
	"sync"
)

func NewPacker() *Packer {
	return &Packer{}
}

type Packer struct {
	pipelines        []*Pipeline
	focusedPipelines []*Pipeline
}

func (p *Packer) Pipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), FileResourceLocator{}, FileAssetLocator{})
	fn(pipeline)
	p.pipelines = append(p.pipelines, pipeline)
}

func (p *Packer) FPipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), FileResourceLocator{}, FileAssetLocator{})
	fn(pipeline)
	p.focusedPipelines = append(p.focusedPipelines, pipeline)
}

func (p *Packer) RunSerial() {
	pipelines, focused := p.activePipelines()
	for _, pipeline := range pipelines {
		p.runPipeline(pipeline)
	}
	if focused {
		log.Fatalln("failing due to focused pipelines")
	}
}

func (p *Packer) RunParallel() {
	pipelines, focused := p.activePipelines()
	wait := &sync.WaitGroup{}
	wait.Add(len(pipelines))
	for _, pipeline := range pipelines {
		go func(pip *Pipeline) {
			p.runPipeline(pip)
			wait.Done()
		}(pipeline)
	}
	wait.Wait()
	if focused {
		log.Fatalln("failing due to focused pipelines")
	}
}

func (p *Packer) activePipelines() ([]*Pipeline, bool) {
	if len(p.focusedPipelines) > 0 {
		return p.focusedPipelines, true
	}
	return p.pipelines, false
}

func (p *Packer) runPipeline(pipeline *Pipeline) {
	if err := pipeline.execute(); err != nil {
		log.Fatalf("pipeline %d error: %v", pipeline.id, err)
	}
}
