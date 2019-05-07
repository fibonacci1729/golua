package lua

import (
	"strconv"
	"fmt"
)

//
// Conversions
//

// AsCallable converts the Value to a Callable value
// (either Lua or Go function).
//
// On success, returns the function; otherwise nil.
func AsCallable(v Value) Callable {
	if fn, ok := v.(Callable); ok {
		return fn
	}
	return nil
}

// AsConstant converts the Value to a Constant value.
//
// On success, returns the constant; otherwise nil.
func AsConstant(v Value) Constant {
	if c, ok := v.(Constant); ok {
		return c
	}
	return nil
}

// AsGoValue converts the Value to a *GoValue.
//
// On success, returns the *GoValue; otherwise nil.
func AsGoValue(v Value) *GoValue {
	if u, ok := v.(*GoValue); ok {
		return u
	}
	return nil
}

// AsGoFunc converts the Value to a *GoFunc.
//
// On success, returns the *GoFunc; otherwise nil.
func AsGoFunc(v Value) *GoFunc {
	if fn, ok := v.(*GoFunc); ok {
		return fn
	}
	return nil
}

// AsThread converts the Value to a *Thread.
//
// On success, returns the *Thread; otherwise nil.
func AsThread(v Value) *Thread {
	if ls, ok := v.(*Thread); ok {
		return ls
	}
	return nil
}

// AsTable converts the Value to a *Table.
//
// On success, returns the *Table; otherwise nil.
func AsTable(v Value) *Table {
	if tbl, ok := v.(*Table); ok {
		return tbl
	}
	return nil
}

// AsString converts the Value to a String.
//
// On success, returns the String and true;
// otherwise "" and false.
func AsString(v Value) (s String, ok bool) {
	switch v := v.(type) {
		case String:
			s, ok = v, true
		case Float:
			s, ok = String(strconv.FormatFloat(float64(v), 'g', -1, 64)), true
		case Int:
			s, ok = String(fmt.Sprintf("%d", v)), true
	}
	return s, ok
}

// AsNumber converts the Value to a Number.
//
// On success, returns the Number; otherwise nil.
func AsNumber(v Value) Number {
	if n, ok := v.(Number); ok {
		return n
	}
	if n, ok := AsInt(v); ok {
		return n
	}
	if n, ok := AsFloat(v); ok {
		return n
	}
	return nil
}

// AsFloat converts the Value to a Float.
//
// On success, returns the Float and true;
// otherwise 0.0 and false.
func AsFloat(v Value) (Float, bool) {
	switch v := v.(type) {
		case String:
			f, ok := str2float(string(v))
			return Float(f), ok
		case Float:
			return v, true
		case Int:
			return Float(v), true
	}
	return 0, false
}

// AsInt converts the Value to a Int.
//
// On success, returns the Int and true;
// otherwise 0 and false.
func AsInt(v Value) (Int, bool) {
	switch v := v.(type) {
		case String:
			i, ok := str2int(string(v))
			return Int(i), ok
		case Float:
			i, ok := float2int(float64(v))
			return Int(i), ok
		case Int:
			return v, true
	}
	return 0, false
}

// Truth converts the Value to a Go bool value.
//
// Returns true for any Lua value different from
// false and nil; otherwise it returns false.
func Truth(v Value) bool {
	b, ok := v.(Bool)
	return v != nil && (!ok || bool(b))
}

//
// Predicates
//

// IsFunction reports whether the Value implements
// the Function interface (i.e. *Func or *GoFunc).
func IsFunction(v Value) bool {
	_, ok := v.(Callable)
	return ok
}

// IsGoValue reports whether the Value is a *GoValue.
func IsGoValue(v Value) bool {
	_, ok := v.(*GoValue)
	return ok
}

// IsGoFunc reports whether the Value is a *GoFunc.
func IsGoFunc(v Value) bool {
	_, ok := v.(*GoFunc)
	return ok
}

// IsThread reports whether the Value is a *Thread.
func IsThread(v Value) bool {
	_, ok := v.(*Thread)
	return ok
}

// IsTable reports whether the Value is a *Table.
func IsTable(v Value) bool {
	_, ok := v.(*Table)
	return ok
}

// IsString reports whether the Value is a String or a
// Number (which is always convertible to a string).
func IsString(v Value) bool {
	_, ok := v.(String)
	return ok || IsNumber(v)
}

// IsNumber reports whether the Value is a Number or a
// String convertible to a Number.
func IsNumber(v Value) bool {
	_, ok := v.(Number)
	return ok
}

// IsFloat reports whether the Value is a Float;
// 
// That is, the value is a Number and is represented
// as a Float.
func IsFloat(v Value) bool {
	_, ok := v.(Float)
	return ok
}

// IsInt reports whether the Value is an Int;
// 
// That is, the value is a Number and is represented
// as an Int.
func IsInt(v Value) bool {
	_, ok := v.(Int)
	return ok
}

// IsBool reports whether the Value is a Bool.
func IsBool(v Value) bool {
	_, ok := v.(Bool)
	return ok
}

// IsNil reports whether the Value is nil.
func IsNil(v Value) bool {
	return v == nil
}