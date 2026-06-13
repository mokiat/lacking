package chunked

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking/core/resource"
)

func NewAsset(store resource.Store, path string) *Asset {
	path = cleanFilePath(path)
	return &Asset{
		store: store,
		path:  path,
	}
}

type Asset struct {
	store resource.Store
	path  string
}

func (a *Asset) Path() string {
	return a.path
}

func (a *Asset) Read(target any) error {
	in, err := a.store.Open(a.path)
	if err != nil {
		return fmt.Errorf("error opening asset file: %w", err)
	}
	defer in.Close()

	dec := decoder{
		in: gblob.NewLittleEndianPackedDecoder(in),
	}
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("error decoding asset: %w", err)
	}

	return nil
}

func (a *Asset) Write(source any) error {
	out, err := a.store.Create(a.path)
	if err != nil {
		return fmt.Errorf("error creating asset file: %w", err)
	}
	defer out.Close()

	enc := encoder{
		out: gblob.NewLittleEndianPackedEncoder(out),
	}
	if err := enc.Encode(source); err != nil {
		return fmt.Errorf("error encoding asset: %w", err)
	}

	return nil
}
