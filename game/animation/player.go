package animation

import (
	"maps"
	"time"
)

func NewPlayer(root Node, boneNames []string) *Player {
	return &Player{
		root:      root,
		boneNames: boneNames,

		previousTransforms: make(map[string]NodeTransform),
		currentTransforms:  make(map[string]NodeTransform),
	}
}

type Player struct {
	root      Node
	boneNames []string

	previousTransforms map[string]NodeTransform
	currentTransforms  map[string]NodeTransform
}

func (p *Player) Update(elapsedTime time.Duration) {
	clear(p.previousTransforms)
	maps.Copy(p.previousTransforms, p.currentTransforms)

	p.root.Synchronize()
	p.root.Advance(elapsedTime.Seconds(), 1.0)

	for _, name := range p.boneNames {
		p.currentTransforms[name] = p.root.BoneTransform(name)
	}
}

func (p *Player) BoneTransform(name string, interpolation float64) NodeTransform {
	previousTransform := p.previousTransforms[name]
	currentTransform := p.currentTransforms[name]
	return BlendNodeTransforms(previousTransform, currentTransform, interpolation)
}
