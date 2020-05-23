package asset

import (
	"compress/zlib"
	"encoding/gob"
	"fmt"
	"io"
)

func Encode(out io.Writer, asset interface{}) error {
	return writeCompressed(out, func(compOut io.Writer) error {
		if err := gob.NewEncoder(compOut).Encode(asset); err != nil {
			return fmt.Errorf("failed to encode gob stream: %w", err)
		}
		return nil
	})
}

func Decode(in io.Reader, asset interface{}) error {
	return readCompressed(in, func(compIn io.Reader) error {
		if err := gob.NewDecoder(compIn).Decode(asset); err != nil {
			return fmt.Errorf("failed to decode gob stream: %w", err)
		}
		return nil
	})
}

func writeCompressed(out io.Writer, fn func(compOut io.Writer) error) error {
	zlibOut := zlib.NewWriter(out)
	if err := fn(zlibOut); err != nil {
		return err
	}
	if err := zlibOut.Close(); err != nil {
		return fmt.Errorf("failed to complete compression: %w", err)
	}
	return nil
}

func readCompressed(in io.Reader, fn func(compIn io.Reader) error) error {
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
