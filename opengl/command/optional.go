package command

import "github.com/mokiat/lacking/opengl"

type optional struct {
	isSet bool
}

func (o optional) IsSet() bool {
	return o.isSet
}

func UnspecifiedBool() OptionalBool {
	return OptionalBool{}
}

func SpecifiedBool(value bool) OptionalBool {
	return OptionalBool{
		optional: optional{
			isSet: true,
		},
		value: value,
	}
}

type OptionalBool struct {
	optional
	value bool
}

func (b OptionalBool) Value() bool {
	return b.value
}

func UnspecifiedFloat32() OptionalFloat32 {
	return OptionalFloat32{}
}

func SpecifiedFloat32(value float32) OptionalFloat32 {
	return OptionalFloat32{
		optional: optional{
			isSet: false,
		},
		value: value,
	}
}

type OptionalFloat32 struct {
	optional
	value float32
}

func (o OptionalFloat32) Value() float32 {
	return o.value
}

func UnspecifiedUint32() OptionalUint32 {
	return OptionalUint32{}
}

func SpecifiedUint32(value uint32) OptionalUint32 {
	return OptionalUint32{
		optional: optional{
			isSet: true,
		},
		value: value,
	}
}

type OptionalUint32 struct {
	optional
	value uint32
}

func (o OptionalUint32) Value() uint32 {
	return o.value
}

func UnspecifiedClearColor() OptionalClearColor {
	return OptionalClearColor{}
}

func SpecifiedClearColor(value ClearColor) OptionalClearColor {
	return OptionalClearColor{
		optional: optional{
			isSet: true,
		},
		value: value,
	}
}

type OptionalClearColor struct {
	optional
	value ClearColor
}

func (o OptionalClearColor) Value() ClearColor {
	return o.value
}

func UnspecifiedFramebuffer() OptionalFramebuffer {
	return OptionalFramebuffer{}
}

func SpecifiedFramebuffer(value *opengl.Framebuffer) OptionalFramebuffer {
	return OptionalFramebuffer{
		optional: optional{
			isSet: true,
		},
		value: value,
	}
}

type OptionalFramebuffer struct {
	optional
	value *opengl.Framebuffer
}

func (o OptionalFramebuffer) Value() *opengl.Framebuffer {
	return o.value
}

func UnspecifiedArea() OptionalArea {
	return OptionalArea{}
}

func SpecifiedArea(value opengl.Area) OptionalArea {
	return OptionalArea{
		optional: optional{
			isSet: true,
		},
		value: value,
	}
}

type OptionalArea struct {
	optional
	value opengl.Area
}

func (o OptionalArea) Value() opengl.Area {
	return o.value
}
