package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
)

type BodyNodeSource struct {
	Body physics.Body
}

func (s BodyNodeSource) ApplyTo(node *hierarchy.Node) {
	translation := s.Body.IntermediatePosition()
	rotation := s.Body.IntermediateRotation()
	scale := dprec.NewVec3(1.0, 1.0, 1.0)
	node.SetAbsoluteMatrix(dprec.TRSMat4(translation, rotation, scale))
}

type CameraNodeTarget struct {
	Camera *graphics.Camera
}

func (t CameraNodeTarget) ApplyFrom(node *hierarchy.Node) {
	t.Camera.SetMatrix(node.AbsoluteMatrix())
}

type MeshNodeTarget struct {
	Mesh *graphics.Mesh
}

func (t MeshNodeTarget) ApplyFrom(node *hierarchy.Node) {
	t.Mesh.SetMatrix(node.AbsoluteMatrix())
}

type BoneNodeTarget struct {
	Armature  *graphics.Armature
	BoneIndex int
}

func (t BoneNodeTarget) ApplyFrom(node *hierarchy.Node) {
	matrix := node.AbsoluteMatrix()
	t.Armature.SetBone(t.BoneIndex, dtos.Mat4(matrix))
}

type PointLightNodeTarget struct {
	Light *graphics.PointLight
}

func (t PointLightNodeTarget) ApplyFrom(node *hierarchy.Node) {
	matrix := node.AbsoluteMatrix()
	t.Light.SetPosition(matrix.Translation())
}

type SpotLightNodeTarget struct {
	Light *graphics.SpotLight
}

func (t SpotLightNodeTarget) ApplyFrom(node *hierarchy.Node) {
	matrix := node.AbsoluteMatrix()
	translation, rotation, _ := matrix.TRS()
	t.Light.SetPosition(translation)
	t.Light.SetRotation(rotation)
}

type DirectionalLightNodeTarget struct {
	Light                 *graphics.DirectionalLight
	UseOnlyParentPosition bool
}

func (t DirectionalLightNodeTarget) ApplyFrom(node *hierarchy.Node) {
	if t.UseOnlyParentPosition {
		matrix := node.BaseAbsoluteMatrix()
		t.Light.SetPosition(dprec.Vec3Sum(
			matrix.Translation(),
			node.Position(),
		))
		t.Light.SetRotation(node.Rotation())
	} else {
		matrix := node.AbsoluteMatrix()
		translation, rotation, _ := matrix.TRS()
		t.Light.SetPosition(translation)
		t.Light.SetRotation(rotation)
	}
}
