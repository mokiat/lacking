package internal

// NewRegistry creates and initializes a new registry.
func NewRegistry() *Registry {
	return &Registry{
		idStorage: NewStorage[uint32](),
	}
}

// Registry represents a registry of component storages. It provides methods for
// registering and retrieving component storages based on component type
// identifiers.
type Registry struct {
	idStorage *Storage[uint32]
	storages  [MaxComponentTypes]AnyStorage
}

// IDStorage returns the storage used for storing ID values.
func (r *Registry) IDStorage() *Storage[uint32] {
	return r.idStorage
}

// Storage returns the component storage associated with the specified component
// type.
func (r *Registry) Storage(id TypeID) AnyStorage {
	return r.storages[id]
}

// Storage returns the component storage associated with the specified component
// type.
func (r *Registry) SetStorage(id TypeID, storage AnyStorage) {
	r.storages[id] = storage
}
