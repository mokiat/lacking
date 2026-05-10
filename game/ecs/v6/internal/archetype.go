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

	idColumn           *Column[ID]
	componentColumnIDs []ColumnID
	componentLookup    TypeLookup
}

// Revive initializes the archetype with the specified type mask. It sets up the
// necessary columns for the component types included in the mask.
func (a *Archetype) Revive(mask TypeMask) {
	a.mask = mask
	a.size = 0

	mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		a.componentLookup[id] = uint8(len(a.componentColumnIDs))
		a.componentColumnIDs = append(a.componentColumnIDs, storage.AllocateColumn())
	})

	entityIDStorage := a.registry.IDStorage()
	a.idColumn = entityIDStorage.NewColumn()
}

// Destroy cleans up the archetype and releases any resources it holds.
// It should be called when the archetype is no longer needed, such as when it
// is being returned to the archetype pool.
func (a *Archetype) Destroy() {
	a.mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		columnID := a.componentColumnIDs[a.componentLookup[id]]
		storage.ReleaseColumn(columnID)
	})
	a.componentLookup = TypeLookup{}
	a.componentColumnIDs = a.componentColumnIDs[:0]

	a.idColumn.Release()
	a.idColumn = nil

	a.mask = EmptyTypeMask()
	a.size = 0
}

// TypeMask returns the type mask associated with the archetype.
func (a *Archetype) TypeMask() TypeMask {
	return a.mask
}

// Size returns the number of entities currently stored in the archetype.
func (a *Archetype) Size() uint32 {
	return a.size
}

// IsEmpty returns whether the archetype has no entities.
func (a *Archetype) IsEmpty() bool {
	return a.size == 0
}

// IDColumn returns the column that stores the entity IDs for the archetype.
func (a *Archetype) IDColumn() *Column[ID] {
	return a.idColumn
}

// ComponentColumnID returns the ID of the column associated with the specified
// component type ID.
func (a *Archetype) ComponentColumnID(id TypeID) ColumnID {
	return a.componentColumnIDs[a.componentLookup[id]]
}

// ComponentColumnIDs returns the columns associated with the component types in
// the archetype, along with a lookup that maps component type IDs to their
// corresponding column indices.
func (a *Archetype) ComponentColumnIDs() ([]ColumnID, TypeLookup) {
	return a.componentColumnIDs, a.componentLookup
}

// LastRow returns the index of the last row in the table represented by the
// archetype.
//
// Calling this for an empty archetype will return an invalid row index.
func (a *Archetype) LastRow() Row {
	return Row(a.size - 1)
}

// Grow appends a new row to the table represented by the archetype.
func (a *Archetype) Grow() Row {
	a.size++
	a.mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		columnID := a.componentColumnIDs[a.componentLookup[id]]
		storage.GrowColumn(columnID)
	})
	a.idColumn.Grow()
	return Row(a.size - 1)
}

// Shrink removes the last row from the table represented by the archetype.
func (a *Archetype) Shrink() {
	a.size--
	a.mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		columnID := a.componentColumnIDs[a.componentLookup[id]]
		storage.ShrinkColumn(columnID)
	})
	a.idColumn.Shrink()
}

// CopyRow copies the component values from the source row to the destination
// row in the table represented by the archetype.
func (a *Archetype) CopyRow(dst, src Row) {
	a.mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		columnID := a.componentColumnIDs[a.componentLookup[id]]
		storage.CopyCell(columnID, dst, columnID, src)
	})
	a.idColumn.Copy(dst, src)
}
