package lrucache

type LRUHandle struct {
	value     []byte
	deleter   func(key, value []byte)
	next_hash *LRUHandle
	next      *LRUHandle;
	prev      *LRUHandle;
	charge    uint64; // TODO(opt): Only allow uint32_t?
	in_cache  bool;   // Whether entry is in the cache.
	refs      uint32; // References, including cache reference, if present.
	hash      uint32; // Hash of key(); used for fast sharding and comparisons
	key  []byte; // Beginning of key
}
