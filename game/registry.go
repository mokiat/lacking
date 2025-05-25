package game

import (
	"github.com/mokiat/lacking/util/async"
)

// Registry is an interface for a resource registry that can be used to
// open and save resources.
type Registry interface {

	// ReadResource opens a resource by its path and loads it into the target
	// object. The path must follow the URL format, e.g. "level/level1".
	//
	// The operation is performed asynchronously. The returned operation should
	// not be waited upon on the main thread as it will cause a deadlock.
	ReadResource(path string, target any) async.Operation

	// WriteResource saves the given object to the specified path.
	// The path must follow the URL format, e.g. "level/level1".
	//
	// The operation is performed asynchronously. The returned operation should
	// not be waited upon on the main thread as it will cause a deadlock.
	WriteResource(path string, source any) async.Operation
}
