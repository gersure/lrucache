package lrucache

import (
	"errors"
	"sync"
)

type LRUCacheShard struct {
	capacity uint64
	mutex    sync.Mutex
	usage    uint64    // usage of memory
	lrulist  LRUHandle // head of lru list;    lru.prev is newest entry, lru.next is oldest entry
	table    HandleTable
}

func NewLRUCacheShard(capacity uint64) *LRUCacheShard {
	lru_shared := &LRUCacheShard{
		capacity: 0,
		usage:    0,
		table:    *NewLRUHandleTable(),
	}

	lru_shared.lrulist.next = &(lru_shared.lrulist)
	lru_shared.lrulist.prev = &(lru_shared.lrulist)
	lru_shared.SetCapacity(capacity)

	return lru_shared
}

/**
create lruhandle and Insert to cache,
*/
func (this *LRUCacheShard) Insert(key []byte, hash uint32, entry interface{}, charge uint64, deleter deleter_callback) error {

	// If the cache is full, we'll have to release it
	// It shouldn't happen very often though.
	this.mutex.Lock();
	defer this.mutex.Unlock()
	return this.insert(key, hash, entry, charge, deleter)
}

/**
find key's lruhandle, return nil if not find;
*/
func (this *LRUCacheShard) Lookup(key []byte, hash uint32) interface{} {
	this.mutex.Lock();
	defer this.mutex.Unlock()
	e := this.handle_lookup_update(key, hash);
	if e != nil {
		return e.entry
	}
	return nil;
}

func (this *LRUCacheShard) Merge(key []byte, hash uint32, entry interface{}, charge uint64, merge merge_operator, charge_opt charge_operator) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	e := this.handle_lookup_update(key, hash)
	var new_value interface{}
	var new_charge uint64
	if e != nil {
		new_value = merge(e.entry, entry)
		new_charge = charge_opt(entry, 0, charge)
	}else{
		new_value = merge(nil, entry)
		new_charge = charge_opt(entry, e.charge, charge)
	}
	this.insert(key, hash, new_value, new_charge, e.deleter)
}

func (this *LRUCacheShard) Remove(key []byte, hash uint32) interface{} {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	return this.lru_remove(key, hash)
}

func (this *LRUCacheShard) Prune() {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	for this.lrulist.next != &this.lrulist {
		e := this.lrulist.next;
		this.lru_remove_handle(e, true)
	}
}

func (this *LRUCacheShard) SetCapacity(capacity uint64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.capacity = capacity

}

func (this *LRUCacheShard) TotalCharge() uint64 {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	return this.usage;
}


/*********** lru method *************/

func (this *LRUCacheShard) insert(key []byte, hash uint32, entry interface{}, charge uint64, deleter deleter_callback) error {
	var err error
	e := &LRUHandle{
		entry:   entry,
		deleter: deleter,
		charge:  charge,
		hash:    hash,
		key:     key,
	};

	// if capacity == 0; will turn off caching
	if this.capacity > 0 {
		this.lru_insert(e, charge)
	} else {
		err = errors.New("cache is turn off")
	}

	for this.usage > this.capacity && this.lrulist.next != &this.lrulist {
		old := this.lrulist.next
		this.lru_remove_handle(old, true)
	}

	return err
}

func (this *LRUCacheShard) handle_lookup(key []byte, hash uint32) *LRUHandle {
	e := this.table.Lookup(key, hash);
	return e;
}

func (this *LRUCacheShard) handle_lookup_update(key []byte, hash uint32) *LRUHandle {
	e := this.table.Lookup(key, hash);
	if (e != nil) {
		this.list_update(e)
	}
	return e;
}

/*********** lru method *************/

func (this *LRUCacheShard) lru_remove(key []byte, hash uint32) interface{} {
	e := this.handle_lookup(key, hash);
	if e != nil {
		this.lru_remove_handle(e, true)
		return e.entry
	}
	return nil
}

/**
lru Remove; if table Insert return's handle, it's aready removed from table,
so also_table is flase
*/
func (this *LRUCacheShard) lru_remove_handle(e *LRUHandle, also_table bool) {
	if also_table {
		this.table.Remove(e.key, e.hash)
	}
	this.list_remove(e)
	if (e.deleter != nil) {
		e.deleter(e.key, e.entry)
	}
	this.usage -= e.charge;
}

func (this *LRUCacheShard) lru_insert(e *LRUHandle, charge uint64) {
	this.list_append(e)
	this.usage += charge
	old := this.table.Insert(e)
	if old != nil {
		//don't need table.Remove; it's aready removed
		this.lru_remove_handle(old, false)
	}
}

/*********** lru list method *************/

func (this *LRUCacheShard) list_remove(e *LRUHandle) {
	e.next.prev = e.prev
	e.prev.next = e.next
}

/*
	Insert before list
*/
func (this *LRUCacheShard) list_append(e *LRUHandle) {
	list := &this.lrulist
	e.next = list;
	e.prev = list.prev;
	e.prev.next = e;
	e.next.prev = e;
}

func (this *LRUCacheShard) list_update(e *LRUHandle) {
	this.list_remove(e)
	this.list_append(e)
}
