package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/debug/log"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/newasset/model"
	"golang.org/x/sync/errgroup"
)

var fragments = make(map[string]Provider[*model.Fragment])

func Run(storage asset.Storage, formatter asset.Formatter) error {
	registry, err := asset.NewRegistry(storage, formatter)
	if err != nil {
		return fmt.Errorf("error creating registry: %w", err)
	}

	var g errgroup.Group

	for name, fragment := range fragments {
		g.Go(func() error {
			log.Info("Fragment %q - processing", name)

			digest, err := digestString(fragment)
			if err != nil {
				return fmt.Errorf("error calculating fragment %q digest: %w", name, err)
			}

			resource := registry.ResourceByName(name)
			if resource != nil && resource.SourceDigest() == digest {
				log.Info("Fragment %q - up to date", name)
				log.Info("Fragment %q - done", name)
				return nil
			}

			log.Info("Fragment %q - building", name)
			fragmentModel, err := fragment.Get()
			if err != nil {
				return fmt.Errorf("error getting fragment %q: %w", name, err)
			}

			fragmentAsset, err := fragmentModel.ToAsset()
			if err != nil {
				return fmt.Errorf("error converting fragment %q to asset: %w", name, err)
			}

			if resource == nil {
				log.Info("Fragment %q - creating", name)
				resource, err = registry.CreateResource(name, fragmentAsset)
				if err != nil {
					return fmt.Errorf("error creating resource: %w", err)
				}
			} else {
				log.Info("Fragment %q - updating", name)
				if err := resource.SaveContent(fragmentAsset); err != nil {
					return fmt.Errorf("error saving resource: %w", err)
				}
			}

			if err := resource.SetSourceDigest(digest); err != nil {
				return fmt.Errorf("error setting resource digest: %w", err)
			}

			log.Info("Fragment %q - done", name)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
