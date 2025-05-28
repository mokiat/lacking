package dsl

import (
	"fmt"
	"maps"
	"sync"

	"github.com/mokiat/lacking/storage/chunked"
)

// Resource represents a generic asset that can be converted by a Converter
// to a chunked format.
type Resource any

// Converter is an interface for converting a Resource into a chunked format.
type Converter interface {

	// CanConvert checks if the converter can handle the given resource.
	CanConvert(asset Resource) bool

	// Convert converts the given resource into a chunked format.
	Convert(asset Resource) (chunked.Chunk, error)
}

var (
	registeredConverters   = make(map[string]Converter)
	registeredConvertersMU sync.RWMutex
)

// RegisterConverter registers a new converter with the given name.
//
// This function is safe to call concurrently, though in reality it should
// mostly be used during application initialization (i.e. inside an init).
func RegisterConverter(name string, converter Converter) {
	registeredConvertersMU.Lock()
	defer registeredConvertersMU.Unlock()

	if _, exists := registeredConverters[name]; exists {
		panic(fmt.Errorf("converter already registered: %s", name))
	}
	registeredConverters[name] = converter
}

// Converters returns a snapshot of the currently registered converters.
//
// This function is safe to call concurrently.
func Converters() map[string]Converter {
	registeredConvertersMU.RLock()
	defer registeredConvertersMU.RUnlock()

	return maps.Clone(registeredConverters)
}
