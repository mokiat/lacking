package opengl

type Texture struct {
	id uint32
}

func (t *Texture) ID() uint32 {
	return t.id
}
