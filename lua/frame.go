package lua

import (
	"strings"
	"fmt"

	"github.com/Azure/golua/lua/code"
)

type (
	// TODO: comment
	FuncInfo struct {
		Source   string
		Short    string // chunk
		Name 	 string
		What 	 string
		Kind 	 string
		Lines    []int
		ParamN 	 int
		UpVarN 	 int
		AtLine 	 int
		LineDef  int
		LineEnd  int
		Vararg   bool
		Tailcall bool
	}

	// TODO: comment
	VarInfo struct {
		Name string
		Kind string
	}

	// TODO: comment
	Caller interface {
		CurrentLine() int
		IsTail() bool
		PC() int
		Caller() Caller
	}

	// TODO: comment
	Frame interface {
		// Closure() Function
		// Function(opts string) Function
		// Call() Call
		// Stack() []Value
		// Event() Hook
		Traceback() StackTrace
		FuncInfo() *FuncInfo
		FuncName() string
		Callable() Callable
		// Unwind(int) Frame
		Caller() Frame
		// Var(int) *VarInfo
		Stack() []Value

		Source() (string, int)
		Detail() string
		Where() string
		Level() int
		IsMain() bool
		IsLua() bool
		IsGo() bool
		Tailcall() bool

		Errorf(string, ...interface{}) *Error
		Error(error) *Error
	}

	Call interface {
		// Name() string
		// What() string
		// Kind() string
		// Line() int
		// PC() int
		IsTail() bool
	}
)

func (vi *VarInfo) String() string {
	if vi.Kind != "" {
		return fmt.Sprintf("(%s '%s')",
			vi.Kind,
			vi.Name,
		)
	}
	return ""
}

type frame struct { *call }

// func (fs *function) Details() (what, kind string) {
// 	what = fs.info().what // "global", "local", "method", "field", "upvalue", ""
// 	kind = fs.info().kind // "Lua", "Go", "main"
// 	return what, kind
// }

// func (fs *function) Source() (file, chunk string) {
// 	file = fs.info().source
// 	chunk = fs.info().short
// 	return file, chunk
// }

// func (fs *function) Stack() []Value {
// 	switch fn := fs.fn.(type) {
// 		case *GoFunc:
// 			return fs.va
// 		case *Func:
// 			return fn.stack[fs.base:]
// 	}
// 	return nil
// }

// func (fs *function) Span() (def, end int) {
// 	def = fs.info().lineDef
// 	end = fs.info().lineEnd
// 	return def, end
// }

// func (fs *function) Name() string {
// 	return fs.info().name
// }

// func (fs *function) Line() int {
// 	return fs.info().atLine
// }

// func (fs *function) Up(n int) Value {
// 	return nil
// }

// func (fs *function) In(n int) Value {
// 	return nil
// }

// func (fs *function) info() *FuncInfo {
// 	return fs.fn.info(fs.call)
// }

// func (fr frame) Function() Function {
// 	if ci := fr.call; ci != nil {
// 		return &function{ci}
// 	}
// 	return nil
// }

	// if name := funcNameFromGlobals(ci); name != "" {
	// 	return fmt.Sprintf("function '%s'", name)
	// }
	// switch info := ci.fn.info(ci); {
	// 	case info.What != "":
	// 		return fmt.Sprintf("%s '%s'", info.What, info.Name)
	// 	case info.Kind == "main":
	// 		return "main chunk"
	// 	case info.Kind != "Go":
	// 		return fmt.Sprintf("function <%s:%d>", info.Short, info.LineDef)
	// }
	// return "?"

func (fr frame) FuncInfo() *FuncInfo {
	if fr.call != nil {
		return fr.fn.info(fr.call)
	}
	return nil
}

func (fr frame) FuncName() string {
	if fr.call != nil {
		return funcName(fr.call)
	}
	return ""
}

func (fr frame) Callable() Callable {
	if fr.call != nil {
		return fr.fn.(Callable)
	}
	return nil
}

func (fr frame) Tailcall() bool {
	return fr.call != nil && fr.flag & tailcall != 0
}

func (fr frame) Errorf(format string, args ...interface{}) *Error {
	return fr.Error(fmt.Errorf(format, args...))
}

func (fr frame) Error(err error) *Error {
	return &Error{frame: fr, value: String(err.Error())}
}

func (fr frame) Stack() []Value {
	// if fr.call == nil || fr.fnID == fr.sp {
	// 	return nil
	// }
	// return fr.ls.stack.stk[fr.fnID+1:fr.ls.stack.top]
	return nil
}

// func (fr frame) Var(n int) *VarInfo {
// 	if fn, ok := fr.fn.(*Func); ok {
// 		if n >= 0 && n < len(fn.stack[fr.base:]) {
// 			return fn.variable(fr.call, n)
// 		}
// 		return nil
// 	}
// 	if n >= 0 && n < len(fr.va) {
// 		return &VarInfo{
// 			Name: fmt.Sprintf("param #%d", n),
// 			Kind: "local",
// 		}
// 	}
// 	return nil
// }

func (fr frame) Detail() (detail string) {
	if ci := fr.call; ci != nil && ci.fn != nil {
		switch fn := ci.fn.(type) {
			case *GoFunc:
				detail = fmt.Sprintf("[Go]: in %s (%s)", funcName(ci), fn.detail())
			case *Func:
				info := fn.info(ci)
				if detail = info.Short; info.AtLine >= 0 {
					detail = fmt.Sprintf("%s:%d", detail, info.AtLine)
				}
				if info.Kind == "main" {
					detail = fmt.Sprintf("%s: in main chunk", detail)
				} else {
					detail = fmt.Sprintf("%s: in %s", detail, funcName(ci))
				}
		}
	}
	return detail
}

