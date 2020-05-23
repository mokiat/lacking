package asset

import "io"

type Program struct {
	VertexSourceCode   string
	FragmentSourceCode string
}

func EncodeProgram(out io.Writer, program *Program) error {
	return Encode(out, program)
}

func DecodeProgram(in io.Reader, program *Program) error {
	return Decode(in, program)
}
