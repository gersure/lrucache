package lrucache

import (
	"bytes"
	"strconv"
	"testing"
)

var case_shard_bits = []struct{
	capacity uint64
	num_bits uint
}{
	{512*1024*1024, 6},
	{512*1024*10, 4},
	{512*1024*6, 3},
	{512*1024*3, 2},
	{512*1024*1, 1},
}

func TestGetDefaultCacheShardBits(t *testing.T)  {
	for _, test := range case_shard_bits {
		bits := getDefaultCacheShardBits(test.capacity)
		if bits != test.num_bits {
			t.Errorf("getDefaultCacheShardBits error, capacity is: %v," +
				" expected: %d, got: %d", test.capacity, test.num_bits, bits)
		}
	}
}

var case_cache = []struct{
	key []byte
	value string
	charge uint64
	deleter DeleteCallback
} {
	{[]byte("key0"),("value0"), 10, nil},
	{[]byte("key1"),("value1"), 15, nil},
	{[]byte("key2"),("value2"), 30, nil},
	{[]byte("key3"),("value3"), 20, nil},
	{[]byte("key4"),("value4"), 10, nil},
	{[]byte("key5"),("value5"), 10, nil},
	{[]byte("key6"),("value6"), 10, nil},
	{[]byte("key7"),("value7"), 10, nil},
	{[]byte("key8"),("value8"), 10, nil},
	{[]byte("key9"),("value9"), 10, nil},
	{[]byte("key10"),("value10"), 10, nil},
	{[]byte("key11"),("value11"), 10, nil},
	{[]byte("key12"),("value12"), 10, nil},
	{[]byte("key13"),("value13"), 10, nil},
	{[]byte("key14"),("value14"), 10, nil},
	{[]byte("key15"),("value15"), 10, nil},
	{[]byte("key16"),("value16"), 10, nil},
	{[]byte("key17"),("value17"), 10, nil},
	{[]byte("key18"),("value18"), 10, nil},
	{[]byte("key19"),("value19"), 10, nil},
}

func TestInitLRUCache(t *testing.T) {
	for _, test := range case_shard_bits {
		InitLRUCache(test.capacity, 0)
		lru := DefaultLRUCache()
		if len(lru.shards) != (1<<test.num_bits) {
			t.Errorf("InitLRUCache error, capacity is: %v," +
				" shards expected: %d, got: %d", test.capacity, 1<<test.num_bits, len(lru.shards) )
		}
		if lru.TotalCharge() != 0 {
			t.Errorf("totalcharge init error, got:%v", lru.TotalCharge())
		}


		if _, ok := lru.Get(("test"));ok {
			t.Errorf("empty cache lookup isn't nil")
		}

		if lru.Lookup([]byte("test")) != nil {
			t.Errorf("empty cache lookup isn't nil")
		}

		if lru.Remove([]byte("test")) != nil {
			t.Errorf("empty cache lookup isn't nil")
		}

		lru.Insert([]byte("test"), []byte("value"), 10, nil)

		lru.Put(("test"), ("value"))
	}

}


func TestLRUCache_PutGetDelete(t *testing.T) {
	for _, test := range case_shard_bits {
		InitLRUCache(test.capacity, 0)

		lru := DefaultLRUCache()
		var total_charge uint64= 0
		for _, test_bar := range case_cache {
			lru.Put(string(test_bar.key[:]), (test_bar.value))
			total_charge += (uint64(len(test_bar.key)+len(test_bar.value)) + namespace_byte_len)

			origin, _ := lru.Get(string(test_bar.key))
			if origin != test_bar.value {
				t.Errorf("put key:%s ,value:%s, got value:%s", (test_bar.key), (test_bar.value), (origin))
			}
			if lru.TotalCharge() != total_charge {
				t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
			}
		}
	}


	var total_charge uint64= 0
	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Put(string(test_bar.key), string(test_bar.value))
		total_charge += uint64(len(test_bar.key)+len(test_bar.value) + namespace_byte_len)
		origin,_ := lru.Get(string(test_bar.key))
		if origin != string(test_bar.value) {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}

		lru.Delete(string(test_bar.key))
		total_charge -= uint64(len(test_bar.key)+len(test_bar.value) + namespace_byte_len)
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}

		if _, ok := lru.Get(string(test_bar.key)); ok {
			t.Errorf("deleted key, still get ok; key:%s", test_bar.key)
		}

	}

}


