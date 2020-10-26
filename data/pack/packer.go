package pack

import (
	"log"
	"sync"
)

func NewPacker() *Packer {
	return &Packer{}
}

type Packer struct {
	pipelines []*Pipeline
}

func (p *Packer) Pipeline(fn func(*Pipeline)) {
	pipeline := newPipeline(len(p.pipelines), FileResourceLocator{}, FileAssetLocator{})
	fn(pipeline)
	p.pipelines = append(p.pipelines, pipeline)
}

func (p *Packer) RunSerial() {
	for _, pipeline := range p.pipelines {
		p.runPipeline(pipeline)
	}
}

func (p *Packer) RunParallel() {
	wait := &sync.WaitGroup{}
	wait.Add(len(p.pipelines))
	for _, pipeline := range p.pipelines {
		go func(pip *Pipeline) {
			p.runPipeline(pip)
			wait.Done()
		}(pipeline)
	}
	wait.Wait()
}

func (p *Packer) runPipeline(pipeline *Pipeline) {
	if err := pipeline.execute(); err != nil {
		log.Fatalf("pipeline %d error: %v", pipeline.id, err)
	}
}
