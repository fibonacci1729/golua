package lua

import (
	"reflect"
	"math"
	"fmt"
)

// Compare returns the result of 'x op y' or an error.
func Compare(ls *Thread, op Op, x, y Value) (bool, error) {
	return compare(ls.thread, op, &x, &y)
}

// Equals returns the result of 'x == y' or an error.
func Equals(ls *Thread, x, y Value) (bool, error) {
	return compare(ls.thread, OpEq, &x, &y)
}

// Length returns the length of 'x' or an error.
func Length(ls *Thread, x Value) (Int, error) {
	v, err := length(ls.thread, &x)
	if err != nil {
		return 0, err
	}
	i, ok := AsInt(v)
	if !ok {
		return 0, fmt.Errorf("object length is not an integer")
	}
	return i, nil
}

// Binary returns the value of 'x op y' or an error.
func Binary(ls *Thread, op Op, x, y Value) (Value, error) {
	return binary(ls.thread, op, &x, &y)
}

// Unary returns the value of 'op x' or an error.
// func Unary(ls *Thread, op Op, x Value) (Value, error) {
// 	return unary(ls.thread, op, x)
// }

// SetIndex performs 'obj[k]=v' returning an error or nil.
// func SetIndex(ls *Thread, obj, k, v Value) error {
// 	return settable(ls.thread, obj, k, v)
// }

// Index returns the result of 'obj[k]'.
// func Index(ls *Thread, obj, key Value) (Value, error) {
// 	return gettable(ls.thread, obj, key)
// }

// Equals first compares the type of its operands. If the types
// are different, then the result is false. Otherwise, the values
// of the operands are compared. Strings are compared in the obvious
// way. Numbers are equal if they denote the same mathematical value.
//
// Tables, userdata, and threads are compared by reference: two objects
// are considered equal only if they are the same object.
//
// You can change the way that Lua compares tables and userdata by using the
// "eq" metamethod (see §2.4).
//
// Equality comparisons do not convert strings to numbers or vice versa.
// Thus, "0"==0 evaluates to false,  and t[0] and t["0"] denote different
// entries in a table.
func equals(ls *thread, x, y *Value) (bool, error) {
	if !sameType(*x, *y) {
		switch x := (*x).(type) {
			case Float:
				if y, ok := (*y).(Int); ok {
					return x == Float(y), nil
				}
				y, ok := (*y).(Float)
				return ok && (x == y), nil
			case Int:
				if y, ok := (*y).(Float); ok {
					return Float(x) == y, nil
				}
				y, ok := (*y).(Int)
				return ok && (x == y), nil
			default:
				return false, nil
		}
	}
	switch x := (*x).(type) {
		case *GoValue:
			if y := (*y).(*GoValue); x == y {
				return true, nil
			} else {
				if x.Meta == nil {
					if y.Meta == nil {
						return false, nil
					}
				}
			}
		case *Table:
			if y := (*y).(*Table); x == y {
				return true, nil
			} else {
				if x.meta == nil {
					if y.meta == nil {
						return false, nil
					}
				}
			}
		case *GoFunc:
			y, ok := (*y).(*GoFunc)
			return ok && (x == y), nil
		case *Func:
			y, ok := (*y).(*Func)
			return ok && (x == y), nil
		case String:
			y, ok := (*y).(String)
			return ok && (x == y), nil
		case Float:
			y, ok := (*y).(Float)
			return ok && (x == y), nil
		case Bool:
			y, ok := (*y).(Bool)
			return ok && (x == y), nil
		case Int:
			y, ok := (*y).(Int)
			return ok && (x == y), nil
		case nil:
			return (*y == nil), nil
	}
	return tryCompareMeta(ls, _eq, x, y)
}

