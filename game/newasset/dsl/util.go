package dsl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
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
		case bool:
			io.WriteString(h, strconv.FormatBool(item))
		case int:
			io.WriteString(h, strconv.Itoa(item))
		case uint8:
			io.WriteString(h, strconv.Itoa(int(item)))
		case string:
			io.WriteString(h, item)
		case float32:
			io.WriteString(h, fmt.Sprintf("%f", item))
		case time.Time:
			io.WriteString(h, item.Format(time.RFC3339))
		case sprec.Vec2:
			io.WriteString(h, item.GoString())
		case sprec.Vec3:
			io.WriteString(h, item.GoString())
		case sprec.Vec4:
			io.WriteString(h, item.GoString())
		case float64:
			io.WriteString(h, fmt.Sprintf("%f", item))
		case dprec.Vec3:
			io.WriteString(h, item.GoString())
		case dprec.Vec4:
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
