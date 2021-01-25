package asset

import "io"

type UITemplate struct {
	Root UINode
}

type UINode struct {
	Name             string
	Attributes       []UIAttribute
	LayoutAttributes []UIAttribute
	Children         []UINode
}

type UIAttribute struct {
	Name  string
	Value string
}

func EncodeUITemplate(out io.Writer, template *UITemplate) error {
	return Encode(out, template)
}

func DecodeUITemplate(in io.Reader, template *UITemplate) error {
	return Decode(in, template)
}
