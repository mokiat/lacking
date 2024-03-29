package ui

import (
	"embed"
	"io"
	"io/fs"
	"strings"

	"github.com/mokiat/lacking/util/resource"
)

//go:embed resources/*
var uiResources embed.FS

// WrapResourceLocator returns a new resource.ReadLocator that is capable of
// providing ui built-in resources as well as custom user resources.
func WrappedLocator(delegate resource.ReadLocator) resource.ReadLocator {
	return &wrappedLocator{
		delegate: delegate,
	}
}

type wrappedLocator struct {
	delegate resource.ReadLocator
}

func (l *wrappedLocator) ReadResource(uri string) (io.ReadCloser, error) {
	const matScheme = "ui:///"
	if strings.HasPrefix(uri, matScheme) {
		dir, err := fs.Sub(uiResources, "resources")
		if err != nil {
			return nil, err
		}
		return dir.Open(strings.TrimPrefix(uri, matScheme))
	}
	return l.delegate.ReadResource(uri)
}
