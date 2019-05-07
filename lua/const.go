package lua

import (
	"fmt"
)

type (
	// TODO: comment
	Constant interface {
		Value
		constant()
	}

	// TODO: comment
	Number interface {
		Constant
		number()
	}
)

type String string
func (v String) String() string { return string(v) }
func (String) constant() {}

// Pos translates a relative string position: negative means back from end.
func (v String) Pos(i Int) Int {
	switch n := Int(len(v)); {
		case i >= 0:
			return i
		case -i > n:
			return 0
		default:
			return i + n + 1
	}
}

type Bool bool

const (
	True  = Bool(true)
	False = Bool(false)
)

func (v Bool) String() string { return fmt.Sprintf("%t", bool(v)) }
func (Bool) constant() {}

type Float float64
func (v Float) String() string { return fmt.Sprintf("%f", float64(v)) }
func (Float) constant() {}
func (Float) number() {}

type Uint uint64
func (v Uint) String() string { return fmt.Sprintf("%d", uint64(v)) }
func (Uint) constant() {}
func (Uint) number() {}

type Int int64
func (v Int) String() string { return fmt.Sprintf("%d", int64(v)) }
func (Int) constant() {}
func (Int) number() {}

type nilType byte
const Nil = nilType(0)
func (nilType) String() string { return "nil" }
func (nilType) constant() {}