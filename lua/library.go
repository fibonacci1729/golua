package lua

import (
	"strings"
	"fmt"
	"os"
)

type Module interface {
	// Opener
	Name() string
}

// TODO: comment
func Library(name string, load Loader) Module {
	return &library{name, load}
}

// TODO: comment
type Loader func(*Thread) (Value, error)

// TODO: comment
type library struct {
	name string
	load Loader
}

// TODO: comment
func (lib *library) Name() string { return lib.name }


// // TODO: comment
// type Lookup func(string) (Loader, error)

// // TODO: comment
// func (fn Lookup) Search(module string) (Loader, error) {
// 	// If fn == nil ?
// 	return fn(module)
// }

// @path templates
//
// A typical path is a list of directories wherein to search for a given file.
// The path used by require is a list of "templates", each of them specifying
// an alternative way to transform a module name (the argument to require) into
// a file name. Each template in the path is a file name containing optional
// question marks. For each template, require substitutes the module name for
// each question mark and checks whether there is a file with the resulting name;
// if not, it goes to the next template. The templates in a path are separated by
// semicolons, a character seldom used for file names in most operating systems.
//
// For instance, consider the path: "?;?.lua;c:\windows\?;/usr/local/lua/?/?.lua"
// 
// With this path, the call require"sql" will try to open the following Lua files:
//		sql
//		sql.lua
//		c:\windows\sql
//		/usr/local/lua/sql/sql.lua
//
// The path require uses to search for Lua files is aways the current value of the
// variable "package.path". When the module package is initialized, it sets the
// variable with the value of the environment variable LUA_PATH_5_3; if this
// environment variable is undefined, Lua tries the environment variable LUA_PATH.
// If both are undefined, Lua uses a compiled-defined default path (-E prevents
// the use of these environment variables and forces the default).
//
// When using the value of an environment variable, Lua substitues the default path
// for any substring ";;". For instance, if we set LUA_PATH_5_3 to "mydir/?.lua;;",
// the final path will be the template "mydir/?.lua" followed by the default path.
//
// Path is a template (or list of templates separated by ';') used to search for
// Lua and Go packages.
type PathTpl string

// Expand searches for the given name in the given path.
//
// A path is a string containing a sequence of templates separated by semicolons.
// For each template, the function replaces each interrogation mark (if any) in
// the template with a copy of name wherein all occurences of sep (a dot, by default)
// were replaced by rep (the system's directory separator, by default), and then tries
// to ope nthe resulting file name.
//
// For instance, if the path is the string:
//
//		"./?.lua;./.lc;/usr/local/?/init.lua"
//
// The search for the name "foo.bar" will try to open the files (in order):
//
//		"./foo/bar.lua"
//		"./foo/bar.lc"
//		"/usr/local/foo/bar/init.lua"
//
// Returns the resulting name of the first file that it can open in read mode
// (after closing the file), or "" and the error if none succeeds (this error
// message lists all the file names it tried to open).
func (tpl PathTpl) Expand(name, sep, dirsep string) (string, error) {
	path := string(tpl)
	if path == "" || name == "" {
		return "", nil
	}
	if sep != "" {
		// non-empty separator then replace it by 'dirsep'
		name = strings.Replace(name, sep, dirsep, -1)
	}
	path = strings.Replace(path, PATH_MARK, name, -1)
	path = strings.TrimSuffix(path, PATH_SEP)
	var b strings.Builder

	for _, file := range strings.Split(path, PATH_SEP) {
		f, err := os.OpenFile(file, os.O_RDONLY, 0666)
		if f.Close(); err != nil {
			switch {
				case os.IsPermission(err):
					// file is not readable
					fmt.Fprintf(&b, "\n\tno file '%s'", file)
					continue
				case os.IsNotExist(err):
					// file does not exist
					fmt.Fprintf(&b, "\n\tno file '%s'", file)
					continue
			}
			// uh-oh
			panic(err)
		}
		return file, nil
	}
	return "", fmt.Errorf("%s", b.String())
}

// Lua path
type Path string

// Search implements the Searcher interface.
func (tpl Path) Search(name string) (Loader, error) {
	f, err := PathTpl(tpl).Expand(name, ".", MOD_SEP)
	if err != nil {
		return nil, fmt.Errorf("module '%s' not found:%v", name, err)
	}
	return Loader(func(ls *Thread) (Value, error) {
		fmt.Printf("%s: load '%s'!\n", f, name)
		return nil, nil
	}), nil
}