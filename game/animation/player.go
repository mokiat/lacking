package animation

import "time"

// TODO: Consider getting rid of this player. The Node can be used directly.

func NewPlayer(root Node) *Player {
	return &Player{
		root: root,
	}
}

type Player struct {
	root Node
}

func (p *Player) Update(elapsedTime time.Duration) {
	p.root.Synchronize()
	p.root.Advance(elapsedTime.Seconds(), 1.0)
}

func (p *Player) BoneTransform(name string) NodeTransform {
	return p.root.BoneTransform(name)
}
