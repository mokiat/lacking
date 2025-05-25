package gendto

type GenChunkHolder struct {
	Gen *GenChunk `chunk:"lacking:gen"`
}

type GenChunk struct {
	Digest string
}
