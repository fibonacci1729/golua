package lua

import (
	"fmt"
	"io"
)

type (
	// TODO: comment
	Causer interface {
		Cause(Frame) error
	}

	// TODO: comment
	TypeError struct {
		Value  Value
		Expect string
	}

	// TODO: comment
	ArgError struct {
		Arg int
		Err error
	}

	// TODO: comment
	OpError struct {
		op  event
		x   Value
		y   Value
		msg string
	}

	// TODO: comment
	Error struct {
		frame frame
		value Value
	}
)

// TODO: comment
func (e *TypeError) Cause(fr Frame) error {
	if fr == nil {
		return fmt.Errorf("%s expected, got %s", e.Expect, typeKind(e.Value))
	}
	return fmt.Errorf("%s expected, got %s", e.Expect, typeKind(e.Value))
}

// TODO: comment
func (e *TypeError) Error() string {
	return e.Cause(nil).Error()	
}

// TODO: comment
func (e *ArgError) Cause(fr Frame) error {
	if fr == nil {
		return fmt.Errorf("bad argument #%d (%v)", e.Arg+1, e.Err)
	}
	info := fr.FuncInfo()
	if info.What == "method" {
		if e.Arg--; e.Arg == 0 {
			if err, ok := e.Err.(Causer); ok {
				e.Err = err.Cause(fr)
			}
			return fmt.Errorf("calling '%s' on bad self (%s)", info.Name, e.Err.Error())
		}
	}
	if info.Name == "" {
		info.Name = fr.FuncName()
	}
	return fmt.Errorf("bad argument #%d to '%s' (%s)", e.Arg+1, info.Name, e.Err.Error())
}

// TODO: comment
func (e *ArgError) Error() string {
	return e.Cause(nil).Error()
}

// TODO: comment
func (e *OpError) Error() string {
	return "OpError"
}

// TODO: comment
func (e *Error) WriteTrace(w io.Writer) error {
	msg := fmt.Sprintf("golua: %s", e.Error())
	return e.Traceback().WriteTo(w, msg)
}

// TODO: comment
func (e *Error) Traceback() (stack StackTrace) {
	if call := e.frame.call; call != nil {
		stack = append(stack, e.frame)
		for ; call.prev != nil; call = call.prev {
			stack = append(stack, frame{call.prev})
		}
	}
	return stack
}

// TODO: comment
func (e *Error) Error() string {
	if ci := e.frame.call; ci != nil {
		return e.format(ci.ls)
	}
	return e.format(nil)
}

// TODO: comment
func (e *Error) Frame() Frame { return e.frame }

// TODO: comment
func (e *Error) Value() Value { return e.value }

func (e *Error) format(ls *thread) string {
	if s, ok := AsString(e.value); ok {
		return string(s)
	}
	if ls != nil {
		v, ok, err := ls.callMeta(e.value, "__tostring")
		if err != nil {
			return err.Error()
		}
		if ok && IsString(v) {
			return v.String()
		}
		return fmt.Sprintf(
			"(error object is a %s value)",
			ls.typeOf(e.value).Name(),
		)
	}
	return fmt.Sprintf("(error object is a %s)", typeKind(e.value))
}

type errorString string

func (e errorString) Error() string { return string(e) }

func (e errorString) Cause(fr Frame) error {
	if file, line := fr.Caller().Source(); line >= 0 {
		e = errorString(fmt.Sprintf("%s:%d: %s", file, line, e))
	}
	return e
}