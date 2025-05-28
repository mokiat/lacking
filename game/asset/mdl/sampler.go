package mdl

import "github.com/mokiat/lacking/game/asset/dto"

const (
	WrapModeClamp          WrapMode = dto.WrapModeClamp
	WrapModeRepeat         WrapMode = dto.WrapModeRepeat
	WrapModeMirroredRepeat WrapMode = dto.WrapModeMirroredRepeat
)

type WrapMode = dto.WrapMode

const (
	FilterModeNearest     FilterMode = dto.FilterModeNearest
	FilterModeLinear      FilterMode = dto.FilterModeLinear
	FilterModeAnisotropic FilterMode = dto.FilterModeAnisotropic
)

type FilterMode = dto.FilterMode

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

func NewSampler() *Sampler {
	return &Sampler{
		Object: NewObject(),
	}
}

type Sampler struct {
	*Object
	BaseTextureReferrer
	BaseWrappable
	BaseFilterable
	BaseMipmappable
}
