package pack

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strconv"
	"time"
)

type Hashable interface {
	Digest(hasher hash.Hash) error
}

type HashableParams map[string]interface{}

func WriteCompositeDigest(hasher hash.Hash, name string, params HashableParams) error {
	io.WriteString(hasher, name)
	io.WriteString(hasher, "(")
	for paramName, paramValue := range params {
		io.WriteString(hasher, paramName)
		io.WriteString(hasher, "=")
		switch paramValue := paramValue.(type) {
		case string:
			io.WriteString(hasher, fmt.Sprintf("%q", paramValue))
		case int:
			io.WriteString(hasher, strconv.Itoa(paramValue))
		case Hashable:
			if err := paramValue.Digest(hasher); err != nil {
				return err
			}
		case time.Duration:
			io.WriteString(hasher, paramValue.String())
		default:
			return fmt.Errorf("could not write digest for %T", paramValue)
		}
		io.WriteString(hasher, ";")
	}
	io.WriteString(hasher, ")")
	return nil
}

func CalculateDigest(hashable Hashable) ([]byte, error) {
	hasher := sha256.New()
	if err := hashable.Digest(hasher); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func EqualDigests(a, b []byte) bool {
	return bytes.Equal(a, b)
}
