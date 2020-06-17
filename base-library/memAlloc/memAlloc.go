package memAlloc

type MemoryAllocator interface {
	Get(n int) []byte
	Put([]byte)
}
