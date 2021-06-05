package ecs

// Having creates a new Query that matches entities
// with the specified component type.
func Having(typeID ComponentTypeID) Query {
	var q Query
	return q.And(typeID)
}

// Query represents a search expression.
// Currently, only a conjunction of component types
// is possible.
type Query uint64

// And adds an additional component type as a requirement
// for an entity to match this query.
func (q Query) And(typeID ComponentTypeID) Query {
	return Query(uint64(q) | typeID.mask())
}

func newResult(s *Scene) *Result {
	return &Result{
		scene:    s,
		entities: make([]*Entity, 0, 1024),
		offset:   -1,
	}
}

// Result represents the outcome of a search
// operation.
// Make sure to call Close on a Result once
// you are done with it.
type Result struct {
	scene *Scene

	entities []*Entity
	offset   int
}

// HasNext returns whether there are any more
// entities in this result set.
func (r *Result) HasNext() bool {
	return r.offset+1 < len(r.entities)
}

// Next returns the next available entity
// in this result set. Make sure to first use
// HasNext before calling this method.
func (r *Result) Next() *Entity {
	r.offset++
	return r.entities[r.offset]
}

// Close releases the resources allocated for
// this result.
func (r *Result) Close() {
	r.entities = r.entities[:0]
	r.offset = -1
	r.scene.cacheResult(r)
}
