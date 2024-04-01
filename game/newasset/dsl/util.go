package dsl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/mokiat/gomath/dprec"
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
		case float64:
			io.WriteString(h, fmt.Sprintf("%f", item))
		case dprec.Vec3:
			io.WriteString(h, item.GoString())
		case dprec.Quat:
			io.WriteString(h, item.GoString())
		case dprec.Angle:
			io.WriteString(h, fmt.Sprintf("%f", item))
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
		default:
			panic(fmt.Errorf("unsupported item type: %T", item))
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
