package ioutil

import "io"

func NopWriteCloser(delegate io.Writer) io.WriteCloser {
	return &nopWriteCloser{
		Writer: delegate,
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error {
	return nil
}
