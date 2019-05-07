package lua

import (
	"strings"
	"fmt"
	"os"
)

//
// Configuration for paths
//
const (
	// Lua search paths.
	LUAPATH_DEFAULT = LUA_DIR + "?.lua;" + LUA_DIR + "?/init.lua;" + GO_DIR + "?.lua;" + GO_DIR + "?/init.lua;" + "./?.lua;" + "./?/init.lua"
	LUA_DIR 	    = ROOT_DEFAULT + "share/lua/5.3/"
	LUAPATH 	    = "GLUA_PATH"

	// Go search paths.
	GOPATH_DEFAULT = GO_DIR + "?.so;" + GO_DIR + "loadall.so;" + "./?.so"
	GO_DIR 		   = ROOT_DEFAULT + "lib/lua/5.3/"
	GOPATH 		   = "GLUA_GOPATH"
	// System root paths.
	ROOT_DEFAULT   = "/usr/local/"
	ROOT   	       = "LUA_ROOT"

	// PATH_MARK is the string that marks the substitution points in a template.
	PATH_MARK = "?"

	// PATH_SEP is the character that separates templates in a path.
	PATH_SEP = ";"

	// EXEC_DIR in a Windows path is replaced by the executable's directory.
	EXEC_DIR = "!"

	// LUA_DIRSEP is the directory separator (for submodules).
	//
	// CHANGE it if your machine does not use "/" as the directory
	// separator and is not Windows.
	//
	// Windows Lua automatically uses "\".
	DIR_SEP = "/"

	// MOD_SEP (LUA_CSUBSEP/LUA_LSUBSEP) is the character that replaces dots
	// in submodule names when searching for a Go/Lua loader.
	MOD_SEP = DIR_SEP

	// LUA_IGMARK is a mark to ignore all before it when building
	// the luaopen_ function name.
	IGNORE_MARK = "-"
)

const (
	// Some space for error handling
	errorStackSize = stackMax + 200

	// RegistryIndex is the pseudo-index for the registry table.
	registryIndex = -stackMax - 1000

	// MainThreadIndex is the registry index for the main thread of the main state.
	mainthread = Int(1) //= lua.Int(lua.MAIN_THREAD_INDEX)

	// GlobalsIndex is the registry index for the global environment.
	globals = Int(2) //= lua.Int(lua.GLOBALS_INDEX)

	// Key, in the registry, for table of loaded modules.
	loadedKey = "_LOADED"

	// Key, in the registry, for table or preloaded loaders.
	preloadKey = "_PRELOAD"

	// Initial space allocate for UpValues.
	initNumUps = 5

	// Maximum number of upvalues in a closure (both lua and go).
	// Value must fit in a VM register.
	maxUpVars = 255

	// Limit for table tag-method chains (to avoid loops).
	maxMetaLoop = 2000

	// Limit for table tag-method chains (to avoid loops).
	metaLoopMax = 10

	// Maximum depth for nested Go calls and syntactical nested non-terminals
	// in a program.
	//
	// Value must be < 255.
	maxCalls = 255

	// Maximum valid index and maximum size of stack.
	stackMax = 1000000

	// Minimum Lua stack available to a function.
	stackMin = 20

	// Size allocated for new stacks.
	stackNew = 2 * stackMin

	// Extra stack space to handle metamethod calls and some other extras.
	extraStack = 5

	// Number of list items to accumulate before a SETLIST instruction.
	fieldsPerFlush = 50

	// Option for multiple returns in 'pcall' and 'call'
	multRet = -1
)

const (
	// System variables / defaults
	GOLUA_ROOT = "/usr/local"
	GOLUA_PKG  =  GOLUA_ROOT + "/lib/lua/5.3/"
	GOLUA_SRC  =  GOLUA_ROOT + "/share/lua/5.3/"
	
	// Go environment variables
	GOLUAGO_V53_ENV = "GOLUAGO_V53"
	GOLUAGO_ENV     = "GOLUAGO"
	
	// Lua environment variables
	GOLUA_V53_ENV = "GOLUA_V53"
	GOLUA_ENV     = "GOLUA"

	GOLUA_INIT_V53_ENV = "GOLUA_INIT_V53"
	GOLUA_INIT_ENV     = "GOLUA_INIT"

	// Environment variable defaults
	GOLUAGO_DEFAULT = GOLUA_SRC+"?.so;"+GOLUA_SRC+"loadall.so;"+"./?.so;"
	GOLUA_DEFAULT   = GOLUA_SRC+"?.lua;"+GOLUA_SRC+"?/?/.lua;"+GOLUA_PKG+"?.lua;"+GOLUA_PKG+"?/?/.lua;"+"./?.lua;"+"./?/init.lua;"
)

type Config struct {
	GOLUA_INIT string
	GOLUAGO    string
	GOLUA      string
	Trace      bool
	NoEnv      bool
}

func (config *Config) init(rt *runtime) {
	var (
		env = &environ{
			GOLUAGO: GOLUAGO_DEFAULT,
			GOLUA:   GOLUA_DEFAULT,
		}
		sys = &system{
			environ: env, 
			stdout:  os.Stdout,
			stderr:  os.Stderr,
			stdin:   os.Stdin,
			config:  config,
		}
	)
	if config.GOLUA_INIT != "" {
		env.GOLUA_INIT = config.GOLUA_INIT
	}
	if config.GOLUAGO != "" {
		env.GOLUAGO = config.GOLUAGO
	}
	if config.GOLUA != "" {
		env.GOLUA = config.GOLUA
	}
	if 	rt.system = sys; !config.NoEnv {
		env.GOLUA_INIT = envvar(config, GOLUA_INIT_ENV, GOLUA_INIT_V53_ENV, env.GOLUA_INIT)
		env.GOLUAGO = envvar(config, GOLUAGO_ENV, GOLUAGO_V53_ENV, env.GOLUAGO)
		env.GOLUA   = envvar(config, GOLUA_ENV, GOLUA_V53_ENV, env.GOLUA)
	}
}

func envvar(config *Config, envVar, envVer, defVal string) (path string) {
	versioned := fmt.Sprintf("%s%s", envVar, "_5_3")
	if path = os.Getenv(versioned); path == "" {
		path = os.Getenv(envVar)
	}
	if path == "" {
		path = defVal
	} else {
		path = strings.Replace(path, ";;", "; ;", -1)
		path = strings.Replace(path, "; ;", defVal, -1)
	}
	return path
}