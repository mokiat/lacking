package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/debug/log"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/newasset/mdl"
	"golang.org/x/sync/errgroup"
)

var modelProviders = make(map[string]Provider[*mdl.Model])

func Run(storage asset.Storage, formatter asset.Formatter) error {
	registry, err := asset.NewRegistry(storage, formatter)
	if err != nil {
		return fmt.Errorf("error creating registry: %w", err)
	}

	var g errgroup.Group

	for name, modelProvider := range modelProviders {
		g.Go(func() error {
			log.Info("Model %q - processing", name)

			digest, err := StringDigest(modelProvider)
			if err != nil {
				return fmt.Errorf("error calculating model %q digest: %w", name, err)
			}

			resource := registry.ResourceByName(name)
			if resource != nil && resource.SourceDigest() == digest {
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

			if resource == nil {
				log.Info("Model %q - creating", name)
				resource, err = registry.CreateResource(name, modelAsset)
				if err != nil {
					return fmt.Errorf("error creating resource: %w", err)
				}
			} else {
				log.Info("Model %q - updating", name)
				if err := resource.SaveContent(modelAsset); err != nil {
					return fmt.Errorf("error saving resource: %w", err)
				}
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
