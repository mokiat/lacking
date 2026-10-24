package animation3

// SkeletonBone describes a single bone that is used to construct a
// [Skeleton].
type SkeletonBone struct {

	// Name is the name of the bone, which identifies it within the skeleton.
	Name string

	// RestTransform is the local transformation of the bone, relative to its
	// parent, when the bone is not animated.
	RestTransform Transform
}

// Skeleton assigns a stable index to each bone name of an animated model, so
// that per-frame work can use bone indices instead of names.
//
// A Skeleton is immutable once created and can be shared by any number of
// consumers.
type Skeleton struct {
	indices        map[string]int
	names          []string
	restTransforms []Transform
}

// NewSkeleton creates a new [Skeleton] from the specified bones. Bones are
// assigned consecutive indices starting from 0, in the order of their first
// occurrence in the list; later bones with a duplicate name are ignored, along
// with their rest transforms, and do not consume an index. The skeleton does
// not retain the specified slice.
func NewSkeleton(bones []SkeletonBone) *Skeleton {
	indices := make(map[string]int, len(bones))
	names := make([]string, 0, len(bones))
	restTransforms := make([]Transform, 0, len(bones))
	for _, bone := range bones {
		if _, ok := indices[bone.Name]; ok {
			continue
		}
		indices[bone.Name] = len(names)
		names = append(names, bone.Name)
		restTransforms = append(restTransforms, bone.RestTransform)
	}
	return &Skeleton{
		indices:        indices,
		names:          names,
		restTransforms: restTransforms,
	}
}

// BoneCount returns the number of bones in the skeleton.
func (s *Skeleton) BoneCount() int {
	return len(s.names)
}

// BoneIndex returns the index of the bone with the specified name and
// whether such a bone exists. An unknown name returns an index of 0 and false.
func (s *Skeleton) BoneIndex(name string) (int, bool) {
	index, ok := s.indices[name]
	return index, ok
}

// BoneName returns the name of the bone at the specified index. The index
// must be in the range [0..BoneCount).
func (s *Skeleton) BoneName(index int) string {
	return s.names[index]
}

// BoneRestTransform returns the rest transformation of the bone at the
// specified index. The index must be in the range [0..BoneCount).
func (s *Skeleton) BoneRestTransform(bone int) Transform {
	return s.restTransforms[bone]
}
