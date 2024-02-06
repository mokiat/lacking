package asset

import "github.com/mokiat/gomath/dprec"

const UnspecifiedNodeIndex = int32(-1)

type Node struct {
	Name        string     `json:"name"`
	ParentIndex int32      `json:"parent_index"`
	Translation dprec.Vec3 `json:"translation"`
	Rotation    dprec.Quat `json:"rotation"`
	Scale       dprec.Vec3 `json:"scale"`
	Mask        NodeMask   `json:"mask"`
}

type NodeMask uint32

const (
	NodeMaskNone       NodeMask = 0
	NodeMaskStationary NodeMask = 1 << iota
	NodeMaskInseparable
)
