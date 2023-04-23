package physics

import (
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/util/spatial"
)

type PropInfo struct {
	CollisionSet collision.Set
}

type Prop struct {
	scene        *Scene
	octreeItem   *spatial.OctreeItem[*Prop]
	collisionSet collision.Set
}

func (p *Prop) Delete() {
	p.octreeItem.Delete()
	p.collisionSet = collision.Set{}
	delete(p.scene.props, p)
}
