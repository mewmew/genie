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
	case *metadata.DICompositeType:
		return typeFromDICompositeType(t)
	case *metadata.DIDerivedType:
		return typeFromDIDerivedType(t)
	case *metadata.DISubroutineType:
		return typeFromDISubroutineType(t)
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

// typeFromDICompositeType returns the C type corresponding to the given LLVM IR
// metadata composite type.
func typeFromDICompositeType(t *metadata.DICompositeType) ctype.Type {
	switch t.Tag {
	case enum.DwarfTagEnumerationType:
		return typeFromDIEnumType(t)
	case enum.DwarfTagStructureType:
		return typeFromDIStructType(t)
	default:
		panic(fmt.Errorf("support for tag %v not yet implemented", t.Tag))
	}
}

// typeFromDIEnumType returns the C type corresponding to the given LLVM IR
// metadata enumerate type.
func typeFromDIEnumType(t *metadata.DICompositeType) ctype.Type {
	return &ctype.EnumType{
		Name: t.Name,
	}
}

// typeFromDIStructType returns the C type corresponding to the given LLVM IR
// metadata structure type.
func typeFromDIStructType(t *metadata.DICompositeType) ctype.Type {
	return &ctype.StructType{
		Name: t.Name,
	}
}

// typeFromDIDerivedType returns the C type corresponding to the given LLVM IR
// metadata derived type.
func typeFromDIDerivedType(t *metadata.DIDerivedType) ctype.Type {
	switch t.Tag {
	case enum.DwarfTagConstType:
		return typeFromDIConstType(t)
	case enum.DwarfTagPointerType:
		return typeFromDIPointerType(t)
	case enum.DwarfTagTypedef:
		return typeFromDITypedef(t)
	default:
		panic(fmt.Errorf("support for tag %v not yet implemented", t.Tag))
	}
}

// typeFromDIConstType returns the C type corresponding to the given LLVM IR
// metadata constant type.
func typeFromDIConstType(t *metadata.DIDerivedType) ctype.Type {
	return &ctype.ConstType{
		Typ: TypeFromField(t.BaseType),
	}
}

// typeFromDIPointerType returns the C type corresponding to the given LLVM IR
// metadata pointer type.
func typeFromDIPointerType(t *metadata.DIDerivedType) ctype.Type {
	return &ctype.PointerType{
		Elem: TypeFromField(t.BaseType),
	}
}

// typeFromDITypedef returns the C type corresponding to the given LLVM IR
// metadata type definition.
func typeFromDITypedef(t *metadata.DIDerivedType) ctype.Type {
	return &ctype.Typedef{
		Name: t.Name,
		Typ:  TypeFromField(t.BaseType),
	}
}

// typeFromDISubroutineType returns the C type corresponding to the given LLVM
// IR metadata subroutine type.
func typeFromDISubroutineType(t *metadata.DISubroutineType) ctype.Type {
	// TODO: parse t.CC.
	var paramTypes []ctype.Type
	retType := TypeFromField(t.Types.Fields[0])
	for _, field := range t.Types.Fields[1:] {
		paramType := TypeFromField(field)
		paramTypes = append(paramTypes, paramType)
	}
	return &ctype.FuncType{
		RetType:    retType,
		ParamTypes: paramTypes,
	}
}
