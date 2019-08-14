package lrucache

type handle_deleter func(key []byte, value interface{})

type Cache interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	Delete(key string)
	NewId() uint64
	Prune()
	TotalCharge() uint64

	Insert(key[]byte, entry interface{}, charge uint64, deleter handle_deleter)
	Lookup(key []byte) interface{}
	Remove(key []byte) interface{}
}
