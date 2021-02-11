package ui

import (
	"github.com/mokiat/lacking/data/asset"
)

type Template struct {
	ID               string
	Name             string
	Attributes       AttributeSet
	LayoutAttributes AttributeSet
	Children         []*Template
}

func buildTemplateFromAsset(node asset.UINode) *Template {
	result := &Template{
		ID:               findIDFromAsset(node.Attributes),
		Name:             node.Name,
		Attributes:       buildAttributeSetFromAsset(node.Attributes),
		LayoutAttributes: buildAttributeSetFromAsset(node.LayoutAttributes),
		Children:         make([]*Template, len(node.Children)),
	}
	for i, child := range node.Children {
		result.Children[i] = buildTemplateFromAsset(child)
	}
	return result
}

func findIDFromAsset(attributes []asset.UIAttribute) string {
	for _, attribute := range attributes {
		if attribute.Name == "id" {
			return attribute.Value
		}
	}
	return ""
}

func buildAttributeSetFromAsset(attributes []asset.UIAttribute) AttributeSet {
	entries := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		if attribute.Name != "id" {
			entries[attribute.Name] = attribute.Value
		}
	}
	return NewMapAttributeSet(entries)
}
