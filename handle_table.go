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
"bytes"
)


type LRUHandle struct {
	entry     interface{}
	deleter   DeleteCallback
	next_hash *LRUHandle
	next      *LRUHandle;
	prev      *LRUHandle;
	charge    uint64; // TODO(opt): Only allow uint32_t?
	hash      uint32; // Hash of key(); used for fast sharding and comparisons
	key  []byte; // Beginning of key
}


type HandleTable struct {
	list   []*LRUHandle
	lenght uint32
	elems  uint32
}

func NewLRUHandleTable() *HandleTable {
	table := &HandleTable{
		lenght: 0,
		elems:  0,
	};
	table.Resize()

	return table
}

func (this *HandleTable) Lookup(key []byte, hash uint32) *LRUHandle {
	return *this.findPointer(key, hash)
}

/**
	when not find return nil;
	else replace handl and return old handle
 */
func (this *HandleTable) Insert(e *LRUHandle) *LRUHandle {
	pptr := this.findPointer(e.key, e.hash)
	old := *pptr
	if (old == nil) {
		e.next_hash = nil
	} else {
		e.next_hash = old.next_hash
	}

	*pptr = e
	if (old == nil) {
		(this.elems)++

		if (this.elems > this.lenght) {
			this.Resize()
		}
	}

	return old
}


func (this *HandleTable) Remove(key []byte, hash uint32) *LRUHandle {
	pptr := this.findPointer(key, hash)
	result := *pptr;
	if (result != nil) {
		*pptr = result.next_hash
		this.elems--
	}
	return result
}



func (this *HandleTable) Resize() {

	var new_length uint32 = 16;
	for ; new_length < (uint32(float32(this.elems) * (1.5))); {
		new_length *= 2
	}

	var new_list = make([]*LRUHandle, new_length)
	var count uint32 = 0
	for i := uint32(0); i < this.lenght; i++ {
		h := this.list[i]
		for (h) != nil {
			next := (h).next_hash
			hash := (h).hash
			pptr := &new_list[hash&(new_length-1)]
			(h).next_hash = *pptr
			*pptr = h
			h = next
			count++
		}
	}
	if this.elems != count {
		panic("LRUHandle elems == count")
	}

	this.list = new_list[:]
	this.lenght = uint32(new_length)

}

func (this *HandleTable) ApplyToAllCacheEntries(travel_fun TravelEntryOperator) {
	for i := uint32(0); i < this.lenght; i++ {
		h := this.list[i];
		for h != nil {
			n := h.next_hash;
			travel_fun(h.key, h.entry);
			h = n;
		}
	}
}



func (this *HandleTable) findPointer(key []byte, hash uint32) **LRUHandle {
	ptr := &this.list[hash&(this.lenght-1)]
	for ; *ptr != nil &&
		((*ptr).hash != hash || bytes.Compare(key, (*ptr).key) != 0); {
		ptr = &(*ptr).next_hash
	}
	return ptr
}