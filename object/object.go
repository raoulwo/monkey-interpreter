package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

type ObjectType string
type BuiltinFunction func(args ...Object) Object

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
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

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType { return FUNCTION_OBJ }
func (fn *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, param := range fn.Parameters {
		params = append(params, param.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType { return STRING_OBJ }
func (str *String) Inspect() string  { return str.Value }

type Builtin struct {
	Fn BuiltinFunction
}

func (builtin *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (builtin *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType { return ARRAY_OBJ }
func (arr *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, element := range arr.Elements {
		elements = append(elements, element.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
