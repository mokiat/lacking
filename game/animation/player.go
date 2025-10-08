package animation

import (
	"time"
)

func NewPlayer(root Node, boneNames []string) *Player {
	return &Player{
		root:           root,
		boneNames:      boneNames,
		boneTransforms: make(map[string]NodeTransform),
	}
}

type Player struct {
	root           Node
	boneNames      []string
	boneTransforms map[string]NodeTransform
}

func (p *Player) Update(elapsedTime time.Duration) {
	p.root.Synchronize()
	p.root.Advance(elapsedTime.Seconds(), 1.0)
	for _, name := range p.boneNames {
		p.boneTransforms[name] = p.root.BoneTransform(name)
	}
}

func (p *Player) BoneTransform(name string) NodeTransform {
	return p.boneTransforms[name]
}
