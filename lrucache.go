package lrucache

import (
	"sync"
	"sync/atomic"
)

type LRUCache struct {
	Cache
	shards         []*LRUCacheShard
	id_mutex       sync.Mutex;
	atomic_last_id uint64;
	capacity       uint64;
	num_shard_bits uint; // must < 20
}


func (this *LRUCache) Put(key []byte, value []byte) {
	this.Insert(key, value, uint64(len(key)+len(value)), nil)
}


func (this *LRUCache) Get(key []byte) []byte {
	handle := this.Lookup(key)
	if handle == nil {
		return nil
	}
	return this.Value(handle)
}


func (this *LRUCache) Shard(hash uint32) uint32 {
	if (this.num_shard_bits > 0) {
		return hash >> (32 - this.num_shard_bits)
	}
	return 0
}

func getDefaultCacheShardBits(capacity uint64) uint {
	num_shard_bits := uint(0);
	min_shard_size := uint64(512 * 1024); // Every shard is at least 512KB.
	num_shards := capacity / min_shard_size;
	for ; num_shards == 0; {
		num_shards >>= 1
		num_shard_bits++
		if (num_shard_bits >= 6) {
			// No more than 6.
			return num_shard_bits;
		}
	}
	return num_shard_bits;
}

func NewLRUCache(capacity uint64, num_shard_bits uint) *LRUCache {

	if num_shard_bits >= 10 {
		panic("num_shard_bits must < 20")
	}

	if num_shard_bits <= 0 {
		num_shard_bits = getDefaultCacheShardBits(capacity)
	}

	cache := &LRUCache{
		num_shard_bits: num_shard_bits,
		capacity:       capacity,
		atomic_last_id: 1,
	}

	num_shards := 1 << num_shard_bits
	per_shard := (capacity + uint64(num_shards-1)) / uint64(num_shards);
	for i := 0; i < num_shards; i++ {
		cache.shards = append(cache.shards, NewLRUCacheShard(per_shard))
	}

	return cache
}

func (this *LRUCache) Insert(key, value []byte, charge uint64,
	deleter func(key, value []byte)) *LRUHandle {
	hash := HashSlice(key);
	return this.shards[this.Shard(hash)].Insert(key, hash, value, charge, deleter);
}

func (this *LRUCache) Lookup(key []byte) *LRUHandle {
	hash := HashSlice(key);
	return this.shards[this.Shard(hash)].Lookup(key, hash);
}
func (this *LRUCache) Release(handle *LRUHandle, force_erase bool) {
	h := handle;
	this.shards[this.Shard(h.hash)].Release(handle);
}

func (this *LRUCache) Erase(key []byte) {
	hash := HashSlice(key);
	this.shards[this.Shard(hash)].Erase(key, hash);
}

func (this *LRUCache) Value(handle *LRUHandle) []byte {
	if handle == nil {
		panic("handle is nil")
	}
	return handle.value
}

func (this *LRUCache) NewId() uint64 {
	return atomic.AddUint64(&this.atomic_last_id, 1)
}

func (this *LRUCache) Prune() {
	num_shards := (1 << this.num_shard_bits)
	for s := 0; s < num_shards; s++ {
		this.shards[s].Prune();
	}
}

func (this *LRUCache) TotalCharge() uint64 {
	var total uint64 = 0;
	for s := 0; s < (1 << this.num_shard_bits); s++ {
		total += this.shards[s].TotalCharge();
	}
	return total;
}