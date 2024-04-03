package mdl

type Blendable interface {
	Blending() bool
	SetBlending(blending bool)
}

type BaseBlendable struct {
	blending bool
}

func (b *BaseBlendable) Blending() bool {
	return b.blending
}

func (b *BaseBlendable) SetBlending(blending bool) {
	b.blending = blending
}

type PropertyHolder interface {
	Property(name string) any
	SetProperty(name string, value any)
}

type BasePropertyHolder struct {
	properties map[string]any
}

func (b *BasePropertyHolder) Property(name string) any {
	if b.properties == nil {
		return nil
	}
	return b.properties[name]
}

func (b *BasePropertyHolder) SetProperty(name string, value any) {
	if b.properties == nil {
		b.properties = make(map[string]any)
	}
	b.properties[name] = value
}

type TextureHolder interface {
	// TODO
}

type BaseTextureHolder struct {
	// TODO
}

type Shadable interface {
	Shader() *Shader
	SetShader(shader *Shader)
}

type BaseShadable struct {
	shader *Shader
}

func (b *BaseShadable) Shader() *Shader {
	return b.shader
}

func (b *BaseShadable) SetShader(shader *Shader) {
	b.shader = shader
}
