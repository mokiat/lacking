package audio

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"
)

// DecodeFunc is a function that decodes audio data from an io.Reader and
// returns a slice of audio frames.
type DecodeFunc func(io.Reader) (MediaData, error)

// RegisterDecoder registers an audio decoder for use by [Decode].
//
// The name parameter is a human-readable identifier for the format (e.g.
// "mp3", "wav"). The magic parameter is a magic byte prefix that
// identifies the format in raw data. The decode parameter is the function that
// will be called to decode data matching any of the magic prefixes.
func RegisterDecoder(name, magic string, decode DecodeFunc) {
	registryMu.Lock()
	defer registryMu.Unlock()

	registeredFormats = append(registeredFormats, decoderFormatEntry{
		name:   name,
		magic:  []byte(magic),
		decode: decode,
	})
}

// Decode decodes audio data encoded in a registered format.
//
// It returns the decoded [MediaData], the name of the detected format (as
// registered via [RegisterDecoder]), and any error encountered. If the format
// cannot be determined, [errors.ErrUnsupported] is returned.
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

var (
	registryMu        sync.Mutex
	registeredFormats []decoderFormatEntry
)

type decoderFormatEntry struct {
	name   string
	magic  []byte
	decode DecodeFunc
}
