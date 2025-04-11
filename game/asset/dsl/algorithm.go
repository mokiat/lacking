package dsl

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/mdl"
	"golang.org/x/sync/errgroup"
)

var modelProviders = make(map[string]Provider[*mdl.Model])

// Run runs the DSL algorithm on the provided registry. If modelNames
// is not empty, only the models with the provided names will be processed.
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
			logger.Info("Processing model",
				slog.String("name", name),
			)

			digest, err := StringDigest(modelProvider)
			if err != nil {
				return fmt.Errorf("error calculating model %q digest: %w", name, err)
			}

			if resource.SourceDigest() == digest {
				logger.Info("Model up to date",
					slog.String("name", name),
				)
				logger.Info("Model processed",
					slog.String("name", name),
				)
				return nil
			}

			logger.Info("Building model",
				slog.String("name", name),
			)
			model, err := modelProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting model %q: %w", name, err)
			}

			modelAsset, err := mdl.NewConverter(model).Convert()
			if err != nil {
				return fmt.Errorf("error converting model %q to asset: %w", name, err)
			}

			logger.Info("Updating model",
				slog.String("name", name),
			)
			if err := resource.SaveContent(modelAsset); err != nil {
				return fmt.Errorf("error saving resource: %w", err)
			}
			if err := resource.SetSourceDigest(digest); err != nil {
				return fmt.Errorf("error setting resource digest: %w", err)
			}

			logger.Info("Model processed",
				slog.String("name", name),
			)
			return nil
		})
	}

	return g.Wait()
}
