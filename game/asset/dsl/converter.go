package dsl

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/storage/chunked"
)

var registeredConverters []Converter

// Use registers a converter for use in the DSL.
func Use(converters ...Converter) any {
	registeredConverters = append(registeredConverters, converters...)
	return nil
}

// Converter is an interface for converting a resource into a chunked format.
type Converter interface {

	// Convert converts the given resource into a number of chunks
	// and appends them to the target list.
	Convert(target *ds.List[chunked.Chunk], resource any) error
}

// ConverterFunc is a function type that implements the Converter interface.
type ConverterFunc func(target *ds.List[chunked.Chunk], resource any) error

// Convert converts the given resource into a number of chunks
// and appends them to the target list.
func (f ConverterFunc) Convert(target *ds.List[chunked.Chunk], resource any) error {
	return f(target, resource)
}
