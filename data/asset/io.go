package asset

import (
	"compress/zlib"
	"fmt"
	"io"
)

func WriteCompressed(out io.Writer, fn func(compOut io.Writer) error) error {
	zlibOut := zlib.NewWriter(out)
	if err := fn(zlibOut); err != nil {
		return err
	}
	if err := zlibOut.Close(); err != nil {
		return fmt.Errorf("failed to complete compression: %w", err)
	}
	return nil
}

func ReadCompressed(in io.Reader, fn func(compIn io.Reader) error) error {
	zlibIn, err := zlib.NewReader(in)
	if err != nil {
		return fmt.Errorf("failed to create decompressor: %w", err)
	}
	if err := fn(zlibIn); err != nil {
		return err
	}
	if err := zlibIn.Close(); err != nil {
		return fmt.Errorf("failed to complete decompression: %w", err)
	}
	return nil
}
