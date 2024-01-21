package physics

import "fmt"

func newIndexReference(index, revision uint32) indexReference {
	return indexReference(uint64(revision)<<32 | uint64(index))
}

type indexReference uint64

func (r indexReference) IsValid() bool {
	return r.Revision() > 0
}

func (r indexReference) Index() uint32 {
	return uint32(r & 0xFFFFFFFF)
}

func (r indexReference) Revision() uint32 {
	return uint32(r >> 32)
}

func (r indexReference) String() string {
	return fmt.Sprintf("%d:%d", r.Index(), r.Revision())
}