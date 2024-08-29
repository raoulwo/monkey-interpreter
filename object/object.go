package object

import "fmt"

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (integer *Integer) Type() ObjectType { return INTEGER_OBJ }
func (integer *Integer) Inspect() string  { return fmt.Sprintf("%d", integer.Value) }

type Boolean struct {
	Value bool
}

func (boolean *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (boolean *Boolean) Inspect() string  { return fmt.Sprintf("%t", boolean.Value) }

// I'm sorry Tony.
type Null struct{}

func (null *Null) Type() ObjectType { return NULL_OBJ }
func (null *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (returnValue *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (returnValue *ReturnValue) Inspect() string  { return returnValue.Value.Inspect() }

type Error struct {
	Message string
}

func (error *Error) Type() ObjectType { return ERROR_OBJ }
func (error *Error) Inspect() string  { return "ERROR: " + error.Message }

func NewEnvironment() *Environment {
	store := make(map[string]Object)
	return &Environment{store: store}
}

type Environment struct {
	store map[string]Object
}

func (env *Environment) Get(name string) (Object, bool) {
	obj, ok := env.store[name]
	return obj, ok
}

func (env *Environment) Set(name string, val Object) Object {
	env.store[name] = val
	return val
}
