package audio

import (
	"sync"
)

// RegisterFormat registers an audio format for use by [Decode].
//
// The name parameter is a human-readable identifier for the format (e.g.
// "mp3", "wav"). The magics parameter is a list of magic byte prefixes that
// identify the format in raw data. The decode parameter is the function that
// will be called to decode data matching any of the magic prefixes.
func RegisterFormat(name string, magics []string, decode DecodeFunc) {
	registryMu.Lock()
	defer registryMu.Unlock()

	for _, magic := range magics {
		registeredFormats = append(registeredFormats, formatEntry{
			name:   name,
			magic:  []byte(magic),
			decode: decode,
		})
	}
}

var (
	registryMu        sync.Mutex
	registeredFormats []formatEntry
)

type formatEntry struct {
	name   string
	magic  []byte
	decode DecodeFunc
}
