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

// TypeMask returns the type mask associated with the archetype.
func (a *Archetype) TypeMask() TypeMask {
	return a.mask
}

// ArchetypeRow represents a single row in an archetype, which corresponds to a
// single entity's worth of component data.
type ArchetypeRow uint32
