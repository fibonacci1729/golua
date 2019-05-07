package lua

import "fmt"

var _ = fmt.Println

type (
	Meta interface {
		SetIndex(ls *Thread, k, v Value) (Value, error)
		Index(ls *Thread, k Value) (Value, error)
		Binary(ls *Thread, op Op, y Value) (Value, error)
		Unary(ls *Thread, op Op) (Value, error)
		Equal(ls *Thread, y Value) (Bool, error)
		Less(ls *Thread, y Value) (Value, error)
		Len(ls *Thread) (Int, error)
	}

	Type interface {
		// luaL_callmeta
		Invoke(ls *Thread, method string, args ...Value) (Value, bool, error)

		// Field returns the field from the metatable of the object.
		//
		// If the object does not have a metatable, or if the metatable
		// does not have this field, returns nil.
		//
		// luaL_getmetafield
		Field(name string) Value
		Meta() *Table
		Name() string
		CanSet() bool
	}

	meta struct {
		mt *Table
		tv Value
	}
)

// TypeOf returns the runtime type information for the Value obj.
func TypeOf(thread *Thread, obj Value) Type { return thread.typeOf(obj) }

func (t *meta) Invoke(ls *Thread, method string, args ...Value) (Value, bool, error) {
	if fn := t.Field(method); fn != nil {
		args = append([]Value{t.tv}, args...)
		rets, err := ls.CallN(fn, args, 1)
		if err != nil {
			return nil, true, err
		}
		return rets[0], true, nil
	}
	return nil, false, nil
}

func (t *meta) Field(name string) Value {
	if t.tv != nil && t.mt != nil {
		return t.mt.Get(String(name))	
	}
	return nil
}

func (t *meta) Meta() *Table {
	if t.tv != nil {
		return t.mt
	}
	return nil
}

func (t *meta) Name() string {
	if t.tv == nil {
		return "nil"
	}
	if mt := t.mt; mt != nil {
		name, ok := t.mt.Get(String("__name")).(String)
		if ok {
			return string(name)
		}
	}
	return t.tv.Kind()
}

func (t *meta) CanSet() bool {
	if t.tv != nil && t.mt != nil {
		return t.mt.Get(String("__metatable")) == nil
	}
	return false
}