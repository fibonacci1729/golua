package lua

// TODO: comment
func CheckCallable(val Value) (Callable, error) {
	if v := AsCallable(val); v != nil {
		return v, nil
	}
	return nil, &TypeError{Expect: "function", Value: val}
}

// TODO: comment
func CheckConstant(val Value) (Constant, error) {
	if v := AsConstant(val); v != nil {
		return v, nil
	}
	return nil, &TypeError{Expect: "constant", Value: val}
}

// TODO: comment
func CheckGoValue(val Value) (*GoValue, error) {
	if v := AsGoValue(val); v != nil {
		return v, nil
	}
	return nil, &TypeError{Expect: "userdata", Value: val}
}

// TODO: comment
func CheckGoFunc(val Value) (*GoFunc, error) {
	if fn := AsGoFunc(val); fn != nil {
		return fn, nil
	}
	return nil, &TypeError{Expect: "function", Value: val}
}

// TODO: comment
func CheckThread(val Value) (*Thread, error) {
	if thread := AsThread(val); thread != nil {
		return thread, nil
	}
	return nil, &TypeError{Expect: "thread", Value: val}
}

// TODO: comment
func CheckNumber(val Value) (Number, error) {
	if num := AsNumber(val); num != nil {
		return num, nil
	}
	return nil, &TypeError{Expect: "number", Value: val}
}

// TODO: comment
func CheckString(val Value) (String, error) {
	if str, ok := AsString(val); ok {
		return str, nil
	}
	return "", &TypeError{Expect: "string", Value: val}
}

// TODO: comment
func CheckTable(val Value) (*Table, error) {
	if table := AsTable(val); table != nil {
		return table, nil
	}
	return nil, &TypeError{Expect: "table", Value: val}
}

// TODO: comment
func CheckFunc(val Value) (Callable, error) {
	if v := AsCallable(val); v != nil {
		return v, nil
	}
	return nil, &TypeError{Expect: "function", Value: val}
}

// TODO: comment
func CheckFloat(val Value) (Float, error) {
	if f64, ok := AsFloat(val); ok {
		return f64, nil
	}
	return 0, &TypeError{Expect: "float", Value: val}
}

// TODO: comment
func CheckBool(val Value) (Bool, error) {
	if IsBool(val) {
		return Bool(Truth(val)), nil
	}
	return false, &TypeError{Expect: "boolean", Value: val}
}

// TODO: comment
func CheckInt(val Value) (Int, error) {
	if i64, ok := AsInt(val); ok {
		return i64, nil
	}
	return 0, &TypeError{Expect: "integer", Value: val}
}

//
// Helpers
//

// // TODO: comment
// func ToString(ls *Thread, v Value) (String, error) {
// 	if v, ok, err := ls.TypeOf(v).Invoke(ls, "__tostring"); ok {
// 		if err != nil {
// 			return "", err
// 		}
// 		if IsString(v) {
// 			return v.(String), nil
// 		}
// 		return "", fmt.Errorf("'__tostring' must return a string")
// 	}
// 	switch v := v.(type) {
// 		case String, Float, Bool, Int:
// 			return String(v.String()), nil
// 		case *Table:
// 			if v.meta == nil || v.meta.Get(String("__name")) == nil {
// 				return String(v.String()), nil
// 			}
// 		case nil:
// 			return String("nil"), nil
// 	}
// 	return String(ls.TypeOf(v).Name()), nil
// }