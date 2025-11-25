package game

import "github.com/mokiat/lacking/util/async"

// AssetLoader represents an async loading process.
type AssetLoader struct {
	engine      *Engine
	resourceSet *ResourceSet
}

// AsyncEngine returns the async engine associated with this asset loader.
func (l *AssetLoader) Engine() *Engine {
	return l.engine
}

// ResourceSet returns the resource set associated with this asset loader.
func (l *AssetLoader) ResourceSet() *ResourceSet {
	return l.resourceSet
}

// ScheduleIO schedules an operation to be executed on the IO worker.
func (l *AssetLoader) ScheduleIO(cb func() error) async.Operation {
	return l.engine.ScheduleIO(cb)
}

// ScheduleMain schedules an operation to be executed on the main thread.
func (l *AssetLoader) ScheduleMain(cb func() error) async.Operation {
	return l.engine.ScheduleMain(cb)
}
