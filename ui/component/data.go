package component

import "reflect"

// IsEqualData compares if the two data objects are equal
func IsEqualData(old, new interface{}) bool {
	oldType := reflect.TypeOf(old)
	newType := reflect.TypeOf(new)
	if oldType != newType {
		return false
	}
	// TODO: Do per-field analysis of tags so that deep comparisons
	// can be made and also special types can be handled.
	// if newType.Kind() != reflect.Struct {
	// 	return false
	// }
	return old == new
}
