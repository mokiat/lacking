package mdl

func NewSky() *Sky {
	return &Sky{
		Object: NewObject(),
	}
}

type Sky struct {
	*Object
	material *Material
}

func (s *Sky) Material() *Material {
	return s.material
}

func (s *Sky) SetMaterial(material *Material) {
	s.material = material
}
