package lrucache

import (
	"sync/atomic"
)

// impl of interface Cache
type LRUCache struct {
	shards         []*LRUCacheShard
	atomic_last_id uint64;
	capacity       uint64;
	num_shard_bits uint; // must < 10
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

func (this *LRUCache) Put(key, value string) {
	this.Insert([]byte(key), (value), uint64(len(key)+len(value)), nil)
}

func (this *LRUCache) Get(key string) (string, bool) {
	value := this.Lookup([]byte(key))
	res,ok := value.(string)
	if !ok {
		return "", false
	}
	return res, true
}

func (this *LRUCache) Delete(key string){
	this.Remove([]byte(key))
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

func (this *LRUCache) shard(hash uint32) uint32 {
	if (this.num_shard_bits > 0) {
		return hash >> (32 - this.num_shard_bits)
	}
	return 0
}

func (this *LRUCache) Insert(key []byte, entry interface{}, charge uint64,	deleter deleter_callback) {
	hash := HashSlice(key);
	this.shards[this.shard(hash)].Insert(key, hash, entry, charge, deleter);
}

func (this *LRUCache) Lookup(key []byte) interface{} {
	hash := HashSlice(key);
	return this.shards[this.shard(hash)].Lookup(key, hash);
}

func (this *LRUCache) Remove(key []byte) interface{} {
	hash := HashSlice(key);
	return this.shards[this.shard(hash)].Remove(key, hash);
}


func (this *LRUCache) Merge(key []byte, entry interface{}, charge uint64, merge_opt merge_operator, charge_opt charge_operator) {
	hash := HashSlice(key);
	this.shards[this.shard(hash)].Merge(key, hash, entry, charge, merge_opt, charge_opt);
}

func getDefaultCacheShardBits(capacity uint64) uint {
	num_shard_bits := uint(0);
	min_shard_size := uint64(512 * 1024); // Every shard is at least 512KB.
	num_shards := capacity / min_shard_size;
	for ; num_shards != 0; {
		num_shards >>= 1
		num_shard_bits++
		if (num_shard_bits >= 6) {
			// No more than 6.
			return num_shard_bits;
		}
	}
	return num_shard_bits;
}
