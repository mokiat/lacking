package asset

type Fragment struct {
	Dependencies      []string           `json:"dependencies,omitempty"`
	Nodes             []Node             `json:"nodes,omitempty"`
	PointLights       []PointLight       `json:"point_lights,omitempty"`
	DirectionalLights []DirectionalLight `json:"directional_lights,omitempty"`
}
