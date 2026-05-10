package internal

func NewStager(registry *Registry) *Stager {
	var (
		columnIDs [MaxComponentTypes]ColumnID
		mask      TypeMask
	)

	for typeID, storage := range registry.storages {
		if storage == nil {
			continue
		}
		column := storage.NewAnyColumn()
		column.Grow()
		columnIDs[typeID] = column.ID()
		mask.AddType(TypeID(typeID))
	}

	return &Stager{
		registry:           registry,
		componentColumnIDs: columnIDs,
		mask:               mask,
		capacity:           1,
		size:               0,
	}
}

type Stager struct {
	registry           *Registry
	componentColumnIDs [MaxComponentTypes]ColumnID
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
			columnID := s.componentColumnIDs[id]
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
		columnID := s.componentColumnIDs[id]
		s.registry.Storage(id).ReclaimColumn(columnID)
	})
}
