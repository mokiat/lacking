package component

// IsEqualData compares if the two data objects are equal
func IsEqualData(old, new any) bool {
	// IDEA: Allow for custom equals logic through the annotation of
	// the struct fields of the data (when it is a struct).
	return old == new
}
