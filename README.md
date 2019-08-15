# lrucache
golang impl lru cache reference leveldb

* it's a lru chche
* very easy to use
* good performance
* adapt most application scenario

this's (1<<num_shard_bits(<10)) hash table in cache; modify every hash table use sync.mutex；so it's provide good performance
when memory use up to capacity; Earliest insert will be drop

## how to use
### method 1
```go
    	// use get set delete (just provide string type support)
    	lru := NewLRUCache(1024*1024/*capacity*/, 0/*num shard bits*/) // num_shard_bit is 0, code will auto make one
		
    	lru.Put("key", "value")
    	value := lru.Get("key")
    	//displayed remove
    	lru.Delete("key")
```
### method 2
```go

	lru := NewLRUCache(1024*1024 /*capacity*/, 0 /*num shard bits*/) // num_shard_bit is 0, code will auto make one

	key := []byte("key")
	type V struct {a int; b int}
	value := V{4, 5}

	lru.Insert(key, value, uint64(len(key)+4*2), func(key []byte, entry interface{}) {
		fmt.Println("key:%s will be deleted from cache", key)
	})
	origin := lru.Lookup(key)
	if origin != nil {
		origin_value := origin.(V)
		fmt.Println(origin_value)
	}
```

### more use case, you can see lrucache_test.go
