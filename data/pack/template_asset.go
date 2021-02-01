package pack

import (
	"fmt"
	"hash"
	"strings"

	"github.com/mokiat/lacking/data/asset"
)

func SaveUITemplateAsset(uri string, templateProvider UITemplateProvider) *SaveUITemplateAssetAction {
	return &SaveUITemplateAssetAction{
		uri:              uri,
		templateProvider: templateProvider,
	}
}

var _ Action = (*SaveUITemplateAssetAction)(nil)

type SaveUITemplateAssetAction struct {
	uri              string
	templateProvider UITemplateProvider
}

func (a *SaveUITemplateAssetAction) Describe() string {
	return fmt.Sprintf("save_ui_template_asset(uri: %q)", a.uri)
}

func (a *SaveUITemplateAssetAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "save_ui_template_asset", HashableParams{
		"uri":      a.uri,
		"template": a.templateProvider,
	})
}

func (a *SaveUITemplateAssetAction) Run(ctx *Context) error {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	template, err := a.templateProvider.Template(ctx)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	var createNodeAsset func(template UITemplate) asset.UINode
	createNodeAsset = func(template UITemplate) asset.UINode {
		result := asset.UINode{
			Name: template.Name,
		}
		for attrName, attrValue := range template.Attributes {
			if strings.HasPrefix(attrName, "layout-") {
				result.LayoutAttributes = append(result.LayoutAttributes, asset.UIAttribute{
					Name:  strings.TrimPrefix(attrName, "layout-"),
					Value: attrValue,
				})
			} else {
				result.Attributes = append(result.Attributes, asset.UIAttribute{
					Name:  attrName,
					Value: attrValue,
				})
			}
		}
		for _, child := range template.Children {
			result.Children = append(result.Children, createNodeAsset(child))
		}
		return result
	}

	templateAsset := &asset.UITemplate{
		Root: createNodeAsset(*template),
	}

	return ctx.IO(func(storage Storage) error {
		out, err := storage.CreateAsset(a.uri)
		if err != nil {
			return err
		}
		defer out.Close()

		if err := asset.EncodeUITemplate(out, templateAsset); err != nil {
			return fmt.Errorf("failed to encode asset: %w", err)
		}
		return nil
	})
}
