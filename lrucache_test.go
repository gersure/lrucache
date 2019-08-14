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
	key, value []byte
	charge uint64
	deleter func(key, value []byte)
} {
	{[]byte("key0"),[]byte("value0"), 10, nil},
	{[]byte("key1"),[]byte("value1"), 15, nil},
	{[]byte("key2"),[]byte("value2"), 30, nil},
	{[]byte("key3"),[]byte("value3"), 20, nil},
	{[]byte("key4"),[]byte("value4"), 10, nil},
	{[]byte("key5"),[]byte("value5"), 10, nil},
	{[]byte("key6"),[]byte("value6"), 10, nil},
	{[]byte("key7"),[]byte("value7"), 10, nil},
	{[]byte("key8"),[]byte("value8"), 10, nil},
	{[]byte("key9"),[]byte("value9"), 10, nil},
	{[]byte("key10"),[]byte("value10"), 10, nil},
	{[]byte("key11"),[]byte("value11"), 10, nil},
	{[]byte("key12"),[]byte("value12"), 10, nil},
	{[]byte("key13"),[]byte("value13"), 10, nil},
	{[]byte("key14"),[]byte("value14"), 10, nil},
	{[]byte("key15"),[]byte("value15"), 10, nil},
	{[]byte("key16"),[]byte("value16"), 10, nil},
	{[]byte("key17"),[]byte("value17"), 10, nil},
	{[]byte("key18"),[]byte("value18"), 10, nil},
	{[]byte("key19"),[]byte("value19"), 10, nil},
}

func TestNewLRUCache(t *testing.T) {
	for _, test := range case_shard_bits {
		lru := NewLRUCache(test.capacity, 0)
		if len(lru.shards) != (1<<test.num_bits) {
			t.Errorf("NewLRUCache error, capacity is: %v," +
				" shards expected: %d, got: %d", test.capacity, 1<<test.num_bits, len(lru.shards) )
		}
		if lru.TotalCharge() != 0 {
			t.Errorf("totalcharge init error, got:%v", lru.TotalCharge())
		}


		if lru.Get([]byte("test")) != nil {
			t.Errorf("empty cache lookup isn't nil")
		}

		if lru.Lookup([]byte("test")) != nil {
			t.Errorf("empty cache lookup isn't nil")
		}

		if lru.Remove([]byte("test")) != nil {
			t.Errorf("empty cache lookup isn't nil")
		}

		lru.Insert([]byte("test"), []byte("value"), 10, nil)

		lru.Put([]byte("test"), []byte("value"))
	}

}


func TestLRUCache_PutGetDelete(t *testing.T) {
	for _, test := range case_shard_bits {
		lru := NewLRUCache(test.capacity, 0)
		var total_charge uint64= 0
		for _, test_bar := range case_cache {
			lru.Put(test_bar.key, test_bar.value)
			total_charge += uint64(len(test_bar.key)+len(test_bar.value))

			origin := lru.Get(test_bar.key)
			if bytes.Compare(origin, test_bar.value) != 0 {
				t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
			}
			if lru.TotalCharge() != total_charge {
				t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
			}
		}
	}


	var total_charge uint64= 0
	lru := NewLRUCache(1024*1024, 1)
	for _, test_bar := range case_cache {
		lru.Put(test_bar.key, test_bar.value)
		total_charge += uint64(len(test_bar.key)+len(test_bar.value))
		origin := lru.Get(test_bar.key)
		if bytes.Compare(origin, test_bar.value) != 0 {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}

		origin = lru.Delete(test_bar.key)
		total_charge -= uint64(len(test_bar.key)+len(test_bar.value))
		if bytes.Compare(origin, test_bar.value) != 0 {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}
	}

}


func TestLRUCache_InsertLookupRemove(t *testing.T) {
	for _, test := range case_shard_bits {
		lru := NewLRUCache(test.capacity, 0)
		var total_charge uint64= 0
		for _, test_bar := range case_cache {
			lru.Insert(test_bar.key, test_bar.value, test_bar.charge, test_bar.deleter)
			total_charge += test_bar.charge

			origin := lru.Lookup(test_bar.key)
			if bytes.Compare(origin, test_bar.value) != 0 {
				t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
			}
			if lru.TotalCharge() != total_charge {
				t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
			}
		}
	}


	var total_charge uint64= 0
	lru := NewLRUCache(1024*1024, 1)
	for _, test_bar := range case_cache {
		lru.Insert(test_bar.key, test_bar.value, test_bar.charge, test_bar.deleter)
		total_charge += test_bar.charge
		origin := lru.Lookup(test_bar.key)
		if bytes.Compare(origin, test_bar.value) != 0 {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}

		origin = lru.Remove(test_bar.key)
		total_charge -= test_bar.charge
		if bytes.Compare(origin, test_bar.value) != 0 {
			t.Errorf("put key: %s ,value : %s, got value : %s", (test_bar.key), (test_bar.value), (origin))
		}
		if lru.TotalCharge() != total_charge {
			t.Errorf("total charge expected: %v, got: %v", total_charge, lru.TotalCharge())
		}
	}

}

func TestLRUCache_Deleter(t *testing.T) {

	var delete = 0
	lru := NewLRUCache(1024*1024, 1)
	for _, test_bar := range case_cache {
		lru.Insert(test_bar.key, test_bar.value, test_bar.charge, func(key, value []byte) {
			delete ++
			if bytes.Compare(test_bar.key, key) != 0  ||
				bytes.Compare(test_bar.value, value) != 0 {
				t.Errorf("put key: %s ,value: %s\n" +
					"got key: %s, value: %s", (test_bar.key), (test_bar.value), key, value)
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

	lru := &LRUCache{
		num_shard_bits: 0,
		capacity:       capacity,
		atomic_last_id: 1,
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
		total_charge += 10

		if total_charge > (per_shard*uint64(num_shards)) {
			if (now_deleted > 0) {
				now_deleted_key := []byte(strconv.FormatInt(int64(now_deleted), 10))
				if lru.Get(now_deleted_key) != nil {
					t.Errorf("now total charge: %v, but got before key:%s", lru.TotalCharge(), now_deleted_key)
				}
			}
			now_deleted ++
		}
	}
}