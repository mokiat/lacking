package pack

import (
	"encoding/xml"
	"fmt"
	"hash"
	"sync"
)

func OpenUITemplateResource(uri string) *OpenUITemplateResourceAction {
	return &OpenUITemplateResourceAction{
		uri: uri,
	}
}

var _ UITemplateProvider = (*OpenUITemplateResourceAction)(nil)

type OpenUITemplateResourceAction struct {
	uri string

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *UITemplate
}

func (a *OpenUITemplateResourceAction) Describe() string {
	return fmt.Sprintf("open_ui_template_resource(uri: %q)", a.uri)
}

func (a *OpenUITemplateResourceAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "open_ui_template_resource", HashableParams{
		"uri": a.uri,
	})
}

func (a *OpenUITemplateResourceAction) Template(ctx *Context) (*UITemplate, error) {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	a.resultMutex.Lock()
	defer a.resultMutex.Unlock()

	digest, err := CalculateDigest(a)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}
	if EqualDigests(digest, a.resultDigest) {
		return a.result, nil
	}

	result, err := a.run(ctx)
	if err != nil {
		return nil, err
	}

	a.result = result
	a.resultDigest = digest
	return result, nil
}

func (a *OpenUITemplateResourceAction) run(ctx *Context) (*UITemplate, error) {
	var template *UITemplate
	readTemplate := func(storage Storage) error {
		in, err := storage.OpenResource(a.uri)
		if err != nil {
			return err
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
		template = &rootTemplate
		return nil
	}

	if err := ctx.IO(readTemplate); err != nil {
		return nil, err
	}
	return template, nil
}
