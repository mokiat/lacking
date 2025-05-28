package dsl

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/mokiat/gog/filter"
	"github.com/mokiat/lacking/game/asset/dto/gendto"
	"github.com/mokiat/lacking/storage/chunked"
	"golang.org/x/sync/errgroup"
)

var resourceProviders = make(map[string]Provider[Resource])

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

func processAsset(storage chunked.Storage, path string, provider Provider[Resource]) error {
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

	var chunks chunked.ChunkList
	chunks = append(chunks, chunked.FromValue(gendto.GenChunkID, &gendto.GenChunk{
		Digest: digest,
	}))

	resource, err := provider.Get()
	if err != nil {
		return fmt.Errorf("provider failed to produce asset: %w", err)
	}

	var appliedConverters []string
	for name, converter := range Converters() {
		if converter.CanConvert(resource) {
			chunk, err := converter.Convert(resource)
			if err != nil {
				return fmt.Errorf("converter %q error: %w", name, err)
			}
			chunks = append(chunks, chunk)
			appliedConverters = append(appliedConverters, name)
		}
	}

	if err := asset.Write(chunks); err != nil {
		return fmt.Errorf("error writing chunks: %w", err)
	}

	logger.Info("Asset updated",
		slog.String("path", path),
		slog.Int("converters", len(appliedConverters)),
		slog.String("duration", time.Since(startTime).String()),
	)
	return nil
}

func retrieveSourceDigest(asset *chunked.Asset) (string, error) {
	var holder gendto.GenChunkHolder
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
