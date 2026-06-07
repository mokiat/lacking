package dsl

import (
	"fmt"
	"io"
	"os"
)

// StreamFile streams a binary file from the provided path.
func StreamFile(path string) Provider[io.ReadCloser] {
	return OnceProvider(FuncProvider(
		// get function
		func() (io.ReadCloser, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open raw file %q: %w", path, err)
			}
			return file, nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return CreateDigest("stream-file", path, info.ModTime())
		},
	))
}
