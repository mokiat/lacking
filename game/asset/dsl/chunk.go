package dsl

var genChunkID = "lacking:gen"

type genChunkHolder struct {
	Gen *genChunk `chunk:"lacking:gen"`
}

type genChunk struct {
	Digest string
}
