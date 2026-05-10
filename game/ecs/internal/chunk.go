package internal

const chunkSize = 128

type DataChunk[T any] *[chunkSize]T
