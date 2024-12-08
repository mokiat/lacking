package asset

// Blob represents a binary large object that is used to store arbitrary data.
type Blob struct {

	// Name is the name of the blob.
	Name string

	// Data is the binary data of the blob.
	Data []byte
}
