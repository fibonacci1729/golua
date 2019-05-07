package lua

import (
	"fmt"
	"github.com/Azure/golua/lua/code"
)

type (
	// TODO: comment
	Callable interface {
		// Function
		callable
		Value

		// Call(Continuation) Continuation
		// CallN(*Thread, []Value, int) []Value
		// Call1(*Thread, ...Value) Value
		// Call(*Thread, ...Value) []Value
	}

	// TODO: comment
	Value interface {
		value
		String() string
		// Type() string
		Kind() string
	}

	// TODO: comment
	value interface {
		kind() code.Type
	}
)

// TODO: comment
type hasMeta interface {
	setMeta(*Table)
	getMeta() *Table
}

//
// Builtin types
//

func (v *GoValue) Kind() string { return v.kind().String() }
func (v *closure) Kind() string { return v.kind().String() }
func (v *Thread) Kind() string { return v.kind().String() }
func (v *Table) Kind() string { return v.kind().String() }
func (v String) Kind() string { return v.kind().String() }
func (v Float) Kind() string { return v.kind().String() }
func (v Int) Kind() string { return v.kind().String() }
func (v Bool) Kind() string { return v.kind().String() }

func (*GoValue) kind() code.Type { return code.GoType }
func (*closure) kind() code.Type { return code.FuncType }
func (*Thread) kind() code.Type { return code.ThreadType }
func (*Table) kind() code.Type { return code.TableType }
func (String) kind() code.Type { return code.StringType }
func (Float) kind() code.Type { return code.FloatType }
func (Int) kind() code.Type {  return code.IntType }
func (Bool) kind() code.Type { return code.BoolType }

//
// Go user types (userdata)
//

// TODO: comment
type GoValue struct {
	Value interface{}
	Meta  *Table
}

// TODO: comment
func (v *GoValue) String() string { return fmt.Sprintf("userdata: %p", v) }

// TODO: comment
func typeKind(v Value) code.Type {
	if v == nil {
		return code.NilType
	}
	return v.kind()
}