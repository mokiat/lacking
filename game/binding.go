package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
)

// GenericBindingSet represents any hierarchy binding set, regardless of the
// target type.
type GenericBindingSet interface {
	ApplyTargetToNode(id hierarchy.NodeID)
	ApplyTargetsToNodes()
	ApplyNodeToTarget(id hierarchy.NodeID, fraction float64)
	ApplyNodesToTargets(fraction float64)
	DeleteStale()
}

// NewAnimationBinding creates a new binding for animations.
func NewAnimationBinding() hierarchy.Binding[*animation.Player] {
	return &animationBinding{}
}

type animationBinding struct{}

var _ hierarchy.Binding[*animation.Player] = (*animationBinding)(nil)

func (b *animationBinding) OnTargetToNode(scene *hierarchy.Scene, player *animation.Player, id hierarchy.NodeID) {
	name := scene.NodeName(id)
	transform := player.BoneTransform(name)
	if transform.Translation.Specified {
		scene.SetNodePosition(id, transform.Translation.Value)
	}
	if transform.Rotation.Specified {
		scene.SetNodeRotation(id, transform.Rotation.Value)
	}
	if transform.Scale.Specified {
		scene.SetNodeScale(id, transform.Scale.Value)
	}
}

func (b *animationBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, player *animation.Player, fraction float64) {
	// Animation is not a target. Nothing to do.
}

func (b *animationBinding) OnStaleBinding(scene *hierarchy.Scene, player *animation.Player) {
	// Nothing to do.
}

// NewBodyBinding creates a new binding for physics bodies.
func NewBodyBinding() hierarchy.Binding[physics.Body] {
	return &bodyBinding{}
}

type bodyBinding struct{}

var _ hierarchy.Binding[physics.Body] = (*bodyBinding)(nil)

func (b *bodyBinding) OnTargetToNode(scene *hierarchy.Scene, body physics.Body, id hierarchy.NodeID) {
	currentTranslation := body.Position()
	currentRotation := body.Rotation()

	scene.SetNodeAbsoluteMatrix(id, dprec.TRSMat4(
		currentTranslation,
		currentRotation,
		dprec.NewVec3(1.0, 1.0, 1.0),
	))
}

func (b *bodyBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, body physics.Body, fraction float64) {
	// Body is not a target. Nothing to do.
}

func (b *bodyBinding) OnStaleBinding(scene *hierarchy.Scene, body physics.Body) {
	body.Delete()
}

// NewSkyBinding creates a new binding for skies.
func NewSkyBinding() hierarchy.Binding[*graphics.Sky] {
	return &skyBinding{}
}

type skyBinding struct{}

var _ hierarchy.Binding[*graphics.Sky] = (*skyBinding)(nil)

func (b *skyBinding) OnTargetToNode(*hierarchy.Scene, *graphics.Sky, hierarchy.NodeID) {
	// Sky is not a source. Nothing to do.
}

func (b *skyBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, sky *graphics.Sky, fraction float64) {
	active := scene.IsNodeVisible(id)

	sky.SetActive(active)
}

func (b *skyBinding) OnStaleBinding(_ *hierarchy.Scene, sky *graphics.Sky) {
	sky.Delete()
}

// NewAmbientLightBinding creates a new binding for ambient lights.
func NewAmbientLightBinding() hierarchy.Binding[*graphics.AmbientLight] {
	return &ambientLightBinding{}
}

type ambientLightBinding struct{}

var _ hierarchy.Binding[*graphics.AmbientLight] = (*ambientLightBinding)(nil)

func (b *ambientLightBinding) OnTargetToNode(scene *hierarchy.Scene, light *graphics.AmbientLight, id hierarchy.NodeID) {
	// Ambient light is not a source. Nothing to do.
}

func (b *ambientLightBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.AmbientLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	light.SetPosition(matrix.Translation())
	light.SetActive(visible)
}

func (b *ambientLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.AmbientLight) {
	light.Delete()
}

// NewPointLightBinding creates a new binding for point lights.
func NewPointLightBinding() hierarchy.Binding[*graphics.PointLight] {
	return &pointLightBinding{}
}

type pointLightBinding struct{}

var _ hierarchy.Binding[*graphics.PointLight] = (*pointLightBinding)(nil)

func (b *pointLightBinding) OnTargetToNode(scene *hierarchy.Scene, light *graphics.PointLight, id hierarchy.NodeID) {
	// Point light is not a source. Nothing to do.
}

