package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/debug/log"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/newasset/model"
	"golang.org/x/sync/errgroup"
)

var sceneProviders = make(map[string]Provider[*model.Scene])

func Run(storage asset.Storage, formatter asset.Formatter) error {
	registry, err := asset.NewRegistry(storage, formatter)
	if err != nil {
		return fmt.Errorf("error creating registry: %w", err)
	}

	var g errgroup.Group

	for name, sceneProvider := range sceneProviders {
		g.Go(func() error {
			log.Info("Scene %q - processing", name)

			digest, err := digestString(sceneProvider)
			if err != nil {
				return fmt.Errorf("error calculating scene %q digest: %w", name, err)
			}

			resource := registry.ResourceByName(name)
			if resource != nil && resource.SourceDigest() == digest {
				log.Info("Scene %q - up to date", name)
				log.Info("Scene %q - done", name)
				return nil
			}

			log.Info("Scene %q - building", name)
			scene, err := sceneProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting scene %q: %w", name, err)
			}

			sceneAsset, err := model.NewConverter(scene).Convert()
			if err != nil {
				return fmt.Errorf("error converting scene %q to asset: %w", name, err)
			}

			if resource == nil {
				log.Info("Scene %q - creating", name)
				resource, err = registry.CreateResource(name, sceneAsset)
				if err != nil {
					return fmt.Errorf("error creating resource: %w", err)
				}
			} else {
				log.Info("Scene %q - updating", name)
				if err := resource.SaveContent(sceneAsset); err != nil {
					return fmt.Errorf("error saving resource: %w", err)
				}
			}

			if err := resource.SetSourceDigest(digest); err != nil {
				return fmt.Errorf("error setting resource digest: %w", err)
			}

			log.Info("Scene %q - done", name)
			return nil
		})
	}

	return g.Wait()
}
