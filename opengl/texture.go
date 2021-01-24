package opengl

// FIXME: REMOVE
func NewTexture(id uint32) *Texture {
	return &Texture{
		id: id,
	}
}

type Texture struct {
	id uint32
}

func (t *Texture) ID() uint32 {
	return t.id
}
