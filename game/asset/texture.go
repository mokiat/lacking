package asset

const UnspecifiedIndex = int32(-1)

type TextureRef struct {
	TextureIndex int32
}

func (r TextureRef) Valid() bool {
	return r.TextureIndex >= 0
}
