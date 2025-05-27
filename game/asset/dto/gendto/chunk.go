package gendto

var GenChunkID = "lacking:gen"

type GenChunkHolder struct {
	Gen *GenChunk `chunk:"lacking:gen"`
}

type GenChunk struct {
	Digest string
}
