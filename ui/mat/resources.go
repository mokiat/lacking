package mat

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
// providing mat resources as well as custom user resources.
func WrappedResourceLocator(delegate resource.ReadLocator) resource.ReadLocator {
	return &wrappedResourceLocator{
		delegate: delegate,
	}
}

type wrappedResourceLocator struct {
	delegate resource.ReadLocator
}

func (l *wrappedResourceLocator) ReadResource(uri string) (io.ReadCloser, error) {
	const matScheme = "mat:///"
	if strings.HasPrefix(uri, matScheme) {
		dir, err := fs.Sub(uiResources, "resources")
		if err != nil {
			return nil, err
		}
		return dir.Open(strings.TrimPrefix(uri, matScheme))
	}
	return l.delegate.ReadResource(uri)
}
