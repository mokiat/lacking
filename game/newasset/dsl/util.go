package dsl

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"
)

type DigestFunc func() ([]byte, error)

type digestable interface {
	Digest() ([]byte, error)
}

func digestItems(name string, items ...any) ([]byte, error) {
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, ":")
	for _, item := range items {
		switch item := item.(type) {
		case int:
			io.WriteString(h, strconv.Itoa(item))
		case string:
			io.WriteString(h, item)
		case digestable:
			digest, err := item.Digest()
			if err != nil {
				return nil, err
			}
			h.Write(digest)
		case []Operation:
			for _, operation := range item {
				digest, err := operation.Digest()
				if err != nil {
					return nil, err
				}
				h.Write(digest)
			}
		}
		io.WriteString(h, ":")
	}
	return h.Sum(nil), nil
}

func digestString[T any](provider Provider[T]) (string, error) {
	digest, err := provider.Digest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
