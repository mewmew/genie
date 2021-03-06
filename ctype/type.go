// Package ctype declares the data types of C.
package ctype

import (
	"bytes"
	"fmt"
)

// === [ Type ] ================================================================

// Type is a C type.
type Type interface {
	fmt.Stringer
}

// --- [ Basic type ] ----------------------------------------------------------

// BasicType is a C basic type.
type BasicType uint

// CString returns the C syntax representation of the definition of the type.
func (t BasicType) CString() string {
	switch t {
	case BasicTypeChar:
		return fmt.Sprintf("typedef char int8_t;")
	default:
		panic(fmt.Errorf("support for basic type %v (%s) not yet implemented", uint(t), t))
	}
}

//go:generate stringer -linecomment -type BasicType

// Basic types.
//
// ref: https://en.wikipedia.org/wiki/C_data_types#Basic_types
const (
	// void type
	BasicTypeVoid BasicType = iota + 1 // void
	// signed or unsigned char
	BasicTypeChar // char
	// [-127, +127]
	BasicTypeSChar // signed char
	// [0, 255]
	BasicTypeUChar // unsigned char
	// [-32,767, +32,767]
	BasicTypeShort     // short
	BasicTypeShortInt  // short int
	BasicTypeSShort    // signed short
	BasicTypeSShortInt // signed short int
	// [0, 65,535]
	BasicTypeUShort    // unsigned short
	BasicTypeUShortInt // unsigned short int
	// [-32,767, +32,767]
	BasicTypeInt    // int
	BasicTypeSigned // signed
	BasicTypeSInt   // signed int
	// [0, 65,535]
	BasicTypeUnsigned // unsigned
	BasicTypeUInt     // unsigned int
	// [-2,147,483,647, +2,147,483,647]
	BasicTypeLong     // long
	BasicTypeLongInt  // long int
	BasicTypeSLong    // signed long
	BasicTypeSLongInt // signed long int
	// [0, 4,294,967,295]
	BasicTypeULong    // unsigned long
	BasicTypeULongInt // unsigned long int
	// [-9,223,372,036,854,775,807, +9,223,372,036,854,775,807]
	BasicTypeLongLong     // long long
	BasicTypeLongLongInt  // long long int
	BasicTypeSLongLong    // signed long long
	BasicTypeSLongLongInt // signed long long int
	// [0, +18,446,744,073,709,551,615]
	BasicTypeULongLong    // unsigned long long
	BasicTypeULongLongInt // unsigned long long int
	// IEEE 754 single-precision binary floating-point format (32 bits)
	BasicTypeFloat // float
	// IEEE 754 double-precision binary floating-point format (64 bits)
	BasicTypeDouble // double
	// IEEE 754 quadruple-precision floating-point format (128 bits)
	BasicTypeLongDouble // long double
)

// --- [ Constant type ] -------------------------------------------------------

// ConstType is a C constant type.
type ConstType struct {
	// Underlying type.
	Typ Type
}

// String returns the C syntax representation of the type.
func (t *ConstType) String() string {
	return fmt.Sprintf("const %v", t.Typ.String())
}

// --- [ Pointer type ] --------------------------------------------------------

// PointerType is a C pointer type.
type PointerType struct {
	// Element type.
	Elem Type
}

// String returns the C syntax representation of the type.
func (t *PointerType) String() string {
	// TODO: use "Var" hack to handle spiral rule?
	return fmt.Sprintf("%v *", t.Elem.String())
}

// --- [ Enum type ] --------------------------------------------------------

// EnumType is a C enumerate type.
type EnumType struct {
	// Enum name (tag).
	Name string
	// TODO: add Enum fields.
}

// String returns the C syntax representation of the type.
func (t *EnumType) String() string {
	return t.Name
}

// --- [ Struct type ] --------------------------------------------------------

// StructType is a C structure type.
type StructType struct {
	// Struct name (tag).
	Name string
	// TODO: add struct fields.
}

// String returns the C syntax representation of the type.
func (t *StructType) String() string {
	return t.Name
}

// --- [ Type definition ] -----------------------------------------------------

// Typedef is a C type definition.
type Typedef struct {
	// Type name.
	Name string
	// Underlying type definition.
	Typ Type
}

// String returns the C syntax representation of the type.
func (t *Typedef) String() string {
	return t.Name
}

// CString returns the C syntax representation of definition of the type.
func (t *Typedef) CString() string {
	// TODO: use "Var" hack to handle spiral rule?
	return fmt.Sprintf("typedef %s %s;", t.Name, t.Typ)
}

// --- [ Function type ] -------------------------------------------------------

// FuncType is a C function type.
type FuncType struct {
	// Return type.
	RetType Type
	// Callee name (used for printing type-ident pairs).
	Callee string
	// Calling convention.
	CallConv CallingConv
	// Parameter types.
	ParamTypes []Type
}

// String returns the C syntax representation of the type.
func (t *FuncType) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(t.RetType.String())
	fmt.Fprintf(buf, " (*%s)", t.Callee)
	buf.WriteString("(")
	for i, param := range t.ParamTypes {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(param.String())
	}
	buf.WriteString(")")
	return buf.String()
}

//go:generate stringer -linecomment -type CallingConv

// CallingConv represents a function calling convention.
type CallingConv uint8

// Calling conventions.
const (
	CallConvFastCall CallingConv = iota + 1 // __fastcall
	CallConvStdCall                         // __stdcall
)
