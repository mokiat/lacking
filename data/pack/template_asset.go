package pack

import (
	"fmt"
	"strings"

	"github.com/mokiat/lacking/data/asset"
)

type SaveUITemplateAssetAction struct {
	locator          AssetLocator
	uri              string
	templateProvider UITemplateProvider
}

func (a *SaveUITemplateAssetAction) Describe() string {
	return fmt.Sprintf("save_ui_template_asset(uri: %q)", a.uri)
}

func (a *SaveUITemplateAssetAction) Run() error {
	template := a.templateProvider.Template()

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

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeUITemplate(out, templateAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}
