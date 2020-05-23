package asset

import (
	"encoding/gob"
	"fmt"
	"io"
)

type Program struct {
	VertexSourceCode   string
	FragmentSourceCode string
}

func EncodeProgram(out io.Writer, program *Program) error {
	return WriteCompressed(out, func(compOut io.Writer) error {
		if err := gob.NewEncoder(compOut).Encode(program); err != nil {
			return fmt.Errorf("failed to encode gob stream: %w", err)
		}
		return nil
	})
}

func DecodeProgram(in io.Reader, program *Program) error {
	return ReadCompressed(in, func(compIn io.Reader) error {
		if err := gob.NewDecoder(compIn).Decode(program); err != nil {
			return fmt.Errorf("failed to decode gob stream: %w", err)
		}
		return nil
	})
}
