package chunked

import (
	"path/filepath"

	"github.com/mokiat/gblob"
)

func cleanFilePath(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}

type skipReader struct {
	count int
}

var _ gblob.PackedDecodable = (*skipReader)(nil)

func (r *skipReader) DecodePacked(reader gblob.TypedReader) error {
	return reader.SkipBytes(r.count)
}

type countedWriter struct {
	count uint32
}

func (w *countedWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.count += uint32(n)
	return n, nil
}
