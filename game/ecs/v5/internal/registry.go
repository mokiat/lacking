package internal

// NewRegistry creates and initializes a new registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Registry represents a registry of component storages. It provides methods for
// registering and retrieving component storages based on component type
// identifiers.
type Registry struct {
	storages [MaxComponentTypes]BaseStorage
}

// Storage returns the component storage associated with the specified component
// type.
func (r *Registry) Storage(id TypeID) BaseStorage {
	return r.storages[id]
}

// Storage returns the component storage associated with the specified component
// type.
func (r *Registry) SetStorage(id TypeID, storage BaseStorage) {
	r.storages[id] = storage
}
