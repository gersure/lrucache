package lrucache

type Cache interface {
	Put(key []byte, value []byte)
	Get(key []byte) []byte
	Delete(key []byte) []byte
	NewId() uint64
	Prune()
	TotalCharge() uint64

	Insert(key, value []byte, charge uint64,
		deleter func(key, value []byte))
	Lookup(key []byte) []byte
	Remove(key []byte) []byte
}
