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
func NewAnimationBinding() hierarchy.SourceBinding[*animation.Player] {
	return &animationBinding{}
}

type animationBinding struct{}

func (b *animationBinding) OnSourceToNode(scene *hierarchy.Scene, player *animation.Player, id hierarchy.NodeID) {
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

func (b *animationBinding) OnStaleBinding(scene *hierarchy.Scene, player *animation.Player) {
	// Nothing to do.
}

// NewBodyBinding creates a new binding for physics bodies.
func NewBodyBinding() hierarchy.SourceBinding[physics.Body] {
	return &bodyBinding{}
}

type bodyBinding struct{}

func (b *bodyBinding) OnSourceToNode(scene *hierarchy.Scene, body physics.Body, id hierarchy.NodeID) {
	currentTranslation := body.Position()
	currentRotation := body.Rotation()

	scene.SetNodeAbsoluteMatrix(id, dprec.TRSMat4(
		currentTranslation,
		currentRotation,
		dprec.NewVec3(1.0, 1.0, 1.0),
	))
}

func (b *bodyBinding) OnStaleBinding(scene *hierarchy.Scene, body physics.Body) {
	body.Delete()
}

// NewSkyBinding creates a new binding for skies.
func NewSkyBinding() hierarchy.InterpolationBinding[*graphics.Sky] {
	return &skyBinding{}
}

type skyBinding struct{}

func (b *skyBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, sky *graphics.Sky, fraction float64) {
	active := scene.IsNodeVisible(id)

	sky.SetActive(active)
}

func (b *skyBinding) OnStaleBinding(_ *hierarchy.Scene, sky *graphics.Sky) {
	sky.Delete()
}

// NewAmbientLightBinding creates a new binding for ambient lights.
func NewAmbientLightBinding() hierarchy.InterpolationBinding[*graphics.AmbientLight] {
	return &ambientLightBinding{}
}

type ambientLightBinding struct{}

func (b *ambientLightBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.AmbientLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	light.SetPosition(matrix.Translation())
	light.SetActive(visible)
}

func (b *ambientLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.AmbientLight) {
	light.Delete()
}

// NewPointLightBinding creates a new binding for point lights.
func NewPointLightBinding() hierarchy.InterpolationBinding[*graphics.PointLight] {
	return &pointLightBinding{}
}

type pointLightBinding struct{}

func (b *pointLightBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.PointLight, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	visible := scene.IsNodeVisible(id)

	light.SetPosition(matrix.Translation())
	light.SetActive(visible)
}

func (b *pointLightBinding) OnStaleBinding(scene *hierarchy.Scene, light *graphics.PointLight) {
	light.Delete()
}

// NewSpotLightBinding creates a new binding for spot lights.
func NewSpotLightBinding() hierarchy.InterpolationBinding[*graphics.SpotLight] {
	return &spotLightBinding{}
}

type spotLightBinding struct{}

func (b *spotLightBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.SpotLight, fraction float64) {
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
func NewDirectionalLightBinding() hierarchy.InterpolationBinding[*graphics.DirectionalLight] {
	return &directionalLightBinding{}
}

type directionalLightBinding struct{}

func (b *directionalLightBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, light *graphics.DirectionalLight, fraction float64) {
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
func NewMeshBinding() hierarchy.InterpolationBinding[*graphics.Mesh] {
	return &meshBinding{}
}

type meshBinding struct{}

func (b *meshBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, mesh *graphics.Mesh, fraction float64) {
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
func NewBoneBinding() hierarchy.InterpolationBinding[BoneTarget] {
	return &boneBinding{}
}

type boneBinding struct{}

func (b *boneBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, target BoneTarget, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	target.Armature.SetBone(target.BoneIndex, dtos.Mat4(matrix))
}

func (b *boneBinding) OnStaleBinding(scene *hierarchy.Scene, target BoneTarget) {
	// Nothing to do.
}

// NewCameraBinding creates a new binding for cameras.
func NewCameraBinding() hierarchy.InterpolationBinding[*graphics.Camera] {
	return &cameraBinding{}
}

type cameraBinding struct{}

func (b *cameraBinding) OnNodeToInterpolation(scene *hierarchy.Scene, id hierarchy.NodeID, camera *graphics.Camera, fraction float64) {
	matrix := scene.NodeInterpolatedAbsoluteMatrix(id, fraction)
	camera.SetMatrix(matrix)
}

func (b *cameraBinding) OnStaleBinding(scene *hierarchy.Scene, camera *graphics.Camera) {
	camera.Delete()
}
