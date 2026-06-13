package resource

import "path/filepath"

func cleanFilePath(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}
