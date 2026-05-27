package audio

import (
	"sync"
)

// RegisterFormat registers an audio format for use by [Decode].
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
