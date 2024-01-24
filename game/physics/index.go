package physics

import "fmt"

func newIndexReference(index, revision uint32) indexReference {
	return indexReference{
		Index:    index,
		Revision: revision,
	}
}

type indexReference struct {
	Index    uint32
	Revision uint32
}

func (r indexReference) IsValid() bool {
	return r.Revision > 0
}

func (r indexReference) String() string {
	return fmt.Sprintf("%d:%d", r.Index, r.Revision)
}