func binary(ls *thread, op Op, x, y *Value) (Value, error) {
	switch op {
		case OpDivF, OpPow:
			if x, ok := AsFloat(*x); ok {
				if y, ok := AsFloat(*y); ok {
					return numop(op, x, y), nil
				}
			}

		case OpBand,
			OpBor,
			OpBxor,
			OpShl, 
			OpShr:
			
			if x, ok := (*x).(Int); ok {
				if y, ok := (*y).(Int); ok {
					return intop(op, x, y), nil
				}
			}

		default:
			if x, ok := (*x).(Int); ok {
				if y, ok := (*y).(Int); ok {
					return intop(op, x, y), nil
				}
			}
			if x, ok := AsFloat(*x); ok {
				if y, ok := AsFloat(*y); ok {
					return numop(op, x, y), nil
				}
			}
	}
	return tryBinaryMeta(ls, event(op-OpAdd)+_add, x, y)
}

func gettable(ls *thread, t, k Value) (Value, error) {
	// - If 't' is a table and 't[k]' is not nil, return value.
	// - Otherwise check 't' for '__index' metamethod.
	// - If metamethod is nil, return nil.
	// - If metamethod exists and table, repeat lookuped with t = m.
	// - If metamethod exists and function, call 't.__index(t, k)'.
	for loop := 0; loop < maxMetaLoop; loop++ {
		if t, ok := t.(*Table); ok {
			if v := t.Get(k); v != nil {
				return v, nil
			}
		}
		switch m := ls.meta(t, "__index").(type) {
			case Callable:
				rets, err := ls.callE(m, []Value{t, k}, 1)
				if err != nil {
					return nil, err
				}
				return rets[0], nil
			case *Table:
				t = m
			default:
				return nil, nil
		}
	}
	return nil, fmt.Errorf("'__index' chain too long; possible loop")
}

func settable(ls *thread, obj, key, val Value) (err error) {
	// - If 't' is a table and 't[k]' is not nil, then 't[k]=v' and return nil.
	// - Otherwise check 't' for '__newindex' metamethod.
	// - If metamethod is nil, return nil.
	// - If metamethod exists and table, repeat lookup with t = m.
	// - If metamethod exists and function, call 't.__index(t, k)'.
	for loop := 0; loop < maxMetaLoop; loop++ {
		if t, ok := obj.(*Table); ok {
			if v := t.Get(key); v != nil {
				t.Set(key, val)
				return nil
			}
		}
		if m := ls.meta(obj, "__newindex"); m != nil {
			switch m := m.(type) {
				case Callable:
					_, err = ls.callE(m, []Value{obj, key, val}, 0)
					return err
				case *Table:
					obj = m
					continue
			}
		}
		if t, ok := obj.(*Table); ok {
			t.Set(key, val)
		}
		return nil
	}
	return fmt.Errorf("'__newindex' chain too long; possible loop")
}

func compare(ls *thread, op Op, x, y *Value) (bool, error) {
	switch op {
		case OpNe, OpEq:
			switch eq, err := equals(ls, x, y); {
				case err != nil:
					return false, err
				case op == OpNe:
					return !eq, nil
				default:
					return eq, nil
			}
		case OpGt:
			lt, err := less(ls, y, x)
			if err != nil {
				return false, err
			}
			return lt, nil
		case OpGe:
			le, err := lesseq(ls, y, x)
			if err != nil {
				return false, err
			}
			return le, nil
		case OpLt:
			return less(ls, x, y)
		case OpLe:
			return lesseq(ls, x, y)
	}
	return false, fmt.Errorf("unexpected comparison operator '%v'", op)
}

// func concat(ls *thread, xs *[]Value) (Value, error) {
// 	var (
// 		y = (*xs)[len(*xs)-1]
// 		i = len(*xs) - 2
// 		x Value
// 		err error
// 	)
// 	for ; i >= 0 ; i-- {
// 		if x = (*xs)[i]; IsNumber(x) || IsString(x) {
// 			if IsNumber(y) || IsString(y) {
// 				s1, _ := AsString(x)
// 				s2, _ := AsString(y)
// 				y = String(s1 + s2)
// 				continue
// 			}
// 		}
// 		y, err = tryBinaryMeta(ls, _concat, &x, &y)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return y, nil
// }

func lesseq(ls *thread, x, y *Value) (bool, error) {
	switch x := (*x).(type) {
		case String:
			if y, ok := (*y).(String); ok {
				return x <= y, nil
			}
		case Float:
			switch y := (*y).(type) {
				case Float:
					return x <= y, nil
				case Int:
					return x <= Float(y), nil
			}
		case Int:
			switch y := (*y).(type) {
				case Float:
					return Float(x) <= y, nil
				case Int:
					return x <= y, nil
			}
	}
	return tryCompareMeta(ls, _le, x, y)
}

