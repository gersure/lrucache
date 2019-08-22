package lrucache

import (
	"sync"
	"sync/atomic"
)

const namespace_byte_len  = 10

type name_space [10]byte

// impl of interface Cache
type lru_cache struct {
	shards         []*LRUCacheShard
	atomic_last_id uint64;
	capacity       uint64;
	num_shard_bits uint; // must < 10
	namespaces     map[name_space]*LRUCache
	mutex          sync.Mutex
}

type LRUCache struct {
	lru_cache
	namespace name_space
}

var s_lru_cache *lru_cache = nil

func InitLRUCache(capacity uint64, num_shard_bits uint)  {

	if num_shard_bits >= 10 {
		panic("num_shard_bits must < 10")
	}

	if num_shard_bits <= 0 {
		num_shard_bits = getDefaultCacheShardBits(capacity)
	}

	cache := &lru_cache{
		num_shard_bits: num_shard_bits,
		capacity:       capacity,
		atomic_last_id: 1,
	}

	num_shards := 1 << num_shard_bits
	per_shard := getPerfShardCapacity(capacity, num_shard_bits);
	for i := 0; i < num_shards; i++ {
		cache.shards = append(cache.shards, NewLRUCacheShard(per_shard))
	}

	s_lru_cache = cache
}

func DefaultLRUCache() *LRUCache  {
	if s_lru_cache == nil {
		panic("use LRUCache must InitLRUCache first")
	}

	return &LRUCache{
		lru_cache: *s_lru_cache,
		namespace: [10]byte{},
	}
}

/**
	namespace max len is 10; and key's namespace's charge doesn't caculate
 */
func NewLRUCache(namespace string) (*LRUCache, bool) {
	if s_lru_cache == nil {
		panic("use LRUCache must InitLRUCache first")
	}

	var tmp name_space
	for i:=0; i<namespace_byte_len && i<len(namespace); i++ {
		tmp[i] = namespace[i]
	}

	return getNamespace(tmp)
}

func getNamespace(namespace name_space) (*LRUCache, bool) {
	var cache *LRUCache = nil
	if  cache, ok := s_lru_cache.namespaces[namespace]; !ok {
		s_lru_cache.mutex.Lock();
		defer s_lru_cache.mutex.Unlock();
		if  _, ok := s_lru_cache.namespaces[namespace]; !ok {
			cache = &LRUCache{
				lru_cache: *s_lru_cache,
				namespace: namespace,
			}
			s_lru_cache.namespaces[namespace] = cache
			return cache, true
		}
	}
	return cache, false
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
	this.mutex.Lock();
	defer this.mutex.Unlock();
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

func (this *LRUCache) Insert(key []byte, entry interface{}, charge uint64,	deleter DeleteCallback) {
	realkey := keyAdaptNamespace(key, this.namespace)
	hash := HashSlice(realkey);
	this.shards[this.shard(hash)].Insert(realkey, hash, entry, charge, deleter);
}

func (this *LRUCache) Lookup(key []byte) interface{} {
	realkey := keyAdaptNamespace(key, this.namespace)
	hash := HashSlice(realkey);
	return this.shards[this.shard(hash)].Lookup(realkey, hash);
}

func (this *LRUCache) Remove(key []byte) interface{} {
	realkey := keyAdaptNamespace(key, this.namespace)
	hash := HashSlice(realkey);
	return this.shards[this.shard(hash)].Remove(realkey, hash);
}


func (this *LRUCache) Merge(key []byte, entry interface{}, charge uint64, merge_opt MergeOperator, charge_opt ChargeOperator) (interface{}) {
	realkey := keyAdaptNamespace(key, this.namespace)
	hash := HashSlice(realkey);
	return this.shards[this.shard(hash)].Merge(realkey, hash, entry, charge, merge_opt, charge_opt);
}

func (this *LRUCache) ApplyToAllCacheEntries(travel_fun TravelEntryOperator) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	for _, shard := range this.shards {
		shard.ApplyToAllCacheEntries(travel_fun)
	}
}

func (this *LRUCache) SetCapacity(capacity uint64)  {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	per_shard := getPerfShardCapacity(capacity, this.num_shard_bits)
	for _, shard := range this.shards {
		shard.SetCapacity(per_shard)
	}
}

func getPerfShardCapacity(capacity uint64, num_shard_bits uint) uint64 {
	num_shards := 1 << num_shard_bits
	return (capacity + uint64(num_shards-1)) / uint64(num_shards);
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

func keyAdaptNamespace(key []byte, namespace name_space) []byte  {
	var real_key []byte
	real_key = append(real_key, namespace[:]...)
	return append(real_key, key...)
}