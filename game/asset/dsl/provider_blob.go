package dsl

import (
	"fmt"
	"os"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// OpenBlob opens an binary file from the provided path.
func OpenBlob(name, path string) Provider[*mdl.Blob] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Blob, error) {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %q: %w", path, err)
			}

			return &mdl.Blob{
				Name: name,
				Data: data,
			}, nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return CreateDigest("open-blob", name, path, info.ModTime())
		},
	))
}
