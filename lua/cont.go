package lua

import (
	"fmt"
)

// TODO: comment
type Continuation interface {
	OptCallableVar(arg int, opt Callable) Callable
	OptConstantVar(arg int, opt Constant) Constant
	OptGoValueVar(arg int, opt *GoValue) *GoValue
	OptGoFuncVar(arg int, opt *GoFunc) *GoFunc
	OptTableVar(arg int, opt *Table) *Table
	OptNumberVar(arg int, opt Number) Number
	OptStringVar(arg int, opt String) String
	OptFloatVar(arg int, opt Float) Float
	OptBoolVar(arg int, opt Bool) Bool
	OptAnyVar(arg int, opt Value) Value
	OptIntVar(arg int, opt Int) Int

	CallableVar(arg int) Callable
	ConstantVar(arg int) Constant
	GoValueVar(arg int) *GoValue
	GoFuncVar(arg int) *GoFunc
	ThreadVar(arg int) *Thread
	TableVar(arg int) *Table
	FloatVar(arg int) Float
	StringVar(arg int) String
	NumberVar(arg int) Number
	BoolVar(arg int) Bool
	AnyVar(arg int) Value
	IntVar(arg int) Int
	Var(arg int) Object

	Errorf(string, ...interface{}) []Value
	Error(error) []Value
	Panic(Value) []Value

	Caller(int) Frame
	Thread() *Thread
	Frame() Frame
	Stack() []Value

	Recover(func(*Error)error)
}


// TODO: comment
func (ls *thread) Recover(catch func(*Error) error) {
	if r := recover(); r != nil {
		if e, ok := r.(*Error); ok {
			ls.calls.err = catch(e)
		}
	}	
}

// TODO: comment
func (ls *thread) Caller(level int) Frame { return frame{ls.calls.unwind(level)} }

// TODO: comment
func (ls *thread) Frame() Frame { return ls.Caller(0) }

// TODO: comment
func (ls *thread) Thread() *Thread { return ls.thread }

//
// Accessors with defaults
//

// TODO: comment
func (ls *thread) OptCallableVar(arg int, opt Callable) Callable {
	v, err := ls.Var(arg).Callable()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptConstantVar(arg int, opt Constant) Constant {
	v, err := ls.Var(arg).Constant()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptGoValueVar(arg int, opt *GoValue) *GoValue {
	v, err := ls.Var(arg).GoValue()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptGoFuncVar(arg int, opt *GoFunc) *GoFunc {
	v, err := ls.Var(arg).GoFunc()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptTableVar(arg int, opt *Table) *Table {
	v, err := ls.Var(arg).Table()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptNumberVar(arg int, opt Number) Number {
	v, err := ls.Var(arg).Number()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptStringVar(arg int, opt String) String {
	v, err := ls.Var(arg).String()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptFloatVar(arg int, opt Float) Float {
	v, err := ls.Var(arg).Float()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptBoolVar(arg int, opt Bool) Bool {
	v, err := ls.Var(arg).Bool()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptAnyVar(arg int, opt Value) Value {
	v, err := ls.Var(arg).Any()
	if err != nil {
		return opt
	}
	return v
}

// TODO: comment
func (ls *thread) OptIntVar(arg int, opt Int) Int {
	v, err := ls.Var(arg).Int()
	if err != nil {
		return opt
	}
	return v
}

//
// Accessors
//

// TODO: comment
func (ls *thread) CallableVar(arg int) Callable {
	v, err := ls.Var(arg).Callable()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) ConstantVar(arg int) Constant {
	v, err := ls.Var(arg).Constant()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) GoValueVar(arg int) *GoValue {
	v, err := ls.Var(arg).GoValue()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) GoFuncVar(arg int) *GoFunc {
	v, err := ls.Var(arg).GoFunc()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) ThreadVar(arg int) *Thread {
	v, err := ls.Var(arg).Thread()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) TableVar(arg int) *Table {
	v, err := ls.Var(arg).Table()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) StringVar(arg int) String {
	v, err := ls.Var(arg).String()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) NumberVar(arg int) Number {
	v, err := ls.Var(arg).Number()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) FloatVar(arg int) Float {
	v, err := ls.Var(arg).Float()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) BoolVar(arg int) Bool {
	v, err := ls.Var(arg).Bool()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) AnyVar(arg int) Value {
	v, err := ls.Var(arg).Any()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) IntVar(arg int) Int {
	v, err := ls.Var(arg).Int()
	if err != nil {
		ls.Error(&ArgError{
			Arg: arg,
			Err: err,
		})
	}
	return v
}

// TODO: comment
func (ls *thread) Var(arg int) Object {
	return ls.object(arg)
}

// TODO: comment
func (ls *thread) Stack() []Value {
	return ls.stack[ls.calls.base:ls.calls.top]
}

//
// Errors
//

// TODO: comment
func (ls *thread) Errorf(format string, args ...interface{}) []Value {
	return ls.Error(errorString(fmt.Sprintf(format, args...)))
}

// TODO: comment
func (ls *thread) Error(err error) []Value {
	if causer, ok := err.(Causer); ok {
		err = causer.Cause(frame{ls.calls})
	} else {
		return ls.Error(errorString(err.Error())) 
	}
	return ls.Panic(String(err.Error()))
}

// TODO: comment
func (ls *thread) Panic(value Value) []Value {
	panic(&Error{value: value, frame: frame{ls.calls}})
}

//
// Stack
//

// TODO: comment
func (ls *thread) object(i int) *object {
	if i += ls.calls.base; i >= ls.calls.base && i < ls.calls.top {
		var (
			fr = frame{ls.calls}
			vv = ls.stack[i]
		)
		return &object{
			value: &vv,
			frame: fr,
			index: i,
		}
	}
	return nil
}

// TODO: comment
func (ls *thread) insert(i int, v Value) {
	ls.push(nil)
	copy(ls.stack[i+1:], ls.stack[i:])
	ls.stack[i] = v
}

// TODO: comment
func (ls *thread) check(n int) *thread {
	if room := len(ls.stack) - ls.top; room < n {
		ls.stack = append(ls.stack[:ls.top], make([]Value, n)...)
	}
	return ls
}

// TODO: comment
func (ls *thread) push(values ...Value) *thread {
	ls.stack = append(ls.stack[:ls.top], values...)
	ls.top += len(values)
	return ls
}

// TODO: comment
func (ls *thread) do(ci *call) {
	ls.calls, ci.ls = ci, ls
	ci.fn.call(ls)
}