package mat

import (
	"embed"
	"io"
	"io/fs"
	"strings"

	"github.com/mokiat/lacking/ui"
)

//go:embed resources/*
var uiResources embed.FS

// WrapResourceLocator returns a new ui.ResourceLocator that is capable of
// providing mat resources as well as custom user resources.
func WrappedResourceLocator(delegate ui.ResourceLocator) ui.ResourceLocator {
	return &wrappedResourceLocator{
		delegate: delegate,
	}
}

type wrappedResourceLocator struct {
	delegate ui.ResourceLocator
}

func (l *wrappedResourceLocator) OpenResource(uri string) (io.ReadCloser, error) {
	const matScheme = "mat:///"
	if strings.HasPrefix(uri, matScheme) {
		dir, err := fs.Sub(uiResources, "resources")
		if err != nil {
			return nil, err
		}
		return dir.Open(strings.TrimPrefix(uri, matScheme))
	}
	return l.delegate.OpenResource(uri)
}
