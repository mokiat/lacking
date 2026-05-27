package audio

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// DecodeFunc is a function that decodes audio data from an io.Reader and
// returns a slice of audio frames.
type DecodeFunc func(io.Reader) (MediaData, error)

// Decode decodes an audio that has been encoded in a registered format.
//
// If the format of the audio data cannot be determined, [errors.ErrUnsupported]
// is returned.
func Decode(r io.Reader) (MediaData, string, error) {
	in := bufio.NewReader(r)

	decodeFn, name, err := findDecoder(in)
	if err != nil {
		return MediaData{}, "", err
	}

	data, err := decodeFn(in)
	return data, name, err
}

func findDecoder(r *bufio.Reader) (DecodeFunc, string, error) {
	registryMu.Lock()
	defer registryMu.Unlock()

	for _, f := range registeredFormats {
		count := len(f.magic)
		actualMagic, err := r.Peek(count)
		if err == nil && bytes.Equal(f.magic, actualMagic) {
			return f.decode, f.name, nil
		}
	}

	return nil, "", errors.ErrUnsupported
}
