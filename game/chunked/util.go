package chunked

import "github.com/mokiat/gblob"

type skipReader struct {
	count int
}

var _ gblob.PackedDecodable = (*skipReader)(nil)

func (r *skipReader) DecodePacked(reader gblob.TypedReader) error {
	// TODO: Add a Skip method to the reader interface, which should check
	// if underlying reader is a Seeker and use Seek to skip the data, otherwise
	// it should use CopyN into discard.
	temp := make([]byte, 1024)
	for r.count > 0 {
		skipCount := min(r.count, len(temp))
		if err := reader.ReadBytes(temp[:skipCount]); err != nil {
			return err
		}
		r.count -= skipCount
	}
	return nil
}

type countedWriter struct {
	count uint32
}

func (w *countedWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.count += uint32(n)
	return n, nil
}
