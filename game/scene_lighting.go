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
func (s *Scene) CreateAmbientLight(info AmbientLightInfo) hierarchy.NodeID {
	nodeID := s.Hierarchy().CreateNode()
	s.PlaceAmbientLight(nodeID, info)
	return nodeID
}

// PlaceAmbientLight places an ambient light on the provided node.
func (s *Scene) PlaceAmbientLight(nodeID hierarchy.NodeID, info AmbientLightInfo) *graphics.AmbientLight {
	light := s.gfxScene.CreateAmbientLight(graphics.AmbientLightInfo{
		Position:          dprec.ZeroVec3(),
		InnerRadius:       25000.0,
		OuterRadius:       25000.0,
		ReflectionTexture: info.ReflectionTexture,
		RefractionTexture: info.RefractionTexture,
		CastShadow:        info.CastShadow.ValueOrDefault(false),
	})
	s.ambientLightBindingSet.Bind(nodeID, light)
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
func (s *Scene) CreatePointLight(info PointLightInfo) hierarchy.NodeID {
	nodeID := s.Hierarchy().CreateNode()
	s.PlacePointLight(nodeID, info)
	return nodeID
}

// PlacePointLight places a point light on the provided node.
func (s *Scene) PlacePointLight(nodeID hierarchy.NodeID, info PointLightInfo) *graphics.PointLight {
	light := s.gfxScene.CreatePointLight(graphics.PointLightInfo{
		Position:   dprec.ZeroVec3(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:  info.EmitDistance.ValueOrDefault(20.0),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	s.pointLightBindingSet.Bind(nodeID, light)
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
func (s *Scene) CreateSpotLight(info SpotLightInfo) hierarchy.NodeID {
	nodeID := s.Hierarchy().CreateNode()
	s.PlaceSpotLight(nodeID, info)
	return nodeID
}

// PlaceSpotLight places a spot light on the provided node.
func (s *Scene) PlaceSpotLight(nodeID hierarchy.NodeID, info SpotLightInfo) *graphics.SpotLight {
	light := s.gfxScene.CreateSpotLight(graphics.SpotLightInfo{
		Position:           dprec.ZeroVec3(),
		Rotation:           dprec.IdentityQuat(),
		EmitColor:          info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:          info.EmitDistance.ValueOrDefault(20.0),
		EmitOuterConeAngle: info.EmitOuterConeAngle.ValueOrDefault(dprec.Degrees(60)),
		EmitInnerConeAngle: info.EmitInnerConeAngle.ValueOrDefault(dprec.Degrees(30)),
		CastShadow:         info.CastShadow.ValueOrDefault(false),
	})
	s.spotLightBindingSet.Bind(nodeID, light)
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
func (s *Scene) CreateDirectionalLight(info DirectionalLightInfo) hierarchy.NodeID {
	nodeID := s.Hierarchy().CreateNode()
	s.PlaceDirectionalLight(nodeID, info)
	return nodeID
}

// PlaceDirectionalLight places a directional light on the provided node.
func (s *Scene) PlaceDirectionalLight(nodeID hierarchy.NodeID, info DirectionalLightInfo) *graphics.DirectionalLight {
	light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
		Position:   dprec.ZeroVec3(),
		Rotation:   dprec.IdentityQuat(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	s.directionalLightBindingSet.Bind(nodeID, light)
	return light
}
