package asset

import (
	"encoding/json"
	"io"

	"github.com/mokiat/gblob"
)

// Formatter represents an interface for formatting assets into storage format.
type Formatter interface {

	// Encode writes the specified value to the specified writer.
	Encode(out io.Writer, value any) error

	// Decode reads the specified value from the specified reader.
	Decode(in io.Reader, target any) error
}

// NewJSONFormatter creates a new Formatter that formats assets in JSON format.
func NewJSONFormatter() Formatter {
	return &jsonFormatter{}
}

type jsonFormatter struct{}

func (*jsonFormatter) Encode(out io.Writer, value any) error {
	return json.NewEncoder(out).Encode(value)
}

func (*jsonFormatter) Decode(in io.Reader, target any) error {
	return json.NewDecoder(in).Decode(target)
}

// NewBlobFormatter creates a new Formatter that formats assets in binary
// format.
func NewBlobFormatter() Formatter {
	return &blobFormatter{}
}

type blobFormatter struct{}

func (*blobFormatter) Encode(out io.Writer, value any) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(value)
}

func (*blobFormatter) Decode(in io.Reader, target any) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(target)
}
