package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/golua/lua"
)

var _ = fmt.Println

func must(err error) {
    if err, ok := err.(*lua.Error); ok {
        err.WriteTrace(os.Stderr)
        os.Exit(0)
    }
    if err != nil {
        fmt.Println(err)
    }
}

func main() {
	flag.Parse()

	thread := lua.Init(new(lua.Config))
	fn, err := lua.LoadFile(thread, "main.lua", 0)
	must(err)
	rets, err := thread.Call(fn)
	must(err)
	fmt.Println(rets)
}