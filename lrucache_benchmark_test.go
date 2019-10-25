package lrucache

import (
	"math/rand"
	"testing"
	"time"
)

var lru *LRUCache

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init()  {
	InitLRUCache(1024*1024*1, 1)
	lru,_ = NewLRUCache("test")
}

func RandomCreateBytes(n int) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().Unix())
	var bytes []byte
	for i:=0; i<n; i++ {
		bytes = append(bytes, alphanum[rand.Int31n(62)])
	}
	return bytes
}

func BenchmarkLRUCache_Put(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lru.Put("aaaaaaaaaa", "aaaaaaaaaaaa")
	}
}

func BenchmarkLRUCache_PutRand(b *testing.B) {
	randbyte := RandomCreateBytes(b.N+5)
	rand.Seed(time.Now().Unix())
	for i := 0; i < b.N; i++ {
		key := (randbyte[i: i+5])
		lru.Put(string(key), "aaaaaaaaaaaaaaa")
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	lru.Put("aaaaaaaaaa", "aaaaaaaaaaaa")
	for i := 0; i < b.N; i++ {
		lru.Get("aaaaaaaaaa")
	}
}

func BenchmarkLRUCache_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lru.Insert([]byte("aaaaa"), nil, 5, nil)
	}
}

func BenchmarkLRUCache_InsertRand(b *testing.B) {
	randbyte := RandomCreateBytes(b.N+5)
	rand.Seed(time.Now().Unix())
	for i := 0; i < b.N; i++ {
		key := (randbyte[i: i+5])
		lru.Insert(key, nil, 1000, nil)
	}
}