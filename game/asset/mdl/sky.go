package mdl

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
	BasePropertyHolder
	BaseSamplerHolder
	BaseSkyLayerable
}

type SkyLayer struct {
	BaseBlendable
	BaseShadable
}
