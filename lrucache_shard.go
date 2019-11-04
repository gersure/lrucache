/**
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lrucache

import (
	"errors"
	"sync"
)

type LRUCacheShard struct {
	capacity   uint64
	mutex      sync.Mutex
	usage      uint64    // usage of memory
	lrulist    LRUHandle // head of lru list;    lru.prev is newest entry, lru.next is oldest entry
	table      HandleTable
	handlePool sync.Pool
}

func NewLRUCacheShard(capacity uint64) *LRUCacheShard {
	lru_shared := &LRUCacheShard{
		capacity: 0,
		usage:    0,
		table:    *NewLRUHandleTable(),
		handlePool: sync.Pool{
			New: func() interface{} {
				return new(LRUHandle)
			},
		},
	}

	lru_shared.lrulist.next = &(lru_shared.lrulist)
	lru_shared.lrulist.prev = &(lru_shared.lrulist)
	lru_shared.SetCapacity(capacity)

	return lru_shared
}

/**
create lruhandle and Insert to cache,
*/
func (this *LRUCacheShard) Insert(key []byte, hash uint32, entry interface{}, charge uint64, deleter DeleteCallback) error {

	// If the cache is full, we'll have to release it
	// It shouldn't happen very often though.
	this.mutex.Lock();
	defer this.mutex.Unlock()
	return this.insert(key, hash, entry, charge, deleter)
}

/**
find key's lruhandle, return nil if not find;
*/
func (this *LRUCacheShard) Lookup(key []byte, hash uint32) (interface{}, bool) {
	this.mutex.Lock();
	defer this.mutex.Unlock()
	e := this.handle_lookup_update(key, hash);
	if e != nil {
		return e.entry, true
	}
	return nil, false
}


/**
manual reomve key
if find return value and true
if not find return nil and false
*/
func (this *LRUCacheShard) Remove(key []byte, hash uint32) (interface{}, bool) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	return this.lruRemove(key, hash)
}

func (this *LRUCacheShard) Reference(key []byte, hash uint32) (interface{}, bool) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	e := this.handle_lookup(key, hash)
	if e != nil {
		e.Reference()
		return e, true
	}
	return nil, false
}

func (this *LRUCacheShard) Release(key []byte, hash uint32) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	e := this.handle_lookup(key, hash)
	if e != nil {
		e.Release()
	}
}

func (this *LRUCacheShard) Merge(key []byte, hash uint32, entry interface{}, charge uint64, merge MergeOperator, charge_opt ChargeOperator) (interface{}, error) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	e := this.handle_lookup_update(key, hash)
	var new_value interface{}
	var new_charge uint64
	var res interface{}
	var deleter DeleteCallback = nil
	if e != nil {
		res = e.entry
		deleter = e.deleter
		new_value = merge(e.entry, entry)
		new_charge = charge_opt(entry, e.charge, charge)
	} else {
		res = nil
		new_value = merge(nil, entry)
		new_charge = charge_opt(entry, 0, charge)
	}
	err := this.insert(key, hash, new_value, new_charge, deleter)
	return res, err
}

func (this *LRUCacheShard) ApplyToAllCacheEntries(travel_fun TravelEntryOperator) {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	this.table.ApplyToAllCacheEntries(travel_fun)
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
	this.EvictLRU()
}

func (this *LRUCacheShard) TotalCharge() uint64 {
	this.mutex.Lock();
	defer this.mutex.Unlock();
	return this.usage;
}

/*********** lru method *************/

func (this *LRUCacheShard) insert(key []byte, hash uint32, entry interface{}, charge uint64, deleter DeleteCallback) error {
	var err error
	e := this.handlePool.Get()
	handle := e.(*LRUHandle)
	handle.entry = entry
	handle.deleter = deleter
	handle.charge = charge
	handle.hash = hash
	handle.key = key

	// if capacity == 0; will turn off caching
	if this.capacity > 0 {
		err = this.lru_insert(handle, charge)
	} else {
		err = errors.New("cache is turn off")
	}

	this.EvictLRU()

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

func (this *LRUCacheShard) EvictLRU() {
	for this.usage > this.capacity && this.lrulist.next != &this.lrulist {
		old := this.lrulist.next
		this.lru_remove_handle(old, true)
	}
}

/*********** lru method *************/

func (this *LRUCacheShard) lruRemove(key []byte, hash uint32) (interface{}, bool) {
	e := this.handle_lookup(key, hash);
	if e != nil {
		if !this.lru_remove_handle(e, true) {
			return e.entry, false
		}
		return e.entry, true
	}
	return nil, false
}

/**
lru Remove; if table Insert return's handle, it's aready removed from table,
so also_table is flase
*/
func (this *LRUCacheShard) lru_remove_handle(e *LRUHandle, also_table bool) bool {

	var remove_succ bool = false
	if also_table {
		// will remove fail if handle's ref > 1
		remove_succ, _ = this.table.Remove(e.key, e.hash)
	}

	if also_table && !remove_succ {
		return false
	} else {
		this.list_remove(e)
		if (e.deleter != nil) {
			e.deleter(e.key, e.entry)
		}
		this.usage -= e.charge;
		this.handlePool.Put(e)
		return true
	}
}

func (this *LRUCacheShard) lru_insert(e *LRUHandle, charge uint64) error {
	this.list_append(e)
	this.usage += charge
	old, err := this.table.Insert(e)
	if err != nil {
		return err
	}
	if old != nil {
		//don't need table.Remove; it's aready removed
		if !this.lru_remove_handle(old, false) {
			panic("unarrive")
		}
	}
	return nil
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
