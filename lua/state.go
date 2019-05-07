package lua

import (
	"sync"
	"fmt"
	"os"
	"github.com/Azure/golua/lua/code"
)

var _ = fmt.Println

type (
	// TODO: comment
	packages struct {
		searchers *Table
		loaded    *Table
		preload   *Table
	}

	// TODO: comment
	environ struct {
		GOLUA_ROOT string
		GOLUA_INIT string
		GOLUAGO    string
		GOLUA      string
	}

	// TODO: comment
	system struct {
		environ *environ
		config  *Config
		stdout  *os.File
		stderr  *os.File
		stdin   *os.File
	}

	// TODO: comment
	runtime struct {
		packages
		globals *Table
		system  *system
		values  *Table
		main    *Thread
		wait    sync.WaitGroup
		types   [code.MaxType]*Table
	}
)

// TODO: comment
func (rt *runtime) init(config *Config) *Thread {
	ls := &thread{runtime: rt, hooks: new(hooks)}

	rt.packages = packages{
		searchers: NewTable(),
		preload:   NewTable(),
		loaded:    NewTable(),
	}

	rt.globals = NewTable()
	rt.values  = NewTable()
	rt.main    = &Thread{ls}
	ls.thread  = rt.main
	config.init(rt)
 
	return ls.thread
}

// TODO: comment
func (rt *runtime) Registry() *Table { return rt.values }
func (rt *runtime) Preload() *Table { return rt.preload }
func (rt *runtime) Loaded() *Table { return rt.loaded }
func (rt *runtime) Stdin() *os.File { return rt.system.stdin }
func (rt *runtime) Stdout() *os.File { return rt.system.stdout }
func (rt *runtime) Stderr() *os.File { return rt.system.stderr }

// TODO: comment
func (rt *runtime) PathVar(env string) Path {
	switch env {
		case GOLUAGO_ENV:
			return Path(rt.system.environ.GOLUAGO)
		case GOLUA_ENV:
			return Path(rt.system.environ.GOLUA)
	}
	return ""
}

// TODO: comment
type thread struct {
	runtime *runtime
	thread  *Thread
	hooks   *hooks
	stack   []Value
	calls   *call
	callN   int
	top     int
}

// return lua.LoadBinary(fn.Thread(), "main.bin")
// return lua.LoadScript(fn.Thread(), "main.lua")
// return lua.LoadChunk(fn.Thread(), "main.lua")

// callMeta calls a metamethod.
//
// If the object has a metatable and this metatable has field, this
// function calls this field passing the object as its only argument.
//
// In this case this function returns the value returned by the call,
// true, and the error (if any).
//
// If there is no metatable or no metamethod, this function returns
// nil, false, and nil.
func (ls *thread) callMeta(object Value, field string) (Value, bool, error) {
	if method := ls.typeOf(object).Field(field); method != nil {
		rets, err := ls.callE(method, []Value{object}, 1)
		if err != nil {
			return nil, true, err
		}
		return rets[0], true, nil
	}
	return nil, false, nil
}

func (ls *thread) recover(err *error) {
	// type formatter interface {
	// 	format(*Thread) error
	// }
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			*err = e
			return
		}
		panic(r)
		// switch e := r.(type) {
		// 	case formatter:
		// 		*err = e.format(ls.thread)
		// 	case error:
		// 		*err = e
		// 	default:
		// 		panic(r)
		// }
	}
}

func (ls *thread) callE(fv Value, args []Value, retc int) (rets []Value, err error) {
	defer ls.recover(&err)
	rets = ls.call(fv, args, retc)
	return rets, err
}

func (ls *thread) call(fv Value, args []Value, retc int) []Value {
	ls.push(append([]Value{fv}, args...)...)
	ci := ls.cont(ls.top-len(args)-1, retc)
	if ls.do(ci); ci.err != nil {
		panic(ci.err)
	}
	return ls.stack[:ls.top]
}

func (ls *thread) cont(fnID, retc int) *call {
	if fn, ok := ls.stack[fnID].(callable); ok {
		return fn.cont(ls, fnID, retc)
	}
	return ls.funcTM(fnID).cont(fnID, retc)
}

// TODO: comment
func (ls *thread) meta(v Value, event string) Value {
	if v != nil {
		return ls.typeOf(v).Field(event)
	}
	return nil
}

// TODO: comment
func (ls *thread) funcTM(fnID int) *thread {
	if mm := ls.meta(ls.stack[fnID], "__call"); mm != nil {
		if _, ok := mm.(callable); ok {
			ls.insert(fnID, mm)
			return ls
		}
	}
	panic(opError(ls, _call, &(ls.stack[fnID]), nil))
}

// TODO: comment
func (ls *thread) typeOf(v Value) *meta {
	switch v := v.(type) {
		case *GoValue:
			return &meta{v.Meta, v}
		case *Table:
			return &meta{v.meta, v}
		default:
			if v == nil {
				return nilMeta
			}
			mt := ls.runtime.types[0x0F&v.kind()]
			return &meta{mt, v}
	}
}

// TODO: comment
func (ls *thread) setMeta(v Value, m *Table) {
	if o, ok := v.(hasMeta); ok {
		o.setMeta(m)
		return
	}
	if v != nil {
		ls.runtime.types[0x0F&v.kind()] = m
	}
}

func (ls *thread) caller(skip int) (ci *call) {
	// for ci = ls.calls; skip > 0; ci = ci.prev {
	// 	skip--
	// }
	// return ci
	return nil
}

// func (ls *thread) traceback(skip int) (stack []*frame) {
// 	for fr := ls.caller(skip); fr != nil; fr = ls.caller(skip) {
// 		stack = append(stack, fr)
// 		skip++
// 	}
// 	return stack

	// var b strings.Builder
	// fmt.Fprint(&b, "stack traceback:")
	// for fp := fr; fp != nil && (fp.call.flag & mainfunc == 0); fp = fp.prev {
	// 	dbg := fp.debug("Slnt")
	// 	fmt.Fprintf(&b, "\n\t%s:", dbg.short)
	// 	if dbg.line > 0 {
	// 		fmt.Fprintf(&b, "%d:", dbg.line)
	// 	}
	// 	// add global function name if necessary
	// 	var name string
	// 	if name = funcNameFromGlobals(fp, dbg); name != "" {
	// 		name = fmt.Sprintf("function '%s'", name)
	// 	} else {
	// 		switch {
	// 			case dbg.what != "":
	// 				name = fmt.Sprintf("%s '%s'", dbg.what, dbg.name)
	// 			case dbg.kind == "main":
	// 				name = "main chunk"
	// 			case dbg.kind != "Go":
	// 				name = fmt.Sprintf("function <%s:%d>", dbg.short, dbg.span[0])
	// 			default:
	// 				name = "?"
	// 		}
	// 	}
	// 	fmt.Fprintf(&b, " in %s", name)
	// 	if dbg.tailcall {
	// 		fmt.Fprintf(&b, "\n\t(...tail calls...)")
	// 	}
	// }
	// fmt.Fprint(&b, "\n\t[Go]: in ?")
	// return b.String()
	// for _, fr := range fr.traceback() {
	// 	stack = append(stack, fr.debug("Slnt"))
	// }
	// return stack
// }

var nilMeta = new(meta)