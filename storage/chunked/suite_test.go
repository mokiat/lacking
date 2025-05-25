package chunked_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mokiat/gog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChunked(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chunked Suite")
}

type IDChunk struct {
	Name string
}

func (IDChunk) ChunkID() uuid.UUID {
	return gog.Must(uuid.Parse("28c55f8a-0828-47cb-9a8d-6512569dc113"))
}

type LocationChunk struct {
	X uint64
	Y uint64
}

func (LocationChunk) ChunkID() uuid.UUID {
	return gog.Must(uuid.Parse("e19ae4e5-3b8d-4512-a707-9f992ee8f126"))
}

type PriorityChunk struct {
	Priority uint64
}

func (PriorityChunk) ChunkID() uuid.UUID {
	return gog.Must(uuid.Parse("d0d6c1f5-0798-4a69-bf1b-ae15f604a91b"))
}
