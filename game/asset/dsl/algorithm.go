package dsl

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/filter"
	"github.com/mokiat/lacking/core/resource"
	"github.com/mokiat/lacking/storage/chunked"
	"golang.org/x/sync/errgroup"
)

// Run runs the DSL algorithm on the provided registry. If modelNames
// is not empty, only the models with the provided names will be processed.
func Run(store resource.Store, pathFilter filter.Func[string]) error {
	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for path, rawProvider := range rawResourceProviders {
		if !pathFilter(path) {
			continue // skip this one
		}
		g.Go(func() error {
			if err := processRawAsset(store, path, rawProvider); err != nil {
				return fmt.Errorf("error processing asset %q: %w", path, err)
			}
			return nil
		})
	}

	for path, modelProvider := range resourceProviders {
		if !pathFilter(path) {
			continue // skip this one
		}
		g.Go(func() error {
			if err := processAsset(store, path, modelProvider); err != nil {
				return fmt.Errorf("error processing asset %q: %w", path, err)
			}
			return nil
		})
	}

	return g.Wait()
}

func processAsset(store resource.Store, path string, provider Provider[any]) error {
	startTime := time.Now()

	currentSourceDigest, err := StringDigest(provider)
	if err != nil {
		return fmt.Errorf("error calculating new digest: %w", err)
	}

	previousSourceDigest, err := openSourceDigest(store, path)
	if err != nil {
		return fmt.Errorf("error retrieving old digest: %w", err)
	}

	if previousSourceDigest == currentSourceDigest {
		logger.Info("Asset skipped",
			slog.String("path", path),
			slog.String("duration", time.Since(startTime).String()),
		)
		return nil
	}

	resource, err := provider.Get()
	if err != nil {
		return fmt.Errorf("provider failed to produce asset: %w", err)
	}

	chunkList := ds.PreallocatedList[chunked.Chunk](1)
	for _, converter := range registeredConverters {
		if err := converter.Convert(chunkList, resource); err != nil {
			return fmt.Errorf("converter %T failed to convert resource: %w", converter, err)
		}
	}
	chunks := chunked.ChunkList(chunkList.Unbox())

	asset := chunked.NewAsset(store, path)
	if err := asset.Write(chunks); err != nil {
		return fmt.Errorf("error writing chunks: %w", err)
	}

	if err := saveSourceDigest(store, path, currentSourceDigest); err != nil {
		return fmt.Errorf("error saving source digest: %w", err)
	}

	logger.Info("Asset updated",
		slog.String("path", path),
		slog.Int("chunks", len(chunks)),
		slog.String("duration", time.Since(startTime).String()),
	)
	return nil
}

func processRawAsset(store resource.Store, path string, provider Provider[io.ReadCloser]) error {
	startTime := time.Now()

	currentSourceDigest, err := StringDigest(provider)
	if err != nil {
		return fmt.Errorf("error calculating new digest: %w", err)
	}

	previousSourceDigest, err := openSourceDigest(store, path)
	if err != nil {
		return fmt.Errorf("error retrieving old digest: %w", err)
	}

	if previousSourceDigest == currentSourceDigest {
		logger.Info("Asset skipped",
			slog.String("path", path),
			slog.String("duration", time.Since(startTime).String()),
		)
		return nil
	}

	in, err := provider.Get()
	if err != nil {
		return fmt.Errorf("provider failed to produce asset: %w", err)
	}
	defer in.Close()

	out, err := store.Create(path)
	if err != nil {
		return fmt.Errorf("error creating asset file: %w", err)
	}
	defer out.Close()

	size, err := io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying raw asset data: %w", err)
	}

	if err := saveSourceDigest(store, path, currentSourceDigest); err != nil {
		return fmt.Errorf("error saving source digest: %w", err)
	}

	logger.Info("Asset updated",
		slog.String("path", path),
		slog.Int("size", int(size)),
		slog.String("duration", time.Since(startTime).String()),
	)
	return nil
}

func openSourceDigest(store resource.Store, path string) (string, error) {
	file, err := store.Open(digestPath(path))
	if errors.Is(err, resource.ErrNotFound) {
		return "", nil // no digest found
	}
	if err != nil {
		return "", fmt.Errorf("error opening digest file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading digest file: %w", err)
	}
	return string(content), nil
}

func saveSourceDigest(store resource.Store, path string, digest string) error {
	file, err := store.Create(digestPath(path))
	if err != nil {
		return fmt.Errorf("error creating digest file: %w", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, digest)
	if err != nil {
		return fmt.Errorf("error writing digest file: %w", err)
	}
	return nil
}

func digestPath(path string) string {
	return fmt.Sprintf("%s.srcsha", path)
}
