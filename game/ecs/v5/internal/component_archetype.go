package internal

// NewArchetype creates a new Archetype.
func NewArchetype(registry *Registry) *Archetype {
	return &Archetype{
		registry: registry,
	}
}

// Archetype represents a unique combination of component types. It is used to
// group entities that have the same set of component types together for
// efficient storage and querying.
type Archetype struct {
	registry *Registry
	mask     TypeMask
	size     uint32
	lookup   TypeLookup
	columns  []BaseColumn
}

func (a *Archetype) Revive(mask TypeMask) {
	a.mask = mask
	a.size = 0

	mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		a.lookup[id] = uint8(len(a.columns))
		a.columns = append(a.columns, storage.CreateColumn())
	})
}

func (a *Archetype) Destroy() {
	a.mask = EmptyTypeMask()
	a.size = 0
	a.lookup = TypeLookup{}
	for i := range a.columns {
		a.columns[i].Destroy()
		a.columns[i] = nil
	}
	a.columns = a.columns[:0]
}

// TypeMask returns the type mask associated with the archetype.
func (a *Archetype) TypeMask() TypeMask {
	return a.mask
}

// IsEmpty returns whether the archetype has no entities.
func (a *Archetype) IsEmpty() bool {
	return a.size == 0
}

func (a *Archetype) AllocateRow() ArchetypeRow {
	row := a.size
	a.size++
	// TODO: Allocate chunks if needed.
	return ArchetypeRow(row)
}

func (a *Archetype) ReleaseRow(row ArchetypeRow) {
	// TODO: Check if archetype is frozen. If it is, then just mark
	// the row for deletion in a stack. Otherwise, just perform a swap with
	// the last row and decrease the size.
}

// PlacementMap returns a mapping from component type identifiers to storage
// positions for the archetype and row.
func (a *Archetype) PlacementMap(row ArchetypeRow) TypePlacementMap {
	var result TypePlacementMap
	a.mask.EachType(func(id TypeID) {
		column := a.columns[a.lookup[id]]
		result[id] = column.StoragePosition(row)
	})
	return result
}

// ArchetypeRow represents a single row in an archetype, which corresponds to a
// single entity's worth of component data.
type ArchetypeRow uint32
