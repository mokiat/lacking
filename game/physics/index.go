package physics

func newIndexReference(index, revision uint32) indexReference {
	return indexReference(uint64(revision)<<32 | uint64(index))
}

type indexReference uint64

func (r indexReference) IsValid() bool {
	return r > 0
}

func (r indexReference) Index() uint32 {
	return uint32(r & 0xFFFFFFFF)
}

func (r indexReference) Revision() uint32 {
	return uint32(r >> 32)
}
