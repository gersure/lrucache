package lrucache

import (
	"reflect"
	"unsafe"
)

type DeleteCallback func(key []byte, entry interface{})
type MergeOperator func(old_entry, new_entry interface{}) interface{}
type ChargeOperator func(entry interface{}, old_charge, new_charge uint64) uint64
type TravelEntryOperator func(key []byte, entry interface{})

type Cache interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	Delete(key string)
	NewId() uint64
	Prune()
	TotalCharge() uint64

	Insert(key[]byte, entry interface{}, charge uint64, deleter DeleteCallback)
	Lookup(key []byte) interface{}
	Remove(key []byte) interface{}
	Merge(key []byte, entry interface{}, charge uint64,  merge_opt MergeOperator, charge_opt ChargeOperator) (old_entry interface{})
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


