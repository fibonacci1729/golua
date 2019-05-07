package lua

import (
	goruntime "runtime"
	"path/filepath"
	"reflect"
	"fmt"
	"io"
	"github.com/Azure/golua/lua/code"
)

type (
	// callable is implemented by all values that are callable: *GoFunc / *Func.
	callable interface {
		cont(*thread, int, int) *call
		info(*call) *FuncInfo
		call(*thread)
	}

	// closure represents a Lua/Go closure.
	closure struct {
		fn callable
		up []*upvar
	}

	// upvar represents a Lua upvalue.
	upvar struct {
		stack  *Value
		value  Value
		index  int
		isopen bool
	}
)

// TODO: comment
func (up *upvar) close() {
	if up.isopen {
		final := up.get()
		up.isopen = false
		up.set(final)
	}
}

// set the upvalue's inner value.
func (up *upvar) set(v Value) {
	if up.isopen {
		*(up.stack) = v
		return
	}
	up.value = v
}

// get the upvalue's inner value.
func (up *upvar) get() Value {
	if up.isopen {
		return *(up.stack)
	}
	return up.value
}

// closure represents a closure value embedded into a callable values.
func (cls *closure) String() string {
	return fmt.Sprintf("function: %p", cls.fn)
}

// Up returns the closures n'th upvalue.
func (cls closure) Up(n int) Value {
	if n >= 0 && n < len(cls.up) {
		if up := cls.up[n]; up != nil {
			return up.get()
		}
	}
	return nil
}

// asClosure returns the closure for the callable value.
func asClosure(fn callable) *closure {
	switch fn := fn.(type) {
		case *GoFunc:
			return &fn.closure
		case *Func:
			return &fn.closure
	}
	return nil
}

// GoFunc represents a Go builtin function value.
type GoFunc struct {
	closure
	name string
	// impl func(*Thread, GoCont) Continuation
	impl func(Continuation) []Value
	//impl func(*Thread, []Value) []Value
}

// NewGoFunc creates and returns a new *Builtin.
func NewGoFunc(name string, impl func(Continuation) []Value) *GoFunc {
	return Closure(&GoFunc{name: name, impl: impl}).(*GoFunc)
}

// GoClosure is a convenience for creating Go functions.
// func GoClosure(impl func(*Thread, GoCont) Continuation, vars ...Value) Callable {
func GoClosure(impl func(Continuation) []Value, vars ...Value) *GoFunc {
	return Closure(NewGoFunc("", impl), vars...).(*GoFunc)
}

// TODO: comment
func Closure(fn Callable, vars ...Value) Callable {
	up := make([]*upvar, len(vars))
	for i, v := range vars {
		up[i] = &upvar{value: v}
	}
	switch fn := fn.(type) {
		case *GoFunc:
			fn.closure = closure{fn, up}
			return fn
		case *Func:
			fn.closure = closure{fn, up}
			return fn
	}
	return nil
}

// TODO: comment
func (fn *GoFunc) detail() string {
	pointer := reflect.ValueOf(fn.impl).Pointer()
	details := goruntime.FuncForPC(pointer)
	file, line := details.FileLine(details.Entry())
	extra := filepath.Join(filepath.Dir(file), details.Name())
	return fmt.Sprintf("%s:%d", extra, line)
}

// TODO: comment
func (fn *GoFunc) cont(ls *thread, fnID, retc int) *call {
	ci := &call{prev: ls.calls, ls: ls}
	ci.top  = ls.top
	ci.fn   = fn
	ci.retc = retc
	ci.fnID = fnID
	ci.base = fnID + 1
	return ci
}

// TODO: comment
func (fn *GoFunc) info(ci *call) *FuncInfo {
	var (
		name, what string
		caller = ci.prev
	)
	if caller != nil && caller.isLua() && (ci.flag & tailcall == 0) {
		if name, what = funcNameFromCode(caller); what == "" {
			name = ""
		}
	}
	return &FuncInfo{
		Source:  "=[Go]",
		Short:   "[Go]",
		Name:    name,
		What:    what,
		Kind:    "Go",
		ParamN:  0,
		UpVarN:  len(fn.up),
		LineDef: -1,
		LineEnd: -1,
		Vararg:  true,
	}
}

// TODO: comment
func (fn *GoFunc) call(ls *thread) {
	call, rets := ls.calls, fn.impl(ls)
	ls.push(rets...).returns(call, ls.top-len(rets), len(rets))
}

// A Func represents a Lua function value.
type Func struct {
	closure
	proto *code.Proto
}

// NewFunc returns a new function value for the compiled Lua chunk
// populating its first upvalue with env.
func NewFunc(chunk *code.Chunk, env *Table) *Func {
	var (
		fn = &Func{proto: chunk.Main}
		up []Value
	)
	if upN := len(fn.proto.UpVars); upN > 0 {
		up = make([]Value, upN)
		up[0] = env
	}
	return Closure(fn, up...).(*Func)
}

