package golua

import (
	"reflect"
	"fmt"
	"github.com/Azure/golua/lua"
)

type valueError struct {
	want reflect.Type
	have lua.Value
}

func (err *valueError) Error() string {
	if err.have == nil {
		return fmt.Sprintf("cannot use nil as type %s", err.want)
	}
	if v, ok := err.have.(*lua.GoValue); ok {
		return fmt.Sprintf("cannot use %v (type '%T') as type '%s'", v.Value, v.Value, err.want)
	}
	return fmt.Sprintf("cannot use %v (type '%T') as type '%s'", err.have, err.have, err.want)
}

func Closure(any interface{}, freeVars ...interface{}) (fn *lua.GoFunc) {
	if fn = Func(any); fn != nil {
		up := make([]lua.Value, len(freeVars))
		for i := 0; i < len(freeVars); i++ {
			up[i] = Value(freeVars[i])
		}
		return lua.Closure(fn, up...).(*lua.GoFunc)
	}
	return fn
}

func Value(any interface{}) (value lua.Value) {
	if v, ok := any.(lua.Value); ok {
		return v
	}
	if any == nil {
		return nil
	}
	switch rv := reflect.ValueOf(any); rv.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
			return lua.Float(rv.Uint())
		case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
			return lua.Int(rv.Int())
		case reflect.Float32, reflect.Float64:
			return lua.Float(rv.Float())
		case reflect.String:
			return lua.String(rv.String())
		case reflect.Bool:
			return lua.Bool(rv.Bool())
		case reflect.Func:
			return Func(any)
		case reflect.Ptr, reflect.Map, reflect.Slice:
			if rv.IsNil() {
				return nil
			}
			fallthrough
		case reflect.Array, reflect.Struct:
			return &lua.GoValue{Value: rv.Interface(), Meta: typeMeta(rv)}
		default:
			return &lua.GoValue{Value: rv.Interface()}
	}
	panic("unreachable")
}

func Func(impl interface{}) (fn *lua.GoFunc) {
	if rv := reflect.ValueOf(impl); rv.Kind() == reflect.Func && !rv.IsNil() {
		if typ := rv.Type(); typ.NumIn() >= 1 && typ.In(0).String() == "*lua.Thread" {
			fn = lua.GoClosure(func(cont lua.Continuation) (rets []lua.Value) {
				if have, want := len(cont.Stack()), typ.NumIn() - 1; have != want {
					return cont.Errorf(
						"not enough arguments in call to %s (have %d, want %d)",
						cont.Frame().FuncName(),
						have,
						want,
					)
				}
				for _, ret := range rv.Call(checkArgs(cont, typ)) {
					rets = append(rets, Value(ret.Interface()))
				}
				return rets
			})
		}
	}
	return fn
}

func checkArgs(cont lua.Continuation, typ reflect.Type) (args []reflect.Value) {
	args = make([]reflect.Value, len(cont.Stack())+1)
	args[0] = reflect.ValueOf(cont.Thread())
	for i := 1; i < typ.NumIn(); i++ {
		args[i] = checkArg(cont, i-1, typ.In(i))
	}

	return args
}

func checkArg(cont lua.Continuation, arg int, target reflect.Type) reflect.Value {
	switch value := cont.Var(arg).Value().(type) {
		case *lua.GoValue:
			rv := reflect.ValueOf(value.Value)
			if !rv.Type().ConvertibleTo(target) {
				cont.Error(&lua.ArgError{
					Arg: arg,
					Err: &valueError{
						have: value,
						want: target,
					},
				})
			}
			return rv.Convert(target)
			
		case *lua.GoFunc:
			panic("GoFunc")
		// case *lua.Thread:
		// case *lua.Table:
		// case *lua.Func:

		case lua.String:
			rv := reflect.ValueOf(string(value))
			if !rv.Type().ConvertibleTo(target) {
				cont.Error(&lua.ArgError{
					Arg: arg,
					Err: &valueError{
						have: value,
						want: target,
					},
				})
			}
			return rv.Convert(target)

		case lua.Float:
			rv := reflect.ValueOf(float64(value))
			if !rv.Type().ConvertibleTo(target) {
				cont.Error(&lua.ArgError{
					Arg: arg,
					Err: &valueError{
						have: value,
						want: target,
					},
				})
			}
			return rv.Convert(target)

		case lua.Int:
			rv := reflect.ValueOf(int64(value))
			if !rv.Type().ConvertibleTo(target) {
				cont.Error(&lua.ArgError{
					Arg: arg,
					Err: &valueError{
						have: value,
						want: target,
					},
				})
			}
			return rv.Convert(target)

		case lua.Bool:
			rv := reflect.ValueOf(bool(value))
			if !rv.Type().ConvertibleTo(target) {
				cont.Error(&lua.ArgError{
					Arg: arg,
					Err: &valueError{
						have: value,
						want: target,
					},
				})
			}
			return rv.Convert(target)
	}
	panic("unreachable")
}