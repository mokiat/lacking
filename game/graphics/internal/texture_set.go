package internal

import (
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
)

func NewShaderTextureSet(shader *lsl.Shader) TextureSet {
	textureBlock, ok := shader.FindTextureBlock()
	if !ok {
		return TextureSet{}
	}

	var names [8]string
	for i := range min(8, len(textureBlock.Fields)) {
		names[i] = textureBlock.Fields[i].Name
	}

	return TextureSet{
		names: names,
	}
}

type TextureSet struct {
	names [8]string

	textures [8]render.Texture
	samplers [8]render.Sampler
}

func (t *TextureSet) TextureCount() int {
	count := 0
	for i, texture := range t.textures {
		if texture != nil {
			count = i + 1
		}
	}
	return count
}

func (t *TextureSet) TextureAt(index int) render.Texture {
	return t.textures[index]
}

func (t *TextureSet) Texture(name string) render.Texture {
	if index, ok := t.findIndex(name); ok {
		return t.textures[index]
	}
	return nil
}

func (t *TextureSet) SetTexture(name string, texture render.Texture) {
	if index, ok := t.findIndex(name); ok {
		t.textures[index] = texture
	}
}

func (t *TextureSet) SamplerAt(index int) render.Sampler {
	return t.samplers[index]
}

func (t *TextureSet) Sampler(name string) render.Sampler {
	if index, ok := t.findIndex(name); ok {
		return t.samplers[index]
	}
	return nil
}

func (t *TextureSet) SetSampler(name string, sampler render.Sampler) {
	if index, ok := t.findIndex(name); ok {
		t.samplers[index] = sampler
	}
}

func (t *TextureSet) findIndex(name string) (int, bool) {
	for i := range t.names {
		if t.names[i] == name {
			return i, true
		}
	}
	return 0, false
}
