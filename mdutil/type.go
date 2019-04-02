package mdutil

import (
	"fmt"

	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/metadata"
	"github.com/mewmew/genie/ctype"
)

// TypeFromField returns the C type corresponding to the given LLVM IR metadata
// type.
func TypeFromField(t metadata.Field) ctype.Type {
	switch t := t.(type) {
	case *metadata.DIBasicType:
		return typeFromDIBasicType(t)
	case *metadata.DIDerivedType:
		return typeFromDIDerivedType(t)
	case *metadata.NullLit:
		return ctype.BasicTypeVoid
	default:
		panic(fmt.Errorf("support for type %T not yet implemented", t))
	}
}

// typeFromDIBasicType returns the C type corresponding to the given LLVM IR
// metadata derived type.
func typeFromDIBasicType(t *metadata.DIBasicType) ctype.Type {
	return BasicTypeFromString(t.Name)
}

// typeFromDIDerivedType returns the C type corresponding to the given LLVM IR
// metadata derived type.
func typeFromDIDerivedType(t *metadata.DIDerivedType) ctype.Type {
	switch t.Tag {
	case enum.DwarfTagPointerType:
		return typeFromDIPointerType(t)
	case enum.DwarfTagTypedef:
		return typeFromDITypedefTo(t)
	default:
		panic(fmt.Errorf("support for tag %v not yet implemented", t.Tag))
	}
}

// typeFromDIPointerType returns the C type corresponding to the given LLVM IR
// metadata pointer type.
func typeFromDIPointerType(t *metadata.DIDerivedType) ctype.Type {
	return &ctype.PointerType{
		Elem: TypeFromField(t.BaseType),
	}
}

// typeFromDITypedefTo returns the C type corresponding to the given LLVM IR
// metadata pointer type.
func typeFromDITypedefTo(t *metadata.DIDerivedType) ctype.Type {
	return &ctype.Typedef{
		Name: t.Name,
		Typ:  TypeFromField(t.BaseType),
	}
}
