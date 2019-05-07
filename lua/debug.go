package lua

import (
	"strings"
	"fmt"
	"io"

	"github.com/Azure/golua/lua/code"
)

// // TODO: comment
type StackTrace []Frame

// TODO: comment
func (stack StackTrace) WriteTo(w io.Writer, msg string) error {
	const (
		levels1 = 10 // size of the 1st part of the stack
		levels2 = 11 // size of the 2nd part of the stack
	)
	fmt.Fprintf(w, "%s\nstack traceback:", msg)
	for _, frame := range stack {
		fmt.Fprintf(w, "\n\t%s", frame.Detail())
		if frame.Tailcall() {
			fmt.Fprint(w, "\n\t(...tail calls...)")
		}
	}
	fmt.Fprintln(w, "\n\t[Go]: in ?")
	return nil
}

// TODO: comment
type Hooks interface {
	// OnReturn sets the hook to be called when the interpreter returns
	// from a function. The hook is called just before Lua leaves the
	// function.
	//
	// There is no standard way to access the values to be returned by
	// the function.
	OnReturn(HookFn) Hooks

	// OnCount sets the hook to be called after the interpreter
	// executes every count instructions.
	//
	// This event only happens while Lua is executing a Lua function.
	OnCount(HookFn, uint) Hooks
	
	// OnLine sets the hook to be called when the interpreter is about
	// to start the execution of a new line of code, or when it jumps
	// back in the code (even to the same line).
	//
	// This event only happens while Lua is executing a Lua function.
	OnLine(HookFn) Hooks

	// OnCall sets the hook to be called the interpreter calls a function.
	// The hook is called just after Lua enters the new function, before
	// the function gets its arguments.
	OnCall(HookFn) Hooks

	// Reset resets the hooks to its initial configuration.
	Reset()

	// String returns a string detailing the current hooks configuration.
	String() string
}

type HookFn func(*Thread, Frame) error

type Hook int

const (
 	HookTailcall Hook = 1 << iota 
	HookReturn
	HookCount
	HookLine
	HookCall
)

type hooks struct {
	hook  HookFn
	mask  Hook
	allow bool
	after uint
	count uint
}

func (h *hooks) OnCount(hook HookFn, count uint) Hooks {
	h.mask |= HookCount
	h.after = count
	h.count = count
	h.hook = hook
	return h
}

func (h *hooks) OnReturn(hook HookFn) Hooks {
	h.mask |= HookReturn
	h.hook = hook
	return h
}

func (h *hooks) OnLine(hook HookFn) Hooks {
	h.mask |= HookLine
	h.hook = hook
	return h
}

func (h *hooks) OnCall(hook HookFn) Hooks {
	h.mask |= (HookCall|HookTailcall)
	h.hook = hook
	return h
}

func (h *hooks) Reset() {
	h.count = h.after
	h.hook  = nil
	h.mask  = 0
}

func (h *hooks) String() string {
	var flags []string
	if h.mask & HookTailcall != 0 {
		flags = append(flags, "tailcall")
	}
	if h.mask & HookReturn != 0 {
		flags = append(flags, "return")
	}
	if h.mask & HookCount != 0 {
		flags = append(flags, "count")
	}
	if h.mask & HookLine != 0 {
		flags = append(flags, "line")
	}
	if h.mask & HookCall != 0 {
		flags = append(flags, "call")
	}
	return strings.Join(flags, ",")
}

func funcNameFromGlobals(ci *call) (name string) {
	var (
		fn = ci.fn.(Value)
		rs = ci.ls.runtime
		ld = rs.loaded
		found bool
	)
	name, found = searchField(ld, &fn, 2)
	if found && strings.HasPrefix(name, "_G.") {
		name = name[3:]
	}
	return name
}

func searchField(env Value, fn *Value, level int) (name string, found bool) {
	if tbl, ok := env.(*Table); ok && level > 0 {
		tbl.foreach(func(k, v Value) bool {
			if k, isstr := k.(String); isstr {
				if found, _ = equals(nil, fn, &v); found {
					name = string(k)
					return false
				}
			}
			var s string
			s, found = searchField(v, fn, level-1)
			if found {
				name = fmt.Sprintf("%s.%s", k, s)
				return false
			}
			return true
		})
	}
	return name, found
}

// funcNameFromCode tries to find a name for a function based on
// the code that called it. Only works when function was called
// by a Lua function.
//
// Returns what the name is (e.g., "for iterator", "method",
// "metamethod") and sets '*name' to point to the name.
func funcNameFromCode(ci *call) (name, what string) {
	if fn, ok := ci.fn.(*Func); ok {
		if ci.flag & hooked != 0 {
			return "?", "hook"
		}
		var (
			inst = fn.proto.Instrs[ci.pc-1]
			meta event
		)
		switch inst.Code() {
			case code.SELF, code.GETTABUP, code.GETTABLE:
				meta = _index
			case code.SETTABUP, code.SETTABLE:
				meta = _newindex
			case code.CALL, code.TAILCALL:
				return code.ObjectName(fn.proto, ci.pc-1, inst.A())
			case code.TFORCALL:
				return "for iterator", "for iterator"
			case code.IDIV,
				code.DIV,
				code.ADD,
				code.SUB,
				code.MUL,
				code.MOD,
				code.POW,
				code.SHL,
				code.SHR,
				code.BOR,
				code.BXOR,
				code.BAND:
				meta = event(inst.Code()-code.ADD) + _add
			case code.CONCAT:
				meta = _concat
			case code.BNOT:
				meta = _bnot
			case code.UNM:
				meta = _unm
			case code.LEN:
				meta = _len
			case code.EQ:
				meta = _eq
			case code.LT:
				meta = _lt
			case code.LE:
				meta = _le
			default:
				return
		}
		return events[meta], "metamethod"
	}
	return "", ""
}

func funcName(ci *call) string {
	if name := funcNameFromGlobals(ci); name != "" {
		return fmt.Sprintf("function '%s'", name)
	}
	switch info := ci.fn.info(ci); {
		case info.What != "":
			return fmt.Sprintf("%s '%s'", info.What, info.Name)
		case info.Kind == "main":
			return "main chunk"
		case info.Kind != "Go":
			return fmt.Sprintf("function <%s:%d>", info.Short, info.LineDef)
	}
	return "?"
}

func chunkID(src string) string {
	if src != "" && len(src) > 1 {
		if src[0] == '@' || src[0] == '=' {
			src = src[1:]
		}
		return src
	}
	return "?"
}