func TestLRUCache_InsertLookupRemove(t *testing.T) {
	for _, test := range case_shard_bits {
		InitLRUCache(test.capacity, 0)
		lru := DefaultLRUCache()
		var total_charge uint64= 0
		for _, test_bar := range case_cache {
			lru.Insert(test_bar.key, test_bar.value, test_bar.charge, test_bar.deleter)
			total_charge += (test_bar.charge + namespace_byte_len)

			origin := lru.Lookup(test_bar.key)
			origin,_ = origin.(string)
			if (origin != test_bar.value) {
				t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
			}
			if lru.TotalCharge() != total_charge {
				t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
			}
		}
	}


	var total_charge uint64= 0
	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Insert(test_bar.key, test_bar.value, test_bar.charge, test_bar.deleter)
		total_charge += (test_bar.charge + namespace_byte_len)
		origin := lru.Lookup(test_bar.key)
		origin,_ = origin.(string)
		if origin != test_bar.value {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}

		origin = lru.Remove(test_bar.key)
		origin,_ = origin.(string)
		total_charge -= (test_bar.charge + namespace_byte_len)
		if (origin !=test_bar.value) {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}
	}

}

func TestLRUCache_Deleter(t *testing.T) {

	var delete = 0
	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Insert(test_bar.key, test_bar.value, test_bar.charge, func(key []byte, entry interface{}) {
			delete ++
			entry,_ = entry.(string)
			if bytes.HasSuffix(test_bar.key, key) || test_bar.value != entry {
				t.Errorf("put key: %s ,value: %s\n" +
					"got key: %s, value: %s", (test_bar.key), (test_bar.value), key, entry)
			}
		})
		lru.Remove(test_bar.key)
	}

	if delete != len(case_cache) {
		t.Error("may be do't call deleter")
	}
}


func TestLRUCache_LRUCharge(t *testing.T) {

	var capacity uint64 = 1024

	InitLRUCache(1024*1024, 1)
	lru_cache := lru_cache{
		num_shard_bits: 0,
		capacity:       capacity,
		atomic_last_id: 1,
	}
	lru := LRUCache{
		lru_cache:lru_cache,
		namespace:name_space{},
	}
	num_shards := 1
	per_shard := (capacity + uint64(num_shards-1)) / uint64(num_shards);
	for i := 0; i < num_shards; i++ {
		lru.shards = append(lru.shards, NewLRUCacheShard(per_shard))
	}

	var total_charge uint64 = 0
	var now_deleted int = 0
	for i:=0; i<10000; i++ {
		key := []byte(strconv.FormatInt(int64(i), 10))
		lru.Insert(key, key, 10, nil)
		total_charge += (10 + namespace_byte_len)

		if total_charge > (per_shard*uint64(num_shards)) {
			if (now_deleted > 0) {
				now_deleted_key := []byte(strconv.FormatInt(int64(now_deleted), 10))
				if _, ok := lru.Get(string(now_deleted_key)); ok {
					t.Errorf("now total charge: %v, but got before key:%s", lru.TotalCharge(), now_deleted_key)
				}
			}
			now_deleted ++
		}
	}
}


