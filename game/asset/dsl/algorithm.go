package dsl

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/filter"
	"github.com/mokiat/lacking/storage/chunked"
	"golang.org/x/sync/errgroup"
)

// Run runs the DSL algorithm on the provided registry. If modelNames
// is not empty, only the models with the provided names will be processed.
func Run(storage chunked.Storage, pathFilter filter.Func[string]) error {
	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for path, modelProvider := range resourceProviders {
		if !pathFilter(path) {
			continue // skip this one
		}
		g.Go(func() error {
			if err := processAsset(storage, path, modelProvider); err != nil {
				return fmt.Errorf("error processing asset %q: %w", path, err)
			}
			return nil
		})
	}

	return g.Wait()
}

func processAsset(storage chunked.Storage, path string, provider Provider[any]) error {
	startTime := time.Now()

	digest, err := StringDigest(provider)
	if err != nil {
		return fmt.Errorf("error calculating new digest: %w", err)
	}

	asset := chunked.NewAsset(storage, path)
	sourceDigest, err := retrieveSourceDigest(asset)
	if err != nil {
		return fmt.Errorf("error retrieving old digest: %w", err)
	}

	if sourceDigest == digest {
		logger.Info("Asset skipped",
			slog.String("path", path),
			slog.String("duration", time.Since(startTime).String()),
		)
		return nil
	}

	chunkList := ds.NewList[chunked.Chunk](1)
	chunkList.Add(chunked.FromValue(genChunkID, &genChunk{
		Digest: digest,
	}))

	resource, err := provider.Get()
	if err != nil {
		return fmt.Errorf("provider failed to produce asset: %w", err)
	}
	for _, converter := range registeredConverters {
		if err := converter.Convert(chunkList, resource); err != nil {
			return fmt.Errorf("converter %T failed to convert resource: %w", converter, err)
		}
	}

	chunks := chunked.ChunkList(chunkList.Unbox())
	if err := asset.Write(chunks); err != nil {
		return fmt.Errorf("error writing chunks: %w", err)
	}

	logger.Info("Asset updated",
		slog.String("path", path),
		slog.Int("chunks", len(chunks)),
		slog.String("duration", time.Since(startTime).String()),
	)
	return nil
}

func retrieveSourceDigest(asset *chunked.Asset) (string, error) {
	var holder genChunkHolder
	if err := asset.Read(&holder); err != nil {
		if errors.Is(err, chunked.ErrNotFound) {
			return "", nil // no digest found
		}
		return "", fmt.Errorf("error reading asset: %w", err)
	}
	if holder.Gen == nil {
		return "", nil // no digest found
	}
	return holder.Gen.Digest, nil
}