func (fr frame) Source() (file string, line int) {
	if ci := fr.call; ci != nil && ci.fn != nil {
		if fn, ok := ci.fn.(*Func); ok {
				info := fn.info(ci)
				file  = info.Short
				line  = info.AtLine
				return file, line
		}
	}
	return "", -1
}

func (fr frame) Where() (where string) {
	if ci := fr.call; ci != nil && ci.fn != nil {
		switch fn := ci.fn.(type) {
			case *GoFunc:
				where = "[Go]"
			case *Func:
				info := fn.info(ci)
				if where = info.Short; info.AtLine >= 0 {
					where = fmt.Sprintf("%s:%d", where, info.AtLine)
				}
		}
	}
	return where
}

func (fr frame) Traceback() (stack StackTrace) {
	for ci := fr.call; ci != nil; ci = ci.prev {
		stack = append(stack, frame{ci})
	}
	return stack
}

func (fr frame) Caller() Frame {
	if prev := fr.call.prev; prev != nil {
		return frame{prev}
	}
	return nil
}

func (fr frame) Level() (level int) {
	for ci := fr.call; ci != nil; ci = ci.prev {
		level++
	}
	return level
}

func (fr frame) IsMain() bool { return fr.call == nil }
func (fr frame) IsGo() bool {  return fr.IsLua() }
func (fr frame) IsLua() bool { return !fr.IsMain() && fr.call.isLua() }

// short_src:[line:] in function '<name>'
// short_src:[line:] in <what> '<name>'
// short_src:[line:] in main chunk
// short_src:[line:] in function <<short_src>:<line_def>>
// short_src:[line:] in ?
func (fr frame) String() string {
	// if fr.call == nil {
	// 	return "[Go]: in ?"
	// }
	// switch info := fr.fn.info(fr.ls); {
	// 	case info.what != "":
	// 		return fmt.Sprintf(
	// 			"%s: in %s '%s'",
	// 			fr.Where(),
	// 			info.what,
	// 			info.name,
	// 		)
	// 	case info.kind == "main":
	// 		return fmt.Sprintf("%s: in main chunk", fr.Where())
	// 	case info.kind != "Go":
	// 		return fmt.Sprintf(
	// 			"%s: in function <%s:%d>",
	// 			fr.Where(),
	// 			info.short,
	// 			info.lineDef,
	// 		)
	// }
	// return fmt.Sprintf("%s: in ?", fr.Where())
	return ""
}

type callstatus int

const (
	allowhooks callstatus = 1 << iota
	luacall
	hooked
	fresh
	ypcall
	tailcall
	hookyield
	lt4le
	finalizer
	mainfunc
)

type call struct {
	flag callstatus
	fn   callable
	va   []Value
	ls   *thread
	pc   int
	top  int
	base int
	retc int
	fnID int
	err  error
	prev *call
	next *call
}

func (ci *call) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "call(prev=%p, next=%p, err=%v) @ ", ci.prev, ci.next, ci.err)
	fmt.Fprintf(&b, "[func=%d base=%d, retc=%d, top=%d, pc=%d]",
		ci.fnID,
		ci.base,
		ci.retc,
		ci.top,
		ci.pc)
	return b.String()
}

func (ci *call) isLua() bool { return ci.flag & luacall != 0 }

func (ci *call) hook(flag Hook) error {
	// if hooks := ls.hs; (hooks.mask & flag != 0) {
	// 	if hook := hooks.hook; hook != nil {
	// 		if flag & HookCount != 0 {
	// 			if hooks.count--; hooks.count == 0 {
	// 				hooks.count = hooks.after
	// 				dbg := ci.debug("")
	// 				dbg.event = flag
	// 				return hook(ls.tt, dbg)
	// 			}
	// 		} else {
	// 			dbg := ci.debug("")
	// 			dbg.event = flag
	// 			return hook(ls.tt, dbg)
	// 		}
	// 	}
	// }
	return nil
}

// TODO: comment
func (ci *call) varinfo(o *Value) *VarInfo {
	var (
		name string
		kind string
	)
	if fn, ok := ci.fn.(*Func); ok {
		for i, up := range fn.up {
			if v := up.get(); &v == o {
				if kind = "upvalue"; len(fn.proto.UpVars) > 0 {
					if name = "?"; fn.proto.UpVars[i].Name != "" {
						name = fn.proto.UpVars[i].Name
					}
				}
				break
			}
		}
		if kind == "" {
			index := -1
			for fp := 0; fp < len(ci.ls.stack[ci.base:ci.top]); fp++ {
				if o == &(ci.ls.stack[ci.base+fp]) {
					index = fp
					break
				}
			}
			if index >= 0 {
				name, kind = code.ObjectName(fn.proto, ci.pc-1, index)
			}
		}
	}
	return &VarInfo{Name: name, Kind: kind}
}

func (ci *call) unwind(level int) (caller *call) {
	for caller = ci; caller != nil && level > 0; {
		caller = caller.prev
		level--
	}
	return caller
}

func (ci *call) errorf(format string, args ...interface{}) (err error) {
	if err = fmt.Errorf(format, args...); ci.isLua() {
		info := ci.fn.info(ci)
		err = fmt.Errorf("%s:%d: %v", info.Short, info.AtLine, err)
	}
	return &Error{frame: frame{ci}, value: String(err.Error())}
}