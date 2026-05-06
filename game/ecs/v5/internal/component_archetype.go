package internal

// NewArchetype creates a new Archetype.
func NewArchetype() *Archetype {
	return &Archetype{}
}

// Archetype represents a unique combination of component types. It is used to
// group entities that have the same set of component types together for
// efficient storage and querying.
type Archetype struct {
	mask       TypeMask
	size       uint32
	lookup     TypeLookup
	components []BaseColumn
}

func (a *Archetype) Revive() {
	a.mask = EmptyTypeMask()
	a.size = 0
}

func (a *Archetype) Destroy() {
	// TODO: Release columns first?
	a.mask = EmptyTypeMask()
	a.size = 0
	a.lookup = TypeLookup{}
	clear(a.components)
	a.components = a.components[:0]
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
	offset := uint32(row) % chunkSize
	chunkIndex := uint32(row) / chunkSize

	var result TypePlacementMap
	a.mask.EachType(func(id TypeID) {
		column := a.components[a.lookup[id]]
		result[id].ChunkIndex = uint32(row) / chunkSize
	})
	for i := range result {
		result[i].Offset = offset
	}
	return result
}

// ArchetypeRow represents a single row in an archetype, which corresponds to a
// single entity's worth of component data.
type ArchetypeRow uint32
