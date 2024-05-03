package mdl

func NewSky() *Sky {
	return &Sky{}
}

type Sky struct {
	material *Material
}

func (s *Sky) Material() *Material {
	return s.material
}

func (s *Sky) SetMaterial(material *Material) {
	s.material = material
}
