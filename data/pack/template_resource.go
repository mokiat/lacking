package pack

import (
	"encoding/xml"
	"fmt"
)

type OpenUITemplateResourceAction struct {
	locator  ResourceLocator
	uri      string
	template *UITemplate
}

func (a *OpenUITemplateResourceAction) Describe() string {
	return fmt.Sprintf("open_ui_template_resource(uri: %q)", a.uri)
}

func (a *OpenUITemplateResourceAction) Template() *UITemplate {
	if a.template == nil {
		panic("reading data from unprocessed action")
	}
	return a.template
}

func (a *OpenUITemplateResourceAction) Run() error {
	in, err := a.locator.Open(a.uri)
	if err != nil {
		return fmt.Errorf("failed to open ui template resource %q: %w", a.uri, err)
	}
	defer in.Close()

	decoder := xml.NewDecoder(in)

	var parseTemplate func(decoder *xml.Decoder, element xml.StartElement) (UITemplate, error)
	parseTemplate = func(decoder *xml.Decoder, element xml.StartElement) (UITemplate, error) {
		result := UITemplate{
			Name:       element.Name.Local,
			Attributes: make(map[string]string),
		}
		for _, attr := range element.Attr {
			result.Attributes[attr.Name.Local] = attr.Value
		}
		for {
			token, err := decoder.Token()
			if err != nil {
				return UITemplate{}, err
			}
			switch token := token.(type) {
			case xml.StartElement:
				child, err := parseTemplate(decoder, token)
				if err != nil {
					return UITemplate{}, err
				}
				result.Children = append(result.Children, child)
			case xml.EndElement:
				return result, nil
			}
		}
	}

	var element xml.StartElement
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if el, ok := token.(xml.StartElement); ok {
			element = el
			break
		}
	}

	rootTemplate, err := parseTemplate(decoder, element)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	a.template = &rootTemplate
	return nil
}