func (b *pointLightBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.PointLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	light.SetPosition(matrix.Translation())
	light.SetActive(visible)
}

func (b *pointLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.PointLight) {
	light.Delete()
}

// NewSpotLightBinding creates a new binding for spot lights.
func NewSpotLightBinding() hierarchy.Binding[*graphics.SpotLight] {
	return &spotLightBinding{}
}

type spotLightBinding struct{}

var _ hierarchy.Binding[*graphics.SpotLight] = (*spotLightBinding)(nil)

func (b *spotLightBinding) OnTargetToNode(scene *hierarchy.Scene, light *graphics.SpotLight, id hierarchy.NodeID) {
	// Spot light is not a source. Nothing to do.
}

func (b *spotLightBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.SpotLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	translation, rotation, _ := matrix.TRS()
	light.SetPosition(translation)
	light.SetRotation(rotation)
	light.SetActive(visible)
}

func (b *spotLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.SpotLight) {
	light.Delete()
}

// NewDirectionalLightBinding creates a new binding for directional lights.
func NewDirectionalLightBinding() hierarchy.Binding[*graphics.DirectionalLight] {
	return &directionalLightBinding{}
}

type directionalLightBinding struct{}

var _ hierarchy.Binding[*graphics.DirectionalLight] = (*directionalLightBinding)(nil)

func (b *directionalLightBinding) OnTargetToNode(scene *hierarchy.Scene, light *graphics.DirectionalLight, id hierarchy.NodeID) {
	// Directional light is not a source. Nothing to do.
}

func (b *directionalLightBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.DirectionalLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	translation, rotation, _ := matrix.TRS()
	light.SetPosition(translation)
	light.SetRotation(rotation)
	light.SetActive(visible)
}

func (b *directionalLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.DirectionalLight) {
	light.Delete()
}

// NewMeshBinding creates a new binding for meshes.
func NewMeshBinding() hierarchy.Binding[*graphics.Mesh] {
	return &meshBinding{}
}

type meshBinding struct{}

var _ hierarchy.Binding[*graphics.Mesh] = (*meshBinding)(nil)

func (b *meshBinding) OnTargetToNode(scene *hierarchy.Scene, mesh *graphics.Mesh, id hierarchy.NodeID) {
	// Mesh is not a source. Nothing to do.
}

func (b *meshBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, mesh *graphics.Mesh, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	mesh.SetMatrix(matrix)
	mesh.SetActive(visible)
}

func (b *meshBinding) OnStaleBinding(scene *hierarchy.Scene, mesh *graphics.Mesh) {
	mesh.Delete()
}

// BoneTarget is a placeholder type for armature bone bindings.
type BoneTarget struct {
	Armature  *graphics.Armature
	BoneIndex int
}

// NewBoneBinding creates a new binding for armature bones.
func NewBoneBinding() hierarchy.Binding[BoneTarget] {
	return &boneBinding{}
}

type boneBinding struct{}

var _ hierarchy.Binding[BoneTarget] = (*boneBinding)(nil)

func (b *boneBinding) OnTargetToNode(scene *hierarchy.Scene, target BoneTarget, id hierarchy.NodeID) {
	// Bone is not a source. Nothing to do.
}

func (b *boneBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, target BoneTarget, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	target.Armature.SetBone(target.BoneIndex, dtos.Mat4(matrix))
}

func (b *boneBinding) OnStaleBinding(scene *hierarchy.Scene, target BoneTarget) {
	// Nothing to do.
}

// NewCameraBinding creates a new binding for cameras.
func NewCameraBinding() hierarchy.Binding[*graphics.Camera] {
	return &cameraBinding{}
}

type cameraBinding struct{}

var _ hierarchy.Binding[*graphics.Camera] = (*cameraBinding)(nil)

func (b *cameraBinding) OnTargetToNode(scene *hierarchy.Scene, camera *graphics.Camera, id hierarchy.NodeID) {
	// Camera is not a source. Nothing to do.
}

func (b *cameraBinding) OnNodeToTarget(scene *hierarchy.Scene, id hierarchy.NodeID, camera *graphics.Camera, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	camera.SetMatrix(matrix)
}

func (b *cameraBinding) OnStaleBinding(scene *hierarchy.Scene, camera *graphics.Camera) {
	camera.Delete()
}