// CallN calls the Lua value fv with args returning want results.
//
// Returns #want results or error.
// func (fn *Func) CallN(ls *Thread, args []Value, want int) ([]Value, error) {}

// Call1 calls the Lua value fv with args returning want results.
//
// Returns 1 result or error.
// func (fn *Func) Call1(ls *Thread, args ...Value) (Value, error) {}

// Call0 calls the Lua value fv with args.
//
// Returns the error if any.
// func (fn *Func) Call0(ls *Thread, args ...Value) error {}

// Call implements the Callable interface.
//
// Call calls the Lua function with args
// returning all results or error (if any).
// func (fn *Func) Call(ls *Thread, args ...Value) ([]Value, error) {}

// TODO: comment
// func (fn *Func) Cont(ls *Thread, next Continuation) Continuation {
// 	return nil
// }

// Dump dumps a function as a binary chunk.
//
// If strip is true, the binary representation may not
// include all debug information about the function,
func (fn *Func) Dump(w io.Writer, strip bool) (int, error) {
	return code.Dump(w, fn.proto, strip)
}

// TODO: comment
func (fn *Func) cont(ls *thread, fnID, retc int) *call {
	ls.check(fn.proto.StackN)
	var ( 
		argc = (ls.top - fnID) - 1
		base int
	)
	if fn.proto.Vararg {
		var (
			fixed = fn.proto.ParamN
			first = ls.top - argc
			param int
		)
		base = ls.top
		for param < fixed && param < argc {
			ls.stack[ls.top] = ls.stack[first+param]
			ls.stack[first+param] = nil
			ls.top++
			param++
		}
		for param < fixed {
			ls.stack[ls.top] = nil
			ls.top++
			param++
		}
	} else {
		for argc < fn.proto.ParamN {
			ls.stack[ls.top] = nil
			ls.top++
			argc++
		}
		base = fnID + 1
	}
	ci := &call{prev: ls.calls, ls: ls}
	ci.top  = base + fn.proto.StackN
	ls.top  = ci.top
	ci.fn   = fn
	ci.fnID = fnID
	ci.base = base
	ci.retc = retc
	ci.flag = luacall
	return ci
}

// TODO: comment
func (fn *Func) call(ls *thread) { execute(ls) }

// func (fn *Func) run(fs *function) (Continuation, error) {
// 	fmt.Println("*Func.run!")
// 	return nil, nil
// }

// TODO: comment
func (fn *Func) close(level int) {
	for _, up := range fn.up {
		if up.index >= level {
			up.close()
		}
	}
}

// TODO: comment
func (fn *Func) open(ls *thread, encup ...*upvar) {
	cls := closure{fn: fn, up: make([]*upvar, len(fn.proto.UpVars))}
	fn.closure = cls
	for i, up := range fn.proto.UpVars {
		if up.Stack {
			// upvalue refers to local variable
			instack := ls.stack[up.Index]
			cls.up[i] = &upvar{
				stack:  &instack,
				index:  up.Index,
				isopen: true,
			}
		} else {
			// upvalue is in enclosing function
			cls.up[i] = encup[up.Index]
		}
	}
}

// TODO: comment
func (fn *Func) info(ci *call) *FuncInfo {
	var (
		name, what string
		caller = ci.prev
		proto  = fn.proto
	)
	if caller != nil && caller.isLua() && (ci.flag & tailcall == 0) {
		if name, what = funcNameFromCode(caller); what == "" {
			name = ""
		}
	}
	var ( source, short, kind string )
	if source = proto.Source; source == "" {
		source = "=?"
	}
	short = chunkID(source)
	if kind = "Lua"; proto.SrcPos == 0 {
		kind = "main"
	}
	line := -1
	if proto.PcLine != nil {
		line = int(proto.PcLine[ci.pc-1])
	}
	return &FuncInfo{
		Source:   source,
		Short:    short,
		Name:     name,
		What:     what,
		Kind:     kind,
		Lines:    nil,
		ParamN:   proto.ParamN,
		UpVarN:   len(fn.up),
		AtLine:   line,
		LineDef:  proto.SrcPos,
		LineEnd:  proto.EndPos,
		Vararg:   proto.Vararg,
		Tailcall: (ci.flag & tailcall != 0),
	}
}

// kst returns the function's i'th constant.
func (fn *Func) kst(i int) (c Constant) {
	switch kst := fn.proto.Consts[i].(type) {
		case float64:
			return Float(kst)
		case string:
			return String(kst)
		case int64:
			return Int(kst)
		case bool:
			if kst {
				return True
			}
			return False
	}
	return c
}