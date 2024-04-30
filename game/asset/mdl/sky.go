package mdl

type Sky struct {
	BaseNode

	material *Material
}

func (s *Sky) Material() *Material {
	return s.material
}

func (s *Sky) SetMaterial(material *Material) {
	s.material = material
}
