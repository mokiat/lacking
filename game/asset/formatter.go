package asset

import (
	"compress/zlib"
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
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
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
	zlibOut := zlib.NewWriter(out)
	if err := gblob.NewLittleEndianPackedEncoder(zlibOut).Encode(value); err != nil {
		return err
	}
	return zlibOut.Close()
}

func (*blobFormatter) Decode(in io.Reader, target any) error {
	zlibIn, err := zlib.NewReader(in)
	if err != nil {
		return err
	}
	if err := gblob.NewLittleEndianPackedDecoder(zlibIn).Decode(target); err != nil {
		return err
	}
	return zlibIn.Close()
}
