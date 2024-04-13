package dsl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
)

// DigestFunc is a function that calculates a digest.
type DigestFunc func() ([]byte, error)

// Digestable represents an object that can be digested.
type Digestable interface {

	// Digest calculates the digest of the object.
	Digest() ([]byte, error)
}

// StringDigest calculates the digest of the provided digestable
// and returns it as a hex-encoded string.
func StringDigest(digestable Digestable) (string, error) {
	digest, err := digestable.Digest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}

// CreateDigest calculates the digest of the provided operation and
// supplementing parameters.
func CreateDigest(name string, params ...any) ([]byte, error) {
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, ":")
	for _, param := range params {
		if opts, ok := param.([]Operation); ok {
			for _, opt := range opts {
				if err := digestParam(h, opt); err != nil {
					return nil, err
				}
				io.WriteString(h, ":")
			}
		} else {
			if err := digestParam(h, param); err != nil {
				return nil, err
			}
			io.WriteString(h, ":")
		}
	}
	return h.Sum(nil), nil
}

func digestParam(out io.Writer, param any) error {
	encoder := hex.NewEncoder(out)
	return writeValue(encoder, param)
}

func writeValue(out io.Writer, value any) error {
	io.WriteString(out, reflect.TypeOf(value).String())
	io.WriteString(out, ":")
	switch value := value.(type) {
	case bool:
		io.WriteString(out, strconv.FormatBool(value))
	case uint8:
		io.WriteString(out, strconv.Itoa(int(value)))
	case uint16:
		io.WriteString(out, strconv.Itoa(int(value)))
	case uint32:
		io.WriteString(out, strconv.Itoa(int(value)))
	case uint64:
		io.WriteString(out, strconv.Itoa(int(value)))
	case int8:
		io.WriteString(out, strconv.Itoa(int(value)))
	case int16:
		io.WriteString(out, strconv.Itoa(int(value)))
	case int32:
		io.WriteString(out, strconv.Itoa(int(value)))
	case int64:
		io.WriteString(out, strconv.Itoa(int(value)))
	case uint:
		io.WriteString(out, strconv.Itoa(int(value)))
	case int:
		io.WriteString(out, strconv.Itoa(value))
	case string:
		io.WriteString(out, value)
	case float32:
		io.WriteString(out, strconv.Itoa(int(math.Float32bits(value))))
	case float64:
		io.WriteString(out, strconv.Itoa(int(math.Float64bits(value))))
	case time.Time:
		io.WriteString(out, value.Format(time.RFC3339))
	case sprec.Vec2:
		io.WriteString(out, value.GoString())
	case sprec.Vec3:
		io.WriteString(out, value.GoString())
	case sprec.Vec4:
		io.WriteString(out, value.GoString())
	case sprec.Quat:
		io.WriteString(out, value.GoString())
	case sprec.Angle:
		io.WriteString(out, fmt.Sprintf("%f", value))
	case dprec.Vec2:
		io.WriteString(out, value.GoString())
	case dprec.Vec3:
		io.WriteString(out, value.GoString())
	case dprec.Vec4:
		io.WriteString(out, value.GoString())
	case dprec.Quat:
		io.WriteString(out, value.GoString())
	case dprec.Angle:
		io.WriteString(out, fmt.Sprintf("%f", value))
	case Digestable:
		digest, err := StringDigest(value)
		if err != nil {
			return err
		}
		io.WriteString(out, digest)
	default:
		panic(fmt.Errorf("unsupported value type: %T", value))
	}
	return nil
}
