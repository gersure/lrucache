package lrucache

import (
	"bytes"
	"testing"
)


var test_case = []struct{
	key, value []byte
} {
	{[]byte("key0"),[]byte("value0")},
	{[]byte("key1"),[]byte("value1")},
	{[]byte("key2"),[]byte("value2")},
	{[]byte("key3"),[]byte("value3")},
	{[]byte("key4"),[]byte("value4")},
	{[]byte("key5"),[]byte("value5")},
	{[]byte("key6"),[]byte("value6")},
	{[]byte("key7"),[]byte("value7")},
	{[]byte("key8"),[]byte("value8")},
	{[]byte("key9"),[]byte("value9")},
	{[]byte("key10"),[]byte("value10")},
	{[]byte("key11"),[]byte("value11")},
	{[]byte("key12"),[]byte("value12")},
	{[]byte("key13"),[]byte("value13")},
	{[]byte("key14"),[]byte("value14")},
	{[]byte("key15"),[]byte("value15")},
	{[]byte("key16"),[]byte("value16")},
	{[]byte("key17"),[]byte("value17")},
	{[]byte("key18"),[]byte("value18")},
	{[]byte("key19"),[]byte("value19")},
}



func initLRUCache() *LRUCache {
	lru := NewLRUCache(1024*1024, 5)
	return lru
}

func TestLRUCache_Insert(t *testing.T) {
	lru := NewLRUCache(1024*1024, 5)
	for _, cases := range test_case {
		lru.Insert(cases.key, cases.value, uint64(len(cases.key)+len(cases.value)), nil)
	}


	for _, cases := range test_case {
		origin := lru.Value(lru.Lookup(cases.key))
		if (bytes.Compare(origin, cases.value) != 0) {
			t.Errorf("key:%s, excepted value:%s, got:%s", cases.key, cases.value, origin)
		}
	}
}