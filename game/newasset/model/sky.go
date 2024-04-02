package model

type SkyLayerable interface {
	AddSkyLayer(layer SkyLayer)
}

type BaseSkyLayerable struct {
	layers []SkyLayer
}

func (b *BaseSkyLayerable) AddSkyLayer(layer SkyLayer) {
	b.layers = append(b.layers, layer)
}

type Sky struct {
	BaseNode
	BaseSkyLayerable
}

type SkyLayer struct {
	BaseBlendable
	BasePropertyHolder
	BaseTextureHolder
	BaseShadable
}
