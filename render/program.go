package render

// ProgramMarker marks a type as being a Program.
type ProgramMarker interface {
	_isProgramType()
}

// Program represents a graphics program.
type Program interface {
	ProgramMarker
	Resource
}

// ProgramCodeMarker marks a type as being a ProgramCode.
type ProgramCodeMarker interface {
	_isProgramCodeType() // ensures interface uniqueness
}

// ProgramCode represents the shader language source code for a program.
//
// This can differ depending on the API implementation. It could be a single
// string or a struct with multiple strings for each shader stage.
type ProgramCode interface {
	ProgramCodeMarker
}

// ProgramInfo represents the information needed to create a Program.
type ProgramInfo struct {

	// Label specifies a human-readable label for the program. Intended for
	// debugging and logging purposes only.
	Label string

	// SourceCode specifies the source code for the program.
	SourceCode ProgramCode

	// TextureBindings specifies the texture bindings for the program, in case
	// the implementation does not support shader-specified bindings.
	TextureBindings []TextureBinding

	// UniformBindings specifies the uniform bindings for the program, in case
	// the implementation does not support shader-specified bindings.
	UniformBindings []UniformBinding
}

// NewTextureBinding creates a new TextureBinding with the specified name and
// index.
func NewTextureBinding(name string, index uint) TextureBinding {
	return TextureBinding{
		Name:  name,
		Index: index,
	}
}

// TextureBinding represents a texture binding for a program.
type TextureBinding struct {

	// Name specifies the name of the texture in the shader.
	Name string

	// Index specifies the binding index.
	Index uint
}

// NewUniformBinding creates a new UniformBinding with the specified name and
// index.
func NewUniformBinding(name string, index uint) UniformBinding {
	return UniformBinding{
		Name:  name,
		Index: index,
	}
}

// UniformBinding represents a uniform binding for a program.
type UniformBinding struct {

	// Name specifies the name of the uniform object in the shader.
	Name string

	// Index specifies the binding index.
	Index uint
}
