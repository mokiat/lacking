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

	idColumn         *Column[ID]
	componentColumns []AnyColumn
	componentLookup  TypeLookup
}

// Revive initializes the archetype with the specified type mask. It sets up the
// necessary columns for the component types included in the mask.
func (a *Archetype) Revive(mask TypeMask) {
	a.mask = mask
	a.size = 0

	mask.EachType(func(id TypeID) {
		storage := a.registry.Storage(id)
		a.componentLookup[id] = uint8(len(a.componentColumns))
		a.componentColumns = append(a.componentColumns, storage.NewAnyColumn())
	})

	entityIDStorage := a.registry.IDStorage()
	a.idColumn = entityIDStorage.NewColumn()
}

// Destroy cleans up the archetype and releases any resources it holds.
// It should be called when the archetype is no longer needed, such as when it
// is being returned to the archetype pool.
func (a *Archetype) Destroy() {
	a.mask = EmptyTypeMask()
	a.size = 0

	a.componentLookup = TypeLookup{}
	for i := range a.componentColumns {
		a.componentColumns[i].Release()
		a.componentColumns[i] = nil
	}
	a.componentColumns = a.componentColumns[:0]

	a.idColumn.Release()
	a.idColumn = nil
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

// ComponentColumn returns the column associated with the specified component
// type ID.
func (a *Archetype) ComponentColumn(id TypeID) AnyColumn {
	return a.componentColumns[a.componentLookup[id]]
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
	for _, column := range a.componentColumns {
		column.Grow()
	}
	a.idColumn.Grow()
	return Row(a.size - 1)
}

// Shrink removes the last row from the table represented by the archetype.
func (a *Archetype) Shrink() {
	a.size--
	for _, column := range a.componentColumns {
		column.Shrink()
	}
	a.idColumn.Shrink()
}

func (a *Archetype) CopyRow(dst, src Row) {
	for _, column := range a.componentColumns {
		column.Copy(dst, src)
	}
	a.idColumn.Copy(dst, src)
}
