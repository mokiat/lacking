package gltf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Source interface {
	Open() (io.ReadCloser, error)
	OpenRelative(uri string) (io.ReadCloser, error)
}

func NewFileSource(filename string) FileSource {
	return FileSource{
		filename: filename,
	}
}

type FileSource struct {
	filename string
}

func (s FileSource) Open() (io.ReadCloser, error) {
	file, err := os.Open(s.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", s.filename, err)
	}
	return file, nil
}

func (s FileSource) OpenRelative(relFilepath string) (io.ReadCloser, error) {
	filename := filepath.Join(filepath.Dir(s.filename), relFilepath)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	return file, nil
}
