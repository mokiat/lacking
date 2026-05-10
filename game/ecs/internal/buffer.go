package internal

import "unsafe"

func NewBuffer(initialCapacity int) *Buffer {
	return &Buffer{
		data: make([]byte, initialCapacity),
	}
}

type Buffer struct {
	data        []byte
	writeOffset uintptr
	readOffset  uintptr
}

func (b *Buffer) HasMoreData() bool {
	return b.writeOffset > b.readOffset
}

func (b *Buffer) Reset() {
	b.writeOffset = 0
	b.readOffset = 0
}

func (b *Buffer) ensure(itemSize int) {
	required := int(b.writeOffset) + itemSize
	available := len(b.data)
	if required > available {
		b.data = append(b.data, make([]byte, required-available)...)
		b.data = b.data[:cap(b.data)]
	}
}

func WriteToBuffer[T any](buffer *Buffer, command T) uint32 {
	size := unsafe.Sizeof(command)
	buffer.ensure(int(size))

	target := (*T)(unsafe.Add(unsafe.Pointer(&buffer.data[0]), buffer.writeOffset))
	*target = command

	result := uint32(buffer.writeOffset)
	buffer.writeOffset += size
	return result
}

func ReadFromBuffer[T any](buffer *Buffer) T {
	target := (*T)(unsafe.Add(unsafe.Pointer(&buffer.data[0]), buffer.readOffset))
	command := *target
	buffer.readOffset += unsafe.Sizeof(command)
	return command
}

func ReadFromBufferOffset[T any](buffer *Buffer, offset uint32) T {
	target := (*T)(unsafe.Add(unsafe.Pointer(&buffer.data[0]), offset))
	command := *target
	return command
}
