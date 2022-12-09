package physics

// MaterialInfo contains the data necessary to create a Material.
type MaterialInfo struct {
	FrictionCoefficient    float64
	RestitutionCoefficient float64
}

// Material represents the surface properties of an object.
type Material struct {
	frictionCoefficient    float64
	restitutionCoefficient float64
}

// FrictionCoefficient returns the friction coefficient of this material.
func (m *Material) FrictionCoefficient() float64 {
	return m.frictionCoefficient
}

// SetFrictionCoefficient changes the friction coefficient of this material.
func (m *Material) SetFrictionCoefficient(coefficient float64) {
	m.frictionCoefficient = coefficient
}

// RestitutionCoefficient returns the coefficient of restitution of
// this material.
func (m *Material) RestitutionCoefficient() float64 {
	return m.restitutionCoefficient
}

// SetRestitutionCoefficient changes the coefficient of restitution of this
// material.
func (m *Material) SetRestitutionCoefficient(coefficient float64) {
	m.restitutionCoefficient = coefficient
}
