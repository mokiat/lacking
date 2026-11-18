package dsl

import (
	"fmt"
	"os"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// OpenSound returns a provider that loads and decodes the audio file
// at the specified path into an [mdl.Sound].
//
// The digest depends on the path and the file's modification time.
func OpenSound(path string) Provider[*mdl.Sound] {
	// get function
	return FuncProvider(
		func() (*mdl.Sound, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %q: %w", path, err)
			}
			defer file.Close()

			sound, err := mdl.ParseSound(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse sound %q: %w", path, err)
			}
			return sound, nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return CreateDigest("open-sound", path, info.ModTime())
		},
	)
}
