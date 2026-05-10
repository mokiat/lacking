package internal

func NewStager(registry *Registry) *Stager {
	var (
		componentColumnIDs []ColumnID
		componentLookup    TypeLookup
		mask               TypeMask
	)

	for typeID, storage := range registry.storages {
		if storage == nil {
			continue
		}

		column := storage.NewAnyColumn()
		column.Grow()

		componentLookup[typeID] = uint8(len(componentColumnIDs))
		componentColumnIDs = append(componentColumnIDs, column.ID())
		mask.AddType(TypeID(typeID))
	}

	return &Stager{
		registry:           registry,
		componentColumnIDs: componentColumnIDs,
		componentLookup:    componentLookup,
		mask:               mask,
		capacity:           1,
		size:               0,
	}
}

type Stager struct {
	registry           *Registry
	componentColumnIDs []ColumnID
	componentLookup    TypeLookup
	mask               TypeMask
	capacity           uint32
	size               uint32
}

func (s *Stager) Clear() {
	s.size = 0
}

func (s *Stager) Grow() Row {
	s.size++
	if s.size > s.capacity {
		s.capacity++
		s.mask.EachType(func(id TypeID) {
			columnID := s.componentColumnIDs[s.componentLookup[id]]
			s.registry.Storage(id).GrowColumn(columnID)
		})
	}
	return Row(s.size - 1)
}

func (s *Stager) ComponentColumnID(id TypeID) ColumnID {
	return s.componentColumnIDs[id]
}

func (s *Stager) Destroy() {
	s.mask.EachType(func(id TypeID) {
		columnID := s.componentColumnIDs[s.componentLookup[id]]
		s.registry.Storage(id).ReclaimColumn(columnID)
	})
}
