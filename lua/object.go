package lua

import (
	"fmt"
)

// TODO: comment
var noValueErr = fmt.Errorf("value expected")

// TODO: comment
type Object interface {
	Callable() (Callable, error)
	Constant() (Constant, error)
	GoValue() (*GoValue, error)
	GoFunc() (*GoFunc, error)
	Thread() (*Thread, error)
	String() (String, error)
	Number() (Number, error)
	Table() (*Table, error)
	Float() (Float, error)
	Bool() (Bool, error)
	Int() (Int, error)
	Any() (Value, error)
	IsType(string) bool
	Valid() bool
	Value() Value
}

// TODO: comment
type object struct {
	value *Value
	frame frame
	index int
}

// TODO: comment
func (o *object) Callable() (Callable, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckCallable(*o.value)
}

// TODO: comment
func (o *object) Constant() (Constant, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckConstant(*o.value)
}

// TODO: comment
func (o *object) GoValue() (*GoValue, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckGoValue(*o.value)
}

// TODO: comment
func (o *object) GoFunc() (*GoFunc, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckGoFunc(*o.value)
}

// TODO: comment
func (o *object) Thread() (*Thread, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckThread(*o.value)
}

// TODO: comment
func (o *object) String() (String, error) {
	if !o.Valid() {
		return "", noValueErr
	}
	return CheckString(*o.value)
}

// TODO: comment
func (o *object) Number() (Number, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckNumber(*o.value)
}

// TODO: comment
func (o *object) Table() (*Table, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return CheckTable(*o.value)
}

// TODO: comment
func (o *object) Float() (Float, error) {
	if !o.Valid() {
		return 0, noValueErr
	}
	return CheckFloat(*o.value)
}

// TODO: comment
func (o *object) Bool() (Bool, error) {
	if !o.Valid() {
		return false, noValueErr
	}
	return CheckBool(*o.value)
}

// TODO: comment
func (o *object) Int() (Int, error) {
	if !o.Valid() {
		return 0, noValueErr
	}
	return CheckInt(*o.value)
}

// TODO: comment
func (o *object) Any() (Value, error) {
	if !o.Valid() {
		return nil, noValueErr
	}
	return *o.value, nil
}

// TODO: comment
func (o *object) IsType(name string) bool {
	if !o.Valid() {
		return false
	}
	typ := o.frame.ls.typeOf(*o.value)
	return typ.Name() == name
}

// TODO: comment
func (o *object) Valid() bool {
	return o != nil
}

// TODO: comment
func (o *object) Value() Value {
	if !o.Valid() {
		return nil
	}
	return *o.value
}