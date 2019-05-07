package lua

import "fmt"

type event int

const (
	_index 	event = iota
	_newindex
	_gc
	_mode
	_len
	_add
	_sub
	_mul
	_mod
	_pow
	_div
	_idiv
	_band
	_bor
	_bxor
	_shl
	_shr
	_unm
	_bnot
	_eq
	_lt
	_le
	_concat
	_call
	maxEvent
)

var events = [...]string{
	_index:    "index",
	_newindex: "newindex",
	_gc:       "gc",
	_mode:     "mode",
	_len:      "len",
	_add:      "add",
	_sub:      "sub",
	_mul:      "mul",
	_mod:      "mod",
	_pow:      "pow",
	_div:      "div",
	_idiv:     "idiv",
	_band:     "band",
	_bor:      "bor",
	_bxor:     "bxor",
	_shl:      "shl",
	_shr:      "shr",
	_unm:      "unm",
	_bnot:     "bnot",
	_eq:       "eq",
	_lt:       "lt",
	_le:       "le",
	_concat:   "concat",
	_call:     "call",
}

func (evt event) String() string { return "__" + events[evt] }

func callBinaryMeta(ls *thread, method event, x, y *Value) (ret Value, ok bool, err error) {
	if ls != nil {
		if fn := ls.meta(*x, method.String()); fn != nil { // try 1st operand
			rets, err := ls.callE(fn, []Value{*x, *y}, 1)
			if err != nil {
				return nil, true, err
			}
			return rets[0], true, nil
		}
		if fn := ls.meta(*y, method.String()); fn != nil { // try 2nd operand
			rets, err := ls.callE(fn, []Value{*x, *y}, 1)
			if err != nil {
				return nil, true, err
			}
			return rets[0], true, nil
		}
	}
	return nil, false, nil
}

func tryBinaryMeta(ls *thread, method event, x, y *Value) (Value, error) {
	v, ok, err := callBinaryMeta(ls, method, x, y)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, opError(ls, method, x, y)
	}
	return v, nil
}

func tryCompareMeta(ls *thread, method event, x, y *Value) (bool, error) {
	v, ok, err := callBinaryMeta(ls, method, x, y)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, opError(ls, method, x, y)
	}
	return Truth(v), nil
}

func tryUnaryMeta(ls *thread, method event, x *Value) (Value, error) {
	v, ok, err := callBinaryMeta(ls, method, x, x)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, opError(ls, method, x, nil)
	}
	return v, nil
}

func opError(ls *thread, op event, v1, v2 *Value) error {
	var (
		call = ls.calls
		msg string
		obj *Value
	)
	switch op {
		case _bnot, _band, _bxor, _bor, _shl, _shr:
			if AsNumber(*v1) != nil && AsNumber(*v2) != nil {
				if _, ok := AsInt(*v1); ok {
					obj = v1
				} else {
					obj = v2
				}
				return fmt.Errorf(
					"number%s has no integer representation",
					call.varinfo(obj),
				)
			}
			obj, msg = v2, "perform bitwise operation on"
			if AsNumber(*v1) == nil {
				obj = v1
			}
		case _eq, _lt, _le:
			var (
				o1 = objectTypeName(*v1)
				o2 = objectTypeName(*v2)
			)
			if o1 == o2 {
				return fmt.Errorf("attempt to compare two %s values", o1)
			}
			return fmt.Errorf("attempt to compare %s with %s", o1, o2)
		case _concat:
			if obj, msg = v1, "concatenate"; IsString(*obj) || IsNumber(*obj) {
				obj = v2
			}
		case _call:
			obj, msg = v1, "call"
		case _len:
			obj, msg = v1, "get length of"
		default:
			obj, msg = v2, "perform arithmetic on"
			if AsNumber(*v1) == nil {
				obj = v1
			}
	}
	return call.errorf("attempt to %s a %s value %s", msg, objectTypeName(*obj), call.varinfo(obj))
}

func objectTypeName(obj Value) string {
	if o, ok := obj.(hasMeta); ok {
		if m := o.getMeta(); m != nil {
			s, ok := m.Get(String("__name")).(String)
			if ok {
				return string(s)
			}
		}
	}
	return typeName(obj)
}

func typeName(v Value) string {
	if v == nil {
		return "nil"
	}
	return v.kind().String()
}