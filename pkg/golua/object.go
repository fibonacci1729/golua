package golua

import (
	"reflect"
	"strings"
	"fmt"
	"github.com/Azure/golua/lua"
)

//
// Object metamethods (common)
//

func object۰tostring(cont lua.Continuation) []lua.Value {
	return cont.Errorf("object۰tostring")
}

//
// Struct metamethods
//

func struct۰index(cont lua.Continuation) []lua.Value {
	o, k := userdata(cont, 0), strings.Title(string(cont.StringVar(1)))

	if v := o.FieldByName(k); !v.IsValid() {
		if v := o.MethodByName(k); v.IsValid() {
			return []lua.Value{Value(v.Interface())}
		}
	} else {
		return []lua.Value{Value(v.Interface())}
	}
	return cont.Errorf("undefined field/method: '%s' (type '%s')", k, o.Type())
}

func struct۰eq(cont lua.Continuation) []lua.Value {
	return cont.Errorf("struct۰eq")
}

//
// Array metamethods
//

func array۰index(cont lua.Continuation) []lua.Value {
	return cont.Errorf("array۰index")
}

func array۰eq(cont lua.Continuation) []lua.Value {
	return cont.Errorf("array۰eq")
}

//
// Slice metamethods
//

func slice۰newindex(cont lua.Continuation) []lua.Value {
	return cont.Errorf("slice۰newindex")
}

func slice۰index(cont lua.Continuation) []lua.Value {
	return cont.Errorf("slice۰index")
}

func slice۰eq(cont lua.Continuation) []lua.Value {
	return cont.Errorf("slice۰eq")
}

//
// Map metamethods
//

func map۰newindex(cont lua.Continuation) []lua.Value {
	return cont.Errorf("map۰newindex")
}

func map۰index(cont lua.Continuation) []lua.Value {
	return cont.Errorf("map۰index")
}

func map۰eq(cont lua.Continuation) []lua.Value {
	return cont.Errorf("map۰eq")
}

//
// Ptr metamethods
//

func struct۰ptr۰newindex(cont lua.Continuation) []lua.Value {
	return cont.Errorf("struct۰ptr۰newindex")
}

func struct۰ptr۰index(cont lua.Continuation) []lua.Value {
	v := cont.GoValueVar(0)
	fmt.Printf("%T\n", v.Value)
	return cont.Errorf("struct۰ptr۰index")
}

func array۰ptr۰newindex(cont lua.Continuation) []lua.Value {
	return cont.Errorf("array۰ptr۰newindex")
}

func array۰ptr۰index(cont lua.Continuation) []lua.Value {
	return cont.Errorf("array۰ptr۰index")
}

func array۰ptr۰len(cont lua.Continuation) []lua.Value {
	return cont.Errorf("array۰ptr۰len")
}

func ptr۰index(cont lua.Continuation) []lua.Value {
	return cont.Errorf("ptr۰index")
}

func ptr۰eq(cont lua.Continuation) []lua.Value {
	return cont.Errorf("ptr۰eq")
}

func userdata(cont lua.Continuation, arg int) reflect.Value {
	return reflect.ValueOf(cont.GoValueVar(arg).Value)
}