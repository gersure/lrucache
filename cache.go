package lrucache

type DeleteCallback func(key []byte, entry interface{})
type MergeOperator func(old_entry, new_entry interface{}) interface{}
type ChargeOperator func(entry interface{}, old_charge, new_charge uint64) uint64

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
	Merge(key []byte, entry interface{}, charge uint64,  merge_opt MergeOperator, charge_opt ChargeOperator)
}
