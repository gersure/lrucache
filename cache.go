package lrucache

type Cache interface {
	Put(key string, value string)
	Get(key string) string
	Delete(key string) []string
	NewId() uint64
	Prune()
	TotalCharge() uint64

	Insert(key, value []byte, charge uint64,
		deleter func(key, value []byte))
	Lookup(key []byte) []byte
	Remove(key []byte) []byte
}
