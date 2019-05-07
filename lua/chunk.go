package lua

import "io"

// TODO: comment
type LoadMode int

const (
	BinaryOnly LoadMode = 1 << iota
	TextOnly
)

// DoFileWithEnvN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoFileWithEnvN(ls *Thread, file string, args []Value, want int, env *Table) ([]Value, error) {
	return ls.ExecWithEnvN(file, nil, args, want, env)
}

// DoFileWithEnv loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoFileWithEnv(ls *Thread, file string, args []Value, env *Table) ([]Value, error) {
	return DoFileWithEnvN(ls, file, args, -1, env)
}

// DoFileN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoFileN(ls *Thread, file string, args []Value, want int) ([]Value, error) {
	return DoFileWithEnvN(ls, file, args, want, ls.Globals())
}

// DoFile loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoFile(ls *Thread, file string, args ...Value) ([]Value, error) {
	return DoFileN(ls, file, args, -1)
}

// DoTextWithEnvN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoTextWithEnvN(ls *Thread, text string, args []Value, want int, env *Table) ([]Value, error) {
	return ls.ExecWithEnvN("=(string)", text, args, want, env)
}

// DoTextWithEnv loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoTextWithEnv(ls *Thread, text string, args []Value, env *Table) ([]Value, error) {
	return DoTextWithEnvN(ls, text, args, -1, env)
}

// DoTextN loads the Lua chunk and executes the function with args; want
// indicates the number of results wanted.
//
// Returns #want results or error.
func DoTextN(ls *Thread, text string, args []Value, want int) ([]Value, error) {
	return DoTextWithEnvN(ls, text, args, want, ls.Globals())
}

// DoText loads the Lua chunk and executes the function with args.
//
// All results are returned or error.
func DoText(ls *Thread, text string, args ...Value) ([]Value, error) {
	return DoTextN(ls, text, args, -1)
}

// DoLibrary calls 'require(name)' and stores the result in a global
// variable with the given name.
func DoLibrary(ls *Thread, name string) error {
	rets, err := ls.CallN(ls.Global("require"), []Value{String(name)}, 1)
	if err != nil {
		return err
	}
	ls.SetGlobal(name, rets[0])
	return nil
}

// LoadScript loads a Lua chunk from the specified file in text-only mode.
//
// Returns the function or error.
func LoadScript(ls *Thread, file string) (*Func, error) {
	return ls.Load("=(stdin)", file, TextOnly)
}

// LoadBinary loads a precompile Lua chunk from the specified file in binary-only mode.
//
// Returns the function or error.
func LoadBinary(ls *Thread, file string) (*Func, error) {
	return ls.Load(file, nil, BinaryOnly)
}

// LoadChunk loads a either a binary or text Lua chunk from the specified path.
//
// Returns the function or error.
func LoadChunk(ls *Thread, file string) (*Func, error) {
	return ls.Load(file, nil, 0)
}

// LoadFrom loads a Lua chunk from the specified reader.
//
// Returns the function or error.
func LoadFrom(ls *Thread, from io.Reader, mode LoadMode) (*Func, error) {
	return ls.Load("=(stdin)", from, mode)
}

// LoadFile loads a Lua chunk from the specified file.
//
// Returns the function or error.
func LoadFile(ls *Thread, file string, mode LoadMode) (*Func, error) {
	return ls.Load(file, nil, mode)
}

// LoadText loads a Lua chunk from the specified text input.
//
// Returns the function or error.
func LoadText(ls *Thread, text string) (*Func, error) {
	return ls.Load("=(string)", text, TextOnly)
}
