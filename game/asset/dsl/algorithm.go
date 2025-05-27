package dsl

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/mokiat/gog/filter"
	"github.com/mokiat/lacking/game/asset/dto/gendto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
	"golang.org/x/sync/errgroup"
)

var modelProviders = make(map[string]Provider[*mdl.Model])

// Run runs the DSL algorithm on the provided registry. If modelNames
// is not empty, only the models with the provided names will be processed.
func Run(storage chunked.Storage, pathFilter filter.Func[string]) error {
	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for path, modelProvider := range modelProviders {
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

func processAsset(storage chunked.Storage, path string, provider Provider[*mdl.Model]) error {
	startTime := time.Now()

	digest, err := StringDigest(provider)
	if err != nil {
		return fmt.Errorf("error calculating new digest: %w", err)
	}

	resource := chunked.NewAsset(storage, path)
	sourceDigest, err := retrieveSourceDigest(resource)
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

	model, err := provider.Get()
	if err != nil {
		return fmt.Errorf("provider failed to produce asset: %w", err)
	}

	var chunks chunked.ChunkList
	chunks = append(chunks, chunked.FromValue(gendto.GenChunkID, &gendto.GenChunk{
		Digest: digest,
	}))

	var appliedConverters []string
	for name, converter := range Converters() {
		if converter.CanConvert(model) {
			chunk, err := converter.Convert(model)
			if err != nil {
				return fmt.Errorf("converter %q error: %w", name, err)
			}
			chunks = append(chunks, chunk)
			appliedConverters = append(appliedConverters, name)
		}
	}

	if err := resource.Write(chunks); err != nil {
		return fmt.Errorf("error writing chunks: %w", err)
	}

	logger.Info("Asset updated",
		slog.String("path", path),
		slog.Int("converters", len(appliedConverters)),
		slog.String("duration", time.Since(startTime).String()),
	)
	return nil
}

func retrieveSourceDigest(resource *chunked.Asset) (string, error) {
	var holder gendto.GenChunkHolder
	if err := resource.Read(&holder); err != nil {
		if errors.Is(err, chunked.ErrNotFound) {
			return "", nil // no digest found
		}
		return "", fmt.Errorf("error reading resource: %w", err)
	}
	if holder.Gen == nil {
		return "", nil // no digest found
	}
	return holder.Gen.Digest, nil
}
