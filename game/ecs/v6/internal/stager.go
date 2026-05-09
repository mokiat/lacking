package internal

func NewStager(registry *Registry) *Stager {
	var (
		lookup  TypeLookup
		columns []AnyColumn
	)

	for typeID, storage := range registry.storages {
		if storage == nil {
			continue
		}
		lookup[typeID] = uint8(len(columns))
		column := storage.NewAnyColumn()
		columns = append(columns, column)
		column.Grow()
	}

	return &Stager{
		registry:         registry,
		componentColumns: columns,
		componentLookup:  lookup,
		capacity:         1,
		size:             0,
	}
}

type Stager struct {
	registry         *Registry
	componentColumns []AnyColumn
	componentLookup  TypeLookup
	capacity         uint32
	size             uint32
}

func (s *Stager) Clear() {
	s.size = 0
}

func (s *Stager) Grow() Row {
	s.size++
	if s.size > s.capacity {
		s.capacity++
		for _, column := range s.componentColumns {
			column.Grow()
		}
	}
	return Row(s.size - 1)
}

func (s *Stager) ComponentColumn(id TypeID) AnyColumn {
	return s.componentColumns[s.componentLookup[id]]
}

func (s *Stager) Destroy() {
	for i := range s.componentColumns {
		s.componentColumns[i].Release()
		s.componentColumns[i] = nil
	}
	s.componentColumns = nil
}