func TestLRUCache_MergeAddInt(t *testing.T) {

	var capacity uint64 = 1024 * 1024

	key := []byte("key")
	var value int = 0
	var merge_value int = 1
	var res_total= 0
	InitLRUCache(capacity, 1)
	lru := DefaultLRUCache()
	lru.Insert(key, value, 4, nil)
	for i := 0; i < 1000; i++ {
		lru.Merge(key, merge_value, 4, IntMergeOperator, IntChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res, _ := res.(int)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}

	merge_value = merge_value * (-1)
	for i := 0; i < 1000; i++ {
		lru.Merge(key, merge_value, 4, IntMergeOperator, IntChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res, _ := res.(int)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}

	lru.Remove(key)
	for i := 0; i < 1000; i++ {
		lru.Merge(key, merge_value, 4, IntMergeOperator, IntChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res, _ := res.(int)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}

}


func TestLRUCache_MergeAddInt64(t *testing.T) {

	var capacity uint64 = 1024*1024

	key := []byte("key64")
	var value int64 = 0
	var merge_value int64 = 1
	var res_total int64= 0
	InitLRUCache(capacity, 1)
	lru := DefaultLRUCache()
	lru.Insert(key, value, 4, nil)
	for i:=0 ; i<1000; i++ {
		lru.Merge(key, merge_value, 4, Int64MergeOperator, Int64ChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res,_ := res.(int64)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}

	merge_value = merge_value*(-1)
	for i:=0 ; i<1000; i++ {
		lru.Merge(key, merge_value, 4, Int64MergeOperator, Int64ChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res,_ := res.(int64)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}

	lru.Remove(key)
	for i:=0 ; i<1000; i++ {
		lru.Merge(key, merge_value, 4, Int64MergeOperator, Int64ChargeOperator)
		res_total += merge_value
		res := lru.Lookup(key)
		add_res,_ := res.(int64)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%d, got:%d", res_total, add_res)
		}
	}
}


func TestLRUCache_MergeAppend(t *testing.T) {

	var capacity uint64 = 1024*1024
	var merge_opt MergeOperator = func(old_entry, new_entry interface{}) interface{} {
		old, _ := old_entry.(string)
		new, _ := new_entry.(string)
		res := old + new
		return res
	}

	var charge_opt ChargeOperator = func(entry interface{}, old_charge, new_charge uint64) uint64 {
		return old_charge+new_charge
	}

	key := []byte("key")
	merge_value := "1"
	var res_total string
	var capacity_totoal uint64 = 0
	var res_string string

	InitLRUCache(capacity, 1)
	lru := DefaultLRUCache()
	for i:=0 ; i<100; i++ {
		old_origin := lru.Merge(key, merge_value, uint64(len(merge_value)), merge_opt, charge_opt)
		res_string += merge_value
		capacity_totoal += uint64(len(merge_value))
		res_total += merge_value
		res := lru.Lookup(key)
		add_res,_ := res.(string)
		if add_res != res_total {
			t.Errorf("merge operator error expected:%s, got:%s", res_total, add_res)
		}

		if lru.TotalCharge() != capacity_totoal {
			t.Errorf("merge charge operator maybe error;expected:%v, got:%v", capacity_totoal, lru.TotalCharge())
		}

		old_origin,_ = old_origin.(string)
		if len(res_string) >=2 && old_origin != res_string[:len(res_string)-1] {
			t.Errorf("merge return old entry error; expected:%s, got:%s", res_string[:(len(res_string)-1)], old_origin)
		}
	}
}

func TestLRUCache_ApplyToAllCacheEntries(t *testing.T) {
	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Put(string(test_bar.key), string(test_bar.value))
	}
	var case_count = 0
	lru.ApplyToAllCacheEntries(func(key []byte, entry interface{}) {
		case_count++;
	})

	if case_count != len(case_cache) {
		t.Errorf("ApplyToAllCacheEntries error, apply count:%d, but this's:%d", case_count, len(case_cache))
	}
}

func TestLRUCache_SetCapacity(t *testing.T) {

	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Put(string(test_bar.key), string(test_bar.value))
	}
	var case_count = 0
	lru.SetCapacity(uint64(0))

	lru.ApplyToAllCacheEntries(func(key []byte, entry interface{}) {
		case_count++;
	})

	if case_count != 0 {
		t.Errorf("trun off cache error, this's [%d] entry in cache", case_count)
	}

	if lru.TotalCharge() != 0 {
		t.Errorf("turn off cache, but totalusage already has:%v", lru.TotalCharge())
	}
}


func TestLRUCache_Prune(t *testing.T) {

	InitLRUCache(1024*1024, 1)
	lru := DefaultLRUCache()
	for _, test_bar := range case_cache {
		lru.Put(string(test_bar.key), string(test_bar.value))
	}
	var case_count = 0
	lru.Prune()

	lru.ApplyToAllCacheEntries(func(key []byte, entry interface{}) {
		case_count++;
	})

	if case_count != 0 {
		t.Errorf("trun off cache error, this's [%d] entry in cache", case_count)
	}

	if lru.TotalCharge() != 0 {
		t.Errorf("turn off cache, but totalusage already has:%v", lru.TotalCharge())
	}
}