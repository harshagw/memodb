package core

var store map[string]*Obj

type Obj struct {
	Value interface{}
}

func init() {
	store = make(map[string]*Obj)
}

func newObj(v interface{}) *Obj {
	return &Obj{Value: v}
}

func set(k string, obj *Obj) {
	store[k] = obj
}

func get(k string) *Obj {
	return store[k]
}

func del(k string) bool {
	_, ok := store[k]
	if !ok {
		return false
	}

	delete(store, k)
	return true
}
