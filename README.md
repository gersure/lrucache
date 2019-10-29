# lrucache
golang impl lru cache reference leveldb

* it's a lru chche
* very easy to use
* good performance
* adapt most application scenario
* support different namespace(only support 10byte)

this's (1<<num_shard_bits(<10)) hash table in cache; modify every hash table use sync.mutex；so it's provide good performance
when memory use up to capacity; Earliest insert will be drop

## how to use
```go
        // before use must init cache to set capacity and bitset
		InitLRUCache(test.capacity, 0)
        // this is default namespace [10]byte{}
		lru := DefaultLRUCache()
        
        // to create a new namespace name1
        lru_name1 := NewLRUCache("names1")
        ...
        // then use lru_name1 to put or delete
```
### method 1
```go
        
    	lru.Put("key", "value")
    	value := lru.Get("key")
    	//displayed remove
    	lru.Delete("key")
```
### method 2
```go

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

### method 3; use like redis incr;but merge return old value
```go
	key := []byte("key")
	var merge_value int = 1
	for i := 0; i < 1000; i++ {
		lru.Merge(key, merge_value, 4, IntMergeOperator, IntChargeOperator) // real value = value+1
	}
```

### more use case, you can see lrucache_test.go
