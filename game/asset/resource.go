package asset

import (
	"compress/zlib"
	"fmt"
	"io"

	"github.com/mokiat/gblob"
)

type headerFlag uint16

const (
	headerFlagZlib headerFlag = 1 << iota
)

type header struct {
	Version uint16
	Flags   headerFlag
}

func (h header) HasFlag(flag headerFlag) bool {
	return (h.Flags & flag) == flag
}

func (h *header) EncodeTo(out io.Writer) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(h)
}

func (h *header) DecodeFrom(in io.Reader) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(h)
}

type versionedEncodable interface {
	encodeVersionTo(out io.Writer, version uint16) error
}

func encodeResource(out io.Writer, h header, resource versionedEncodable) error {
	if err := h.EncodeTo(out); err != nil {
		return fmt.Errorf("error encoding header: %w", err)
	}
	switch {
	case h.HasFlag(headerFlagZlib):
		return writeZlibCompressed(out, func(compOut io.Writer) error {
			if err := resource.encodeVersionTo(compOut, h.Version); err != nil {
				return fmt.Errorf("error encoding resource: %w", err)
			}
			return nil
		})
	default:
		if err := resource.encodeVersionTo(out, h.Version); err != nil {
			return fmt.Errorf("error encoding resource: %w", err)
		}
	}
	return nil
}

type versionedDecodable interface {
	decodeVersionFrom(in io.Reader, version uint16) error
}

func decodeResource(in io.Reader, resource versionedDecodable) error {
	var h header
	if err := h.DecodeFrom(in); err != nil {
		return fmt.Errorf("error decoding header: %w", err)
	}
	switch {
	case h.HasFlag(headerFlagZlib):
		return readZlibCompressed(in, func(compIn io.Reader) error {
			if err := resource.decodeVersionFrom(compIn, h.Version); err != nil {
				return fmt.Errorf("error decoding resource: %w", err)
			}
			return nil
		})
	default:
		if err := resource.decodeVersionFrom(in, h.Version); err != nil {
			return fmt.Errorf("error decoding resource: %w", err)
		}
	}
	return nil
}

func writeZlibCompressed(out io.Writer, fn func(io.Writer) error) error {
	zlibOut := zlib.NewWriter(out)
	if err := fn(zlibOut); err != nil {
		return err
	}
	if err := zlibOut.Close(); err != nil {
		return fmt.Errorf("error to complete compression: %w", err)
	}
	return nil
}

func readZlibCompressed(in io.Reader, fn func(io.Reader) error) error {
	zlibIn, err := zlib.NewReader(in)
	if err != nil {
		return fmt.Errorf("error to create decompressor: %w", err)
	}
	if err := fn(zlibIn); err != nil {
		return err
	}
	if err := zlibIn.Close(); err != nil {
		return fmt.Errorf("error to complete decompression: %w", err)
	}
	return nil
}
