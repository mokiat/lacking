package asset

import "github.com/mokiat/gomath/dprec"

type PointLight struct {
	NodeIndex     int32      `json:"node_index"`
	EmitColor     dprec.Vec3 `json:"emit_color"`
	EmitIntensity float64    `json:"emit_intensity"`
	EmitRange     float64    `json:"emit_range"`
}

type DirectionalLight struct {
	NodeIndex     int32      `json:"node_index"`
	EmitColor     dprec.Vec3 `json:"emit_color"`
	EmitIntensity float64    `json:"emit_intensity"`
	EmitRange     float64    `json:"emit_range"`
}
