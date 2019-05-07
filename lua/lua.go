package lua

import (
	"os"
)

const EnvID = "_ENV"

// TODO
type Runtime interface {
	// NewMetaType(string) *Table
	// MainThread() *Thread
	// Preload(Module)
	Registry() *Table
	Preload() *Table
	Loaded() *Table
	Stdin() *os.File
	Stdout() *os.File
	Stderr() *os.File
	PathVar(string) Path
}

type Op int

const (
	OpNone Op = iota

	// binary ops
	OpAdd
	OpSub
	OpMul
	OpMod
	OpPow
	OpDivF
	OpDivI
	OpBand
	OpBor
	OpBxor
	OpShl
	OpShr
	OpConcat
	OpEq
	OpLt
	OpLe
	OpNe
	OpGt
	OpGe
	OpAnd
	OpOr

	// unary ops
	OpMinus
	OpBnot
	OpNot
	OpLen
)

var opnames = [...]string{
	OpNone:   "none",
	OpAdd:    "add",
	OpSub:    "sub",
	OpMul:    "mul",
	OpMod:    "mod",
	OpPow:    "pow",
	OpDivF:   "fdiv",
	OpDivI:   "idiv",
	OpBand:   "band",
	OpBor:    "bor",
	OpBxor:   "bxor",
	OpShl:    "shl",
	OpShr:    "shr",
	OpConcat: "concat",
	OpEq:     "eq",
	OpLt:     "lt",
	OpLe:     "le",
	OpNe:     "ne",
	OpGt:     "gt",
	OpGe:     "ge",
	OpAnd:    "and",
	OpOr:     "or",
	OpMinus:  "minus",
	OpBnot:   "bnot",
	OpNot:    "not",
	OpLen:    "len",
}

// TODO: comment
func (op Op) String() string { return opnames[op] }

// TODO: comment
func Init(config *Config) *Thread {
	return new(runtime).init(config)
}