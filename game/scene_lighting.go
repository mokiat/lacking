package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/render"
)

// AmbientLightInfo contains the information required to create an ambient
// light.
type AmbientLightInfo struct {
	ReflectionTexture render.Texture
	RefractionTexture render.Texture
	OuterRadius       opt.T[float64]
	InnerRadius       opt.T[float64]
	CastShadow        opt.T[bool]
}

// CreateAmbientLight creates a new ambient light and appends it to the root of
// the scene.
func (s *Scene) CreateAmbientLight(info AmbientLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceAmbientLight(node, info)
	return node
}

// PlaceAmbientLight places an ambient light on the provided node.
func (s *Scene) PlaceAmbientLight(node *hierarchy.Node, info AmbientLightInfo) *graphics.AmbientLight {
	light := s.gfxScene.CreateAmbientLight(graphics.AmbientLightInfo{
		Position:          dprec.ZeroVec3(),
		InnerRadius:       25000.0,
		OuterRadius:       25000.0,
		ReflectionTexture: info.ReflectionTexture,
		RefractionTexture: info.RefractionTexture,
		CastShadow:        info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(AmbientLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
	return light
}

// PointLightInfo contains the information required to create a point light.
type PointLightInfo struct {
	EmitColor    opt.T[dprec.Vec3]
	EmitDistance opt.T[float64]
	CastShadow   opt.T[bool]
}

// CreatePointLight creates a new point light and appends it to the root of the
// scene.
func (s *Scene) CreatePointLight(info PointLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlacePointLight(node, info)
	return node
}

// PlacePointLight places a point light on the provided node.
func (s *Scene) PlacePointLight(node *hierarchy.Node, info PointLightInfo) *graphics.PointLight {
	light := s.gfxScene.CreatePointLight(graphics.PointLightInfo{
		Position:   dprec.ZeroVec3(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:  info.EmitDistance.ValueOrDefault(20.0),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(PointLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
	return light
}

// SpotLightInfo contains the information required to create a spot light.
type SpotLightInfo struct {
	EmitColor          opt.T[dprec.Vec3]
	EmitDistance       opt.T[float64]
	EmitOuterConeAngle opt.T[dprec.Angle]
	EmitInnerConeAngle opt.T[dprec.Angle]
	CastShadow         opt.T[bool]
}

// CreateSpotLight creates a new spot light and appends it to the root of the
// scene.
func (s *Scene) CreateSpotLight(info SpotLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceSpotLight(node, info)
	return node
}

// PlaceSpotLight places a spot light on the provided node.
func (s *Scene) PlaceSpotLight(node *hierarchy.Node, info SpotLightInfo) *graphics.SpotLight {
	light := s.gfxScene.CreateSpotLight(graphics.SpotLightInfo{
		Position:           dprec.ZeroVec3(),
		Rotation:           dprec.IdentityQuat(),
		EmitColor:          info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:          info.EmitDistance.ValueOrDefault(20.0),
		EmitOuterConeAngle: info.EmitOuterConeAngle.ValueOrDefault(dprec.Degrees(60)),
		EmitInnerConeAngle: info.EmitInnerConeAngle.ValueOrDefault(dprec.Degrees(30)),
		CastShadow:         info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(SpotLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
	return light
}

// DirectionalLightInfo contains the information required to create a
// directional light.
type DirectionalLightInfo struct {
	EmitColor  opt.T[dprec.Vec3]
	CastShadow opt.T[bool]
}

// CreateDirectionalLight creates a new directional light and appends it to the
// root of the scene.
func (s *Scene) CreateDirectionalLight(info DirectionalLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceDirectionalLight(node, info)
	return node
}

// PlaceDirectionalLight places a directional light on the provided node.
func (s *Scene) PlaceDirectionalLight(node *hierarchy.Node, info DirectionalLightInfo) *graphics.DirectionalLight {
	light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
		Position:   dprec.ZeroVec3(),
		Rotation:   dprec.IdentityQuat(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(DirectionalLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
	return light
}
