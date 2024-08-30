package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
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
	HASH_OBJ         = "HASH"
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
func (integer *Integer) HashKey() HashKey {
	return HashKey{Type: integer.Type(), Value: uint64(integer.Value)}
}

type Boolean struct {
	Value bool
}

func (boolean *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (boolean *Boolean) Inspect() string  { return fmt.Sprintf("%t", boolean.Value) }
func (boolean *Boolean) HashKey() HashKey {
	var value uint64

	if boolean.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: boolean.Type(), Value: value}
}

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
func (str *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(str.Value))

	return HashKey{Type: str.Type(), Value: h.Sum64()}
}

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

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (hash *Hash) Type() ObjectType { return HASH_OBJ }
func (hash *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range hash.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
