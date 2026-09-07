package game

import (
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
)

type SkyBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.Sky] = (*SkyBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.Sky] = (*SkyBindingSolver)(nil)

func NewSkyBindingSolver() *SkyBindingSolver {
	return &SkyBindingSolver{}
}

func (s *SkyBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, sky *graphics.Sky) {
	sky.Delete()
}

func (s *SkyBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, sky *graphics.Sky, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()

	sky.SetActive(visible)
}

type AmbientLightBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.AmbientLight] = (*AmbientLightBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.AmbientLight] = (*AmbientLightBindingSolver)(nil)

func NewAmbientLightBindingSolver() *AmbientLightBindingSolver {
	return &AmbientLightBindingSolver{}
}

func (s *AmbientLightBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.AmbientLight) {
	light.Delete()
}

func (s *AmbientLightBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.AmbientLight, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	light.SetActive(visible)
	light.SetPosition(matrix.Translation())
}

type PointLightBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.PointLight] = (*PointLightBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.PointLight] = (*PointLightBindingSolver)(nil)

func NewPointLightBindingSolver() *PointLightBindingSolver {
	return &PointLightBindingSolver{}
}

func (s *PointLightBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.PointLight) {
	light.Delete()
}

func (s *PointLightBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.PointLight, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	light.SetActive(visible)
	light.SetPosition(matrix.Translation())
}

type SpotLightBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.SpotLight] = (*SpotLightBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.SpotLight] = (*SpotLightBindingSolver)(nil)

func NewSpotLightBindingSolver() *SpotLightBindingSolver {
	return &SpotLightBindingSolver{}
}

func (s *SpotLightBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.SpotLight) {
	light.Delete()
}

func (s *SpotLightBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.SpotLight, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	light.SetActive(visible)
	translation, rotation, _ := matrix.TRS()
	light.SetPosition(translation)
	light.SetRotation(rotation)
}

type DirectionalLightBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.DirectionalLight] = (*DirectionalLightBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.DirectionalLight] = (*DirectionalLightBindingSolver)(nil)

func NewDirectionalLightBindingSolver() *DirectionalLightBindingSolver {
	return &DirectionalLightBindingSolver{}
}

func (s *DirectionalLightBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.DirectionalLight) {
	light.Delete()
}

func (s *DirectionalLightBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, light *graphics.DirectionalLight, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	light.SetActive(visible)
	translation, rotation, _ := matrix.TRS()
	light.SetPosition(translation)
	light.SetRotation(rotation)
}

type MeshBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.Mesh] = (*MeshBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.Mesh] = (*MeshBindingSolver)(nil)

func NewMeshBindingSolver() *MeshBindingSolver {
	return &MeshBindingSolver{}
}

func (s *MeshBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, mesh *graphics.Mesh) {
	mesh.Delete()
}

func (s *MeshBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, mesh *graphics.Mesh, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	visible := node.IsAbsoluteVisible()
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	mesh.SetActive(visible)
	mesh.SetMatrix(matrix)
}

type BoneTarget struct {
	Armature  *graphics.Armature
	BoneIndex int
}

type BoneBindingSolver struct{}

var _ hierarchy.InterpolationBindingSolver[BoneTarget] = (*BoneBindingSolver)(nil)

func NewBoneBindingSolver() *BoneBindingSolver {
	return &BoneBindingSolver{}
}

func (s *BoneBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, target BoneTarget, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	target.Armature.SetBone(target.BoneIndex, dtos.Mat4(matrix))
}

type CameraBindingSolver struct{}

var _ hierarchy.LifecycleBindingSolver[*graphics.Camera] = (*CameraBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[*graphics.Camera] = (*CameraBindingSolver)(nil)

func NewCameraBindingSolver() *CameraBindingSolver {
	return &CameraBindingSolver{}
}

func (s *CameraBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, camera *graphics.Camera) {
	camera.Delete()
}

func (s *CameraBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, camera *graphics.Camera, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	matrix := node.InterpolatedAbsoluteMatrix(fraction)

	camera.SetMatrix(matrix)
}
