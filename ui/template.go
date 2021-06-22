package ui

import (
	"encoding/xml"
	"io"
	"strings"
)

// Template represents a user interface template.
//
// It is a mechanism by which a hierarchy of controls can be
// stored offline in a declarative and quickly instantiated.
type Template struct {
	name             string
	attributes       AttributeSet
	layoutAttributes AttributeSet
	children         []*Template
}

// Name returns the name of the Control. This allows
// the build sequence to determine which Control Builder
// should be used.
func (t *Template) Name() string {
	return t.name
}

// Attributes returns the attributes that should be applied
// to the Control.
func (t *Template) Attributes() AttributeSet {
	return t.attributes
}

// LayoutAttributes returns the attributes that should be
// applied to the Control's layout configuration.
func (t *Template) LayoutAttributes() AttributeSet {
	return t.layoutAttributes
}

// Children returns any children that this template may
// have. Depending on the context, these children may represent
// other Controls or they may represents additional
// settings for the current Control.
func (t *Template) Children() []*Template {
	result := make([]*Template, len(t.children))
	for i, v := range t.children {
		result[i] = v
	}
	return result
}

func parseTemplate(in io.Reader) (*Template, error) {
	decoder := xml.NewDecoder(in)
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if rootElement, ok := token.(xml.StartElement); ok {
			return parseTemplateNode(decoder, rootElement)
		}
	}
}

func parseTemplateNode(decoder *xml.Decoder, element xml.StartElement) (*Template, error) {
	result := &Template{
		name: element.Name.Local,
	}

	entries := make(map[string]string)
	layoutEntries := make(map[string]string)
	for _, attr := range element.Attr {
		attrName := strings.ToLower(attr.Name.Local)
		switch {
		case strings.HasPrefix(attrName, "layout-"):
			layoutEntries[strings.TrimPrefix(attrName, "layout-")] = attr.Value
		default:
			entries[attrName] = attr.Value
		}
	}
	result.attributes = NewMapAttributeSet(entries)
	result.layoutAttributes = NewMapAttributeSet(layoutEntries)

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case xml.StartElement:
			childTemplate, err := parseTemplateNode(decoder, token)
			if err != nil {
				return nil, err
			}
			result.children = append(result.children, childTemplate)
		case xml.EndElement:
			return result, nil
		}
	}
}
