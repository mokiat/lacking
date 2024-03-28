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
			log.Info("Processing fragment %q", name)

			digest, err := digestString(fragment)
			if err != nil {
				return fmt.Errorf("error calculating fragment %q digest: %w", name, err)
			}

			resource := registry.ResourceByName(name)
			if resource != nil && resource.SourceDigest() == digest {
				log.Info("Resource %q is up to date", name)
				return nil
			}

			fragmentModel, err := fragment.Get()
			if err != nil {
				return fmt.Errorf("error getting fragment %q: %w", name, err)
			}

			fragmentAsset, err := fragmentModel.ToAsset()
			if err != nil {
				return fmt.Errorf("error converting fragment %q to asset: %w", name, err)
			}

			if resource == nil {
				log.Info("Resource %q needs to be created", name)
				resource, err = registry.CreateResource(name, fragmentAsset)
				if err != nil {
					return fmt.Errorf("error creating resource: %w", err)
				}
			} else {
				log.Info("Resource %q needs updating", name)
				if err := resource.SaveContent(fragmentAsset); err != nil {
					return fmt.Errorf("error saving resource: %w", err)
				}
			}

			if err := resource.SetSourceDigest(digest); err != nil {
				return fmt.Errorf("error setting resource digest: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
