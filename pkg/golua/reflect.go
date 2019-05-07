package golua

import (
	"reflect"
	"fmt"
	"github.com/Azure/golua/lua"
)

func TypeMeta(any interface{}) (meta *lua.Table) {
	if rv := reflect.ValueOf(any); !rv.IsNil() {
		meta = typeMeta(rv)
	}
	return meta
}

func typeMeta(val reflect.Value) (meta *lua.Table) {
	methods := map[string]lua.Value{
		"__tostring":  lua.GoClosure(object۰tostring),
		"__metatable": lua.True,
	}
	switch typ := val.Type(); typ.Kind() {
	case reflect.Struct:
		return typeMetaFromStruct(val, methods)
	case reflect.Array:
		return typeMetaFromArray(val, methods)
	case reflect.Slice:
		return typeMetaFromSlice(val, methods)
	case reflect.Map:
		return typeMetaFromMap(val, methods)
	case reflect.Ptr:
		return typeMetaFromPtr(val, methods)
	}
	panic(fmt.Errorf("unexpected kind %s", val.Type().Kind()))
}

func typeMetaFromStruct(val reflect.Value, kvs map[string]lua.Value) *lua.Table {
	kvs["__index"] = lua.GoClosure(struct۰index)
	kvs["__eq"] = lua.GoClosure(struct۰eq)
	// fields(val, kvs)
	// methods(val, kvs)
	return lua.NewTableFromMap(kvs)
}

func typeMetaFromArray(val reflect.Value, kvs map[string]lua.Value) *lua.Table {
	// __tostring
	// __index
	// __call
	// __len
	// __eq
	return lua.NewTableFromMap(kvs)
}

func typeMetaFromSlice(val reflect.Value, kvs map[string]lua.Value) *lua.Table {
	// __tostring
	// __newindex
	// __index
	// __call
	// __len
	// __add
	// __concat
	return lua.NewTableFromMap(kvs)
}

func typeMetaFromPtr(val reflect.Value, kvs map[string]lua.Value) *lua.Table {
	kvs["__index"] = lua.GoClosure(ptr۰index)
	kvs["__eq"] = lua.GoClosure(ptr۰eq)
	switch typ := val.Type(); typ.Elem().Kind() {
	case reflect.Array:
		kvs["__newindex"] = lua.GoClosure(array۰ptr۰newindex)
		kvs["__index"] = lua.GoClosure(array۰ptr۰index)
		kvs["__len"] = lua.GoClosure(array۰ptr۰len)
	case reflect.Struct:
		kvs["__newindex"] = lua.GoClosure(struct۰ptr۰newindex)
		kvs["__index"] = lua.GoClosure(struct۰ptr۰index)
	}
	return lua.NewTableFromMap(kvs)
}

func typeMetaFromMap(val reflect.Value, kvs map[string]lua.Value) *lua.Table {
	// __tostring
	// __newindex
	// __index
	// __len
	// __call
	return lua.NewTableFromMap(kvs)
}

// func methods(val reflect.Value, kvs map[string]lua.Value) {
// 	for i := 0; i < val.Type().NumMethod(); i++ {
// 		if method := val.Type().Method(i); method.PkgPath == "" {
// 			kvs[snakeCase(method.Name)] = Func(method.Func)
// 		}
// 	}
// }

// func fields(val reflect.Value, kvs map[string]lua.Value) {
// 	for i := 0; i < val.Type().NumField(); i++ {
// 		if field := val.Type().Field(i); field.PkgPath == "" {
// 			if !field.Anonymous {
// 				continue
// 			}
// 			k := snakeCase(strings.ToLower(field.Name))
// 			v := val.
// 			kvs[key] = 
// 		}
// 		// if m := typ.Method(i); m.PkgPath == "" {
// 		// 	kvs[snakeCase(m.Name)] = Func(m.Func)
// 		// }
// 	}
// }