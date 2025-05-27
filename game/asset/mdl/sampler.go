package mdl

import (
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
)

const (
	WrapModeClamp          WrapMode = shadingdto.WrapModeClamp
	WrapModeRepeat         WrapMode = shadingdto.WrapModeRepeat
	WrapModeMirroredRepeat WrapMode = shadingdto.WrapModeMirroredRepeat
)

type WrapMode = shadingdto.WrapMode

const (
	FilterModeNearest     FilterMode = shadingdto.FilterModeNearest
	FilterModeLinear      FilterMode = shadingdto.FilterModeLinear
	FilterModeAnisotropic FilterMode = shadingdto.FilterModeAnisotropic
)

type FilterMode = shadingdto.FilterMode

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