func less(ls *thread, x, y *Value) (bool, error) {
	switch x := (*x).(type) {
		case String:
			if y, ok := (*y).(String); ok {
				return x < y, nil
			}
		case Float:
			switch y := (*y).(type) {
				case Float:
					return x < y, nil
				case Int:
					return x < Float(y), nil
			}
		case Int:
			switch y := (*y).(type) {
				case Float:
					return Float(x) < y, nil
				case Int:
					return x < y, nil
			}
	}
	return tryCompareMeta(ls, _lt, x, y)
}

func length(ls *thread, x *Value) (ret Value, err error) {
	switch x := (*x).(type) {
		case String:
			return Int(len(x)), nil
		case *Table:
			if ls == nil || x.meta == nil {
				return x.Length(), nil
			}		
	}
	if ls != nil {
		if m := ls.meta(*x, "__len"); m != nil {
			rets, err := ls.callE(m, []Value{*x}, 1)
			if err != nil {
				return nil, err
			}
			return rets[0], err
		}
	}
	return nil, fmt.Errorf("attempt to get length of a %s value", objectTypeName(*x))
}

// UNM, BNOT, NOT, LEN
func unary(ls *thread, op Op, x *Value) (Value, error) {
	switch op {
		case OpMinus:
			if x, ok := (*x).(Int); ok {
				return intop(op, x, Int(0)), nil
			}
			if x, ok := AsFloat(*x); ok {
				return numop(op, x, Float(0)), nil
			}
			return tryUnaryMeta(ls, _unm, x)
		case OpBnot:
			if x, ok := AsInt(*x); ok {
				return intop(op, x, Int(0)), nil
			}
			return tryUnaryMeta(ls, _bnot, x)
		case OpNot:
			return Bool(!Truth(*x)), nil
		case OpLen:
			return length(ls, x)
	}
	return nil, fmt.Errorf("unexpected unary operator '%v'", op)
}

func numop(op Op, x, y Float) Float {
	switch op {
		case OpMinus:
			return -x
		case OpDivF:
			return x / y
		case OpDivI:
			return Float(math.Floor(float64(x/y)))
		case OpAdd:
			return x + y
		case OpSub:
			return x - y
		case OpMul:
			return x * y
		case OpPow:
			f64 := math.Pow(float64(x), float64(y))
			return Float(f64)
		case OpMod:
			f64 := Float(math.Mod(float64(x), float64(y)))
			if f64 * y < 0 {
				f64 += y
			}
			return f64
	}
	panic(fmt.Errorf("unexpected binary operator '%v'", op))
}

func intop(op Op, x, y Int) Int {
	switch op {
		case OpMinus:
			return -x
		case OpDivI:
			return x / y
		case OpBand:
			return x & y
		case OpBnot:
			return ^x 
		case OpBxor:
			return x ^ y
		case OpBor:
			return x | y
		case OpAdd:
			return x + y
		case OpSub:
			return x - y
		case OpMul:
			return x * y
		case OpMod:
			if r := (x % y); r != 0 && (x ^ y) < 0 { // 'm/n' would be non-integer negative?
				r += y // correct result for different rounding
				return Int(r)
			} else {
				return Int(r)
			}
		case OpShl:
			return shiftLeft(x, y)
		case OpShr:
			return shiftRight(x, y)
	}
	panic(fmt.Errorf("unexpected binary operator '%v'", op))
}

// shift left operation
func shiftLeft(x, y Int) Int {
	if y >= 0 {
		return x << uint64(y)
	}
	return shiftRight(x, -y)
}

// shift right operation
func shiftRight(x, y Int) Int {
	if y >= 0 {
		return Int(uint64(x) >> uint64(y))
	}
	return shiftLeft(x, -y)
}

func sameType(x, y Value) bool {
	if x == nil || y == nil {
		return x == y
	}
	return reflect.TypeOf(x) == reflect.TypeOf(y) || x.kind() == y.kind()
}