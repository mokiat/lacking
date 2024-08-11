package mdl

import "github.com/mokiat/lacking/game/asset"

const (
	WrapModeClamp          WrapMode = asset.WrapModeClamp
	WrapModeRepeat         WrapMode = asset.WrapModeRepeat
	WrapModeMirroredRepeat WrapMode = asset.WrapModeMirroredRepeat
)

type WrapMode = asset.WrapMode

const (
	FilterModeNearest     FilterMode = asset.FilterModeNearest
	FilterModeLinear      FilterMode = asset.FilterModeLinear
	FilterModeAnisotropic FilterMode = asset.FilterModeAnisotropic
)

type FilterMode = asset.FilterMode

type TextureReferrer interface {
	Texture() *Texture
	SetTexture(texture *Texture)
}

type BaseTextureReferrer struct {
	texture *Texture
}

func (b *BaseTextureReferrer) Texture() *Texture {
	return b.texture
}

func (b *BaseTextureReferrer) SetTexture(texture *Texture) {
	b.texture = texture
}

type Wrappable interface {
	WrapMode() WrapMode
	SetWrapMode(mode WrapMode)
}

type BaseWrappable struct {
	wrapMode WrapMode
}

func (b *BaseWrappable) WrapMode() WrapMode {
	return b.wrapMode
}

func (b *BaseWrappable) SetWrapMode(mode WrapMode) {
	b.wrapMode = mode
}

type Filterable interface {
	FilterMode() FilterMode
	SetFilterMode(mode FilterMode)
}

type BaseFilterable struct {
	filterMode FilterMode
}

func (b *BaseFilterable) FilterMode() FilterMode {
	return b.filterMode
}

func (b *BaseFilterable) SetFilterMode(mode FilterMode) {
	b.filterMode = mode
}

type Mipmappable interface {
	Mipmapping() bool
	SetMipmapping(mipmapping bool)
}

type BaseMipmappable struct {
	mipmapping bool
}

func (b *BaseMipmappable) Mipmapping() bool {
	return b.mipmapping
}

func (b *BaseMipmappable) SetMipmapping(mipmapping bool) {
	b.mipmapping = mipmapping
}

type Sampler struct {
	BaseTextureReferrer
	BaseWrappable
	BaseFilterable
	BaseMipmappable
}
