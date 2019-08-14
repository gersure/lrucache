package lrucache

import "github.com/spaolacci/murmur3"

func HashSlice(key []byte) uint32 {
	return murmur3.Sum32(key)
}

