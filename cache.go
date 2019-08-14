package lrucache

type deleter_callback func(key []byte, entry interface{})
type merge_operator func(old_entry, new_entry interface{}) interface{}
type charge_operator func(entry interface{}, old_charge, new_charge uint64) uint64

type Cache interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	Delete(key string)
	NewId() uint64
	Prune()
	TotalCharge() uint64

	Insert(key[]byte, entry interface{}, charge uint64, deleter deleter_callback)
	Lookup(key []byte) interface{}
	Remove(key []byte) interface{}
	Merge(key []byte, entry interface{}, charge uint64,  merge_opt merge_operator, charge_opt charge_operator)
}
