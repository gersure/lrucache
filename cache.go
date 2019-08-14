package lrucache

type Cache interface {
	Name() string

	Put(key, value []byte)
	Get(key []byte) []byte

	Insert(key, value []byte, deleter func(key, value []byte)) *LRUHandle
	Lookup(key []byte) *LRUHandle
	Release(handle *LRUHandle)
	Value(handle *LRUHandle) []byte
	Erase(key []byte)
	NewId() uint64
	Prune()
	TotalCharge() uint64
}
