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
	"reflect"
	"unsafe"
)

type DeleteCallback func(key []byte, entry interface{})
type MergeOperator func(old_entry, new_entry interface{}) interface{}
type ChargeOperator func(entry interface{}, old_charge, new_charge uint64) uint64
type TravelEntryOperator func(key []byte, entry interface{})

type LRUCache interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	Delete(key string)
	Prune()
	TotalCharge() uint64

	Insert(key[]byte, entry interface{}, charge uint64, deleter DeleteCallback) error
	Lookup(key []byte) (interface{}, bool)
	Remove(key []byte) (interface{}, bool)
	Merge(key []byte, entry interface{}, charge uint64,  merge_opt MergeOperator, charge_opt ChargeOperator) (old_entry interface{})
	ApplyToAllCacheEntries(TravelEntryOperator)
}

type RefCache interface {
	Insert(key[]byte, entry interface{}, charge uint64, deleter DeleteCallback) error
	Lookup(key []byte) (interface{}, bool)
	Reference(key []byte) (interface{}, bool)
	Release(key []byte)

	NewId() uint64
	TotalCharge() uint64
	ApplyToAllCacheEntries(TravelEntryOperator)
}

var IntMergeOperator MergeOperator = func(old_entry, new_entry interface{}) interface{} {
	if old_entry == nil {
		return new_entry
	}

	old, ok_old := old_entry.(int)
	new, ok_new := new_entry.(int)
	if !ok_old || !ok_new {
		panic("error of merge type, old:" + reflect.TypeOf(old_entry).Name() +
			"  new:" + reflect.TypeOf(new_entry).Name())
	}
	res := old + new
	return res
}

var IntChargeOperator ChargeOperator = func(entry interface{}, old_charge, new_charge uint64) uint64 {
	var a int
	return uint64(unsafe.Sizeof(a))
}

var Int64MergeOperator MergeOperator = func(old_entry, new_entry interface{}) interface{} {
	if old_entry == nil {
		return new_entry
	}

	old, ok_old := old_entry.(int64)
	new, ok_new := new_entry.(int64)
	if !ok_old || !ok_new {
		panic("error of merge type, old:" + reflect.TypeOf(old_entry).Name() +
			"  new:" + reflect.TypeOf(new_entry).Name())
	}
	res := old + new
	return res
}

var Int64ChargeOperator ChargeOperator = func(entry interface{}, old_charge, new_charge uint64) uint64 {
	var a int64
	return uint64(unsafe.Sizeof(a))
}


