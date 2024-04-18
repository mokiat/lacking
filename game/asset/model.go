package asset

import (
	"fmt"
	"io"

	newasset "github.com/mokiat/lacking/game/newasset"
)

type Model struct {
	Nodes             []newasset.Node
	GeometryShaders   []newasset.Shader
	ShadowShaders     []newasset.Shader
	ForwardShaders    []newasset.Shader
	Animations        []Animation
	Armatures         []newasset.Armature
	Textures          []newasset.Texture
	Materials         []newasset.Material
	MeshDefinitions   []MeshDefinition
	MeshInstances     []newasset.Mesh
	BodyMaterials     []newasset.BodyMaterial
	BodyDefinitions   []newasset.BodyDefinition
	BodyInstances     []newasset.Body
	PointLights       []newasset.PointLight
	SpotLights        []newasset.SpotLight
	DirectionalLights []newasset.DirectionalLight
}

func (m *Model) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, m)
}

func (m *Model) DecodeFrom(in io.Reader) error {
	return decodeResource(in, m)
}

func (m *Model) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return m.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (m *Model) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return m.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}
