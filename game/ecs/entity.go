package ecs

// Entity represents a game entity. It can be a number
// of things depending on the components that are attached
// to it.
type Entity struct {
	scene *Scene
	prev  *Entity
	next  *Entity

	components    [64]interface{}
	componentMask uint64
}

// HasComponent returns whether this Entity has a component of the
// specified type attached.
func (e *Entity) HasComponent(typeID ComponentTypeID) bool {
	return e.components[typeID] != nil
}

// Component returns the component with the specified type
// that is attached to this entity or nil if there is none.
func (e *Entity) Component(typeID ComponentTypeID) interface{} {
	return e.components[typeID]
}

// SetComponent attaches a component of the specified type
// to this entity.
func (e *Entity) SetComponent(typeID ComponentTypeID, value interface{}) {
	e.components[typeID] = value
	e.componentMask |= typeID.mask()
}

// DeleteComponent removes the component with the specified
// type from this entity.
func (e *Entity) DeleteComponent(typeID ComponentTypeID) {
	e.components[typeID] = nil
	e.componentMask &= ^typeID.mask()
}

// Delete removes this entity from the scene.
func (e *Entity) Delete() {
	for i := range e.components {
		e.components[i] = nil
	}
	e.componentMask = 0x00
	e.scene.detachEntity(e)
	e.scene.cacheEntity(e)
	e.scene = nil
}

func (e *Entity) matches(query Query) bool {
	return uint64(query)&uint64(e.componentMask) == uint64(query)
}
