package dsl

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/mokiat/lacking/game/asset/gendto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
	"golang.org/x/sync/errgroup"
)

var modelProviders = make(map[string]Provider[*mdl.Model])

// Run runs the DSL algorithm on the provided registry. If modelNames
// is not empty, only the models with the provided names will be processed.
func Run(storage chunked.Storage, pickedPaths []string) error {
	var g errgroup.Group

	for path, modelProvider := range modelProviders {
		if len(pickedPaths) > 0 {
			if !slices.Contains(pickedPaths, path) {
				continue // skip this one
			}
		}

		resource := chunked.NewAsset(storage, path)

		g.Go(func() error {
			logger.Info("Processing model",
				slog.String("path", path),
			)

			digest, err := StringDigest(modelProvider)
			if err != nil {
				return fmt.Errorf("error calculating model %q digest: %w", path, err)
			}

			sourceDigest, err := retrieveSourceDigest(resource)
			if err != nil {
				return fmt.Errorf("error retrieving model %q source digest: %w", path, err)
			}

			if sourceDigest == digest {
				logger.Info("Model up to date",
					slog.String("path", path),
				)
				logger.Info("Model processed",
					slog.String("path", path),
				)
				return nil
			}

			logger.Info("Building model",
				slog.String("path", path),
			)
			model, err := modelProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting model %q: %w", path, err)
			}

			modelAsset, err := mdl.NewConverter(model).Convert()
			if err != nil {
				return fmt.Errorf("error converting model %q to asset: %w", path, err)
			}
			modelAsset.Gen = &gendto.GenChunk{
				Digest: digest,
			}

			logger.Info("Updating model",
				slog.String("path", path),
			)
			if err := resource.Write(modelAsset); err != nil {
				return fmt.Errorf("error saving resource: %w", err)
			}

			logger.Info("Model processed",
				slog.String("path", path),
			)
			return nil
		})
	}

	return g.Wait()
}

func retrieveSourceDigest(resource *chunked.Asset) (string, error) {
	type digestHolder struct {
		*gendto.GenChunk
	}
	var holder digestHolder
	if err := resource.Read(&holder); err != nil {
		if errors.Is(err, chunked.ErrNotFound) {
			return "", nil // no digest found
		}
		return "", fmt.Errorf("error reading resource: %w", err)
	}
	if holder.GenChunk == nil {
		return "", nil // no digest found
	}
	return holder.Digest, nil
}
