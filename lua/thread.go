package lua

import (
	"fmt"

	"github.com/Azure/golua/lua/luac"
)

// Thread is a Lua thread.
type Thread struct {
	*thread
}

// ExecWithEnvN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func (ls *Thread) ExecWithEnvN(file string, src interface{}, args []Value, want int, env *Table) ([]Value, error) {
	fn, err := ls.LoadWithEnv(file, src, 0, env)
	if err != nil {
		return nil, err
	}
	return ls.CallN(fn, args, want)
}

// ExecWithEnv loads the Lua chunk and executes the function with args.
//
// All results are returned or error.
func (ls *Thread) ExecWithEnv(file string, src interface{}, args []Value, env *Table) ([]Value, error) {
	return ls.ExecWithEnvN(file, src, args, -1, env)
}

// ExecN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func (ls *Thread) ExecN(file string, src interface{}, args []Value, want int) ([]Value, error) {
	return ls.ExecWithEnvN(file, src, args, want, ls.Globals())
}

// Exec loads the Lua chunk and executes the function with args.
//
// All results are returned or error.
func (ls *Thread) Exec(file string, src interface{}, args ...Value) ([]Value, error) {
	return ls.ExecN(file, src, args, -1)
}

// LoadWithEnv loads a compiled Lua chunk into a function value
// setting env as the first upvalue.
//
// Returns the function or error.
func (ls *Thread) LoadWithEnv(file string, src interface{}, mode LoadMode, env *Table) (*Func, error) {
	chunk, err := luac.Compile(file, src)
	if err != nil {
		return nil, err
	}
	return NewFunc(chunk, env), nil
}

// Load loads a compiled Lua chunk into a function value.
//
// Returns the function or error.
func (ls *Thread) Load(file string, src interface{}, mode LoadMode) (*Func, error) {
	return ls.LoadWithEnv(file, src, mode, ls.Globals())
}

// CallN calls the Lua value fv with args returning want results.
//
// Returns #want results or error.
func (ls *Thread) CallN(fv Value, args []Value, want int) (rets []Value, err error) {
	defer ls.recover(&err)
	rets = ls.call(fv, args, want)
	return rets, err
}

// Call1 calls the Lua value fv with args returning want results.
//
// Returns 1 result or error.
func (ls *Thread) Call1(fv Value, args ...Value) (Value, error) {
	rets, err := ls.CallN(fv, args, 1)
	if err != nil {
		return nil, err
	}
	return rets[0], nil
}

// Call0 calls the Lua value fv with args.
//
// Returns the error if any.
func (ls *Thread) Call0(fv Value, args ...Value) error {
	_, err := ls.CallN(fv, args, 0)
	return err
}

// Call calls the Lua value fv with args.
//
// Returns all results or error.
func (ls *Thread) Call(fv Value, args ...Value) ([]Value, error) {
	return ls.CallN(fv, args, -1)
}

// SetGlobal sets the global name to the value global.
func (ls *Thread) SetGlobal(name string, global Value) *Thread {
	if env := ls.Globals(); env != nil {
		env.Set(String(name), global)
	}
	return ls
}

// SetGlobals sets the globals table.
func (ls *Thread) SetGlobals(env *Table) {
	ls.runtime.loaded.Set(String("_G"), env)
	ls.runtime.globals = env
}

// Global returns the global value for name.
func (ls *Thread) Global(name string) Value {
	if env := ls.Globals(); env != nil {
		return env.Get(String(name))
	}
	return nil
}

// Globals returns the current globals.
func (ls *Thread) Globals() *Table {
	return ls.runtime.globals
}

// Runtime returns the Thread's runtime context.
func (ls *Thread) Runtime() Runtime { return ls.runtime }

// SetMeta sets the metatable of the object v as the metatable associated
// with type name in the registry (see luaL_newmetatable).
func (ls *Thread) SetMeta(v Value, meta *Table) Value {
	if v != nil {
		ls.setMeta(v, meta)
	}
	return v
}

// TypeOf returns the runtime type information for the Value obj.
func (ls *Thread) TypeOf(obj Value) Type { return ls.typeOf(obj) }

// Debug returns the Thread's debug interface.
func (ls *Thread) Hooks() Hooks { return ls.hooks }

// String returns a formatted string described the thread.
//
// Implements lua.Value.
func (ls *Thread) String() string { return fmt.Sprintf("thread: %p", ls) }

// Traceback returns a formatted string of the current call stack
// traceback; skip indicates how many frames to skip before tracing.
func (ls *Thread) Traceback() StackTrace { return StackTrace(ls.Callers(0)) }

// Callers returns the call stack after skipping skip levels.
func (ls *Thread) Callers(skip int) (stack []Frame) {
	for ci := ls.caller(skip); ci != nil; ci = ci.prev {
		stack = append(stack, frame{ci})
	}
	return stack	
}

// Caller returns a lua.Debug for the call stack after skipping skip
// levels.
func (ls *Thread) Caller(skip int) Frame {
	return frame{ls.caller(skip)}
}

// IsMainThread reports whether this thread is the main thread.
func (ls *Thread) IsMainThread() bool { return ls == ls.runtime.main }