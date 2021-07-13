package optional

// NewBool returns a specified Bool value.
func NewBool(value bool) Bool {
	return Bool{
		Specified: true,
		Value:     value,
	}
}

// Bool represents an optional bool value.
type Bool struct {
	Specified bool
	Value     bool
}
