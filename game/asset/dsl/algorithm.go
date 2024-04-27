package dsl

import (
	"fmt"
	"slices"

	"github.com/mokiat/lacking/debug/log"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/mdl"
	"golang.org/x/sync/errgroup"
)

var modelProviders = make(map[string]Provider[*mdl.Model])

func Run(registry *asset.Registry, modelNames []string) error {
	var g errgroup.Group

	for name, modelProvider := range modelProviders {
		if len(modelNames) > 0 {
			if !slices.Contains(modelNames, name) {
				continue // skip this one
			}
		}

		resource := registry.ResourceByName(name)
		if resource == nil {
			var err error
			resource, err = registry.CreateResource(name, asset.Model{})
			if err != nil {
				return fmt.Errorf("error creating resource: %w", err)
			}
		}

		g.Go(func() error {
			log.Info("Model %q - processing", name)

			digest, err := StringDigest(modelProvider)
			if err != nil {
				return fmt.Errorf("error calculating model %q digest: %w", name, err)
			}

			if resource.SourceDigest() == digest {
				log.Info("Model %q - up to date", name)
				log.Info("Model %q - done", name)
				return nil
			}

			log.Info("Model %q - building", name)
			model, err := modelProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting model %q: %w", name, err)
			}

			modelAsset, err := mdl.NewConverter(model).Convert()
			if err != nil {
				return fmt.Errorf("error converting model %q to asset: %w", name, err)
			}

			log.Info("Model %q - updating", name)
			if err := resource.SaveContent(modelAsset); err != nil {
				return fmt.Errorf("error saving resource: %w", err)
			}
			if err := resource.SetSourceDigest(digest); err != nil {
				return fmt.Errorf("error setting resource digest: %w", err)
			}

			log.Info("Model %q - done", name)
			return nil
		})
	}

	return g.Wait()
}
