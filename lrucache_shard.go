package lrucache

import (
	"sync"
	"unsafe"
)

type LRUCacheShard struct {
	capacity uint64
	mutex    sync.Mutex
	usage    uint64
	lru      LRUHandle
	in_use   LRUHandle
	table    HandleTable
}

func NewLRUCacheShard(capacity uint64) *LRUCacheShard {
	lru_shared := &LRUCacheShard{capacity: 0,
		usage: 0,
		table: *NewLRUHandleTable(),
	}

	lru_shared.lru.next = &(lru_shared.lru)
	lru_shared.lru.prev = &(lru_shared.lru)
	lru_shared.in_use.next = &lru_shared.in_use
	lru_shared.in_use.prev = &lru_shared.in_use
	lru_shared.SetCapacity(capacity)

	return lru_shared
}

func (this *LRUCacheShard) SetCapacity(capacity uint64) {
	this.capacity = capacity
}

func (this *LRUCacheShard) Distroy() {
	if (this.in_use.next != &this.in_use) {
		panic("caller has an unreleased handle")
	}

	for e := this.lru.next; e != &this.lru; {
		next := e.next
		if (!e.in_cache) {
			panic("handle not in cache")
		}
		e.in_cache = false
		if (e.refs != 1) {
			panic(" handle refs is not 1")
		}
		this.Unref(e);
		e = next;
	}
}

func (this *LRUCacheShard) Ref(e *LRUHandle) {
	if e.refs == 1 && e.in_cache {
		this.LRU_Remove(e)
		this.LRU_Append(&this.in_use, e)
	}
	e.refs++
}

func (this *LRUCacheShard) Unref(e *LRUHandle) {
	if e.refs <= 0 {
		panic("refs <= 0")
	}
	e.refs--

	if (e.refs == 0) { // Deallocate.
		if e.in_cache {
			panic("handle still in cache ")
		}
		(e.deleter)(e.key, e.value);
	} else if (e.in_cache && e.refs == 1) {
		// No longer in use; move to lru_ list.
		this.LRU_Remove(e);
		this.LRU_Append(&this.lru, e);
	}
}

func (this *LRUCacheShard) LRU_Remove(e *LRUHandle) {
	e.next.prev = e.prev
	e.prev.next = e.next
}

func (this *LRUCacheShard) LRU_Append(list *LRUHandle, e *LRUHandle) {
	// Make "e" newest entry by inserting just before *list
	e.next = list;
	e.prev = list.prev;
	e.prev.next = e;
	e.next.prev = e;
}

func (this *LRUCacheShard) Lookup(key []byte, hash uint32) *LRUHandle {
	this.mutex.Lock();
	defer this.mutex.Unlock();;
	e := this.table.Lookup(key, hash);
	if (e != nil) {
		this.Ref(e)
	}
	return e;
}

func (this *LRUCacheShard) Release(handle *LRUHandle) {
	this.mutex.Lock();
	defer this.mutex.Unlock();;
	this.Unref((*LRUHandle)(unsafe.Pointer(handle)));
}

func (this *LRUCacheShard) Insert(key []byte, hash uint32, value []byte, charge uint64,
	deleter func(key, value []byte)) *LRUHandle {
	// Allocate the memory here outside of the mutex
	// If the cache is full, we'll have to release it
	// It shouldn't happen very often though.
	e := &LRUHandle{
		value:    value,
		deleter:  deleter,
		charge:   charge,
		hash:     hash,
		in_cache: false,
		refs:     1,
		key:      key,
	};
	if this.capacity > 0 {
		e.refs++
		e.in_cache = true
		this.LRU_Append(&this.in_use, e)
		this.usage += charge
		this.FinishErase(this.table.Insert(e))
	} else {
		e.next = nil
	}

	for this.usage > this.capacity && this.lru.next != &this.lru {
		old := this.lru.next
		if old.refs != 1 {
			panic("old refs != 1")
		}
		erase := this.FinishErase(this.table.Remove(old.key, old.hash))
		if !erase {
			panic("FinishErase")
		}
	}
	return e
}

// If e != nullptr, finish removing *e from the cache; it has already been
// removed from the hash table.  Return whether e != nullptr.
func (this *LRUCacheShard) FinishErase(e *LRUHandle) bool {
	if (e != nil) {
		if !e.in_cache {
			panic("handle not in cache")
		}
		this.LRU_Remove(e);
		e.in_cache = false;
		this.usage -= e.charge;
		this.Unref(e);
	}
	return e != nil;
}

func (this *LRUCacheShard) Erase(key []byte, hash uint32) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	this.FinishErase(this.table.Remove(key, hash))
}

func (this *LRUCacheShard) Prune() {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	for this.lru.next != &this.lru {
		e := this.lru.next;
		if e.refs != 1 {
			panic("old refs != 1")
		}
		erased := this.FinishErase(this.table.Remove(e.key, e.hash));
		if (!erased) { // to avoid unused variable when compiled NDEBUG
			panic("FinishErase")
		}
	}
}

func (this *LRUCacheShard) TotalCharge() uint64 {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	return this.usage;
}
