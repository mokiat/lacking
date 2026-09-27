package animation3

// Skeleton assigns a stable index to each bone name of an animated model, so
// that per-frame work can use bone indices instead of names.
//
// A Skeleton is immutable once created and can be shared by any number of
// consumers.
type Skeleton struct {
	names   []string
	indices map[string]int
}

// NewSkeleton creates a new [Skeleton] from the specified bone names. Bones
// are assigned consecutive indices starting from 0, in the order of their
// first occurrence in the list; later duplicates of a name are ignored and do
// not consume an index. The skeleton does not retain the specified slice.
func NewSkeleton(boneNames []string) *Skeleton {
	names := make([]string, 0, len(boneNames))
	indices := make(map[string]int, len(boneNames))
	for _, name := range boneNames {
		if _, ok := indices[name]; ok {
			continue
		}
		indices[name] = len(names)
		names = append(names, name)
	}
	return &Skeleton{
		names:   names,
		indices: indices,
	}
}

// BoneCount returns the number of bones in the skeleton.
func (s *Skeleton) BoneCount() int {
	return len(s.names)
}

// BoneName returns the name of the bone at the specified index. An index
// outside the valid range returns an empty string.
func (s *Skeleton) BoneName(index int) string {
	if index < 0 || index >= len(s.names) {
		return ""
	}
	return s.names[index]
}

// BoneIndex returns the index of the bone with the specified name and
// whether such a bone exists. An unknown name returns an index of 0 and false.
func (s *Skeleton) BoneIndex(name string) (int, bool) {
	index, ok := s.indices[name]
	return index, ok
}
