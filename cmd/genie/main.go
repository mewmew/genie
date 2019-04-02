package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"
	"text/template"

	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/metadata"
	"github.com/llir/llvm/ir/value"
	"github.com/mewkiz/pkg/goutil"
	"github.com/mewmew/genie/ctype"
	"github.com/mewmew/genie/mdutil"
	"github.com/mewmew/pe"
	"github.com/pkg/errors"
)

func usage() {
	const use = `
Usage: genie [OPTION]... FILE.ll...
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	var (
		// Path to original PE binary executable.
		origPath string
		// Output path of C source code.
		output string
	)
	flag.StringVar(&origPath, "orig", "orig.exe", "path to original PE binary executable")
	flag.StringVar(&output, "o", "", "output path of C source code (default stdout)")
	flag.Usage = usage
	flag.Parse()
	llPaths := flag.Args()
	for _, llPath := range llPaths {
		if err := genie(llPath, origPath, output); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// genie converts the given LLVM IR assembly file into a Go package containing
// the same exported functions.
func genie(llPath, origPath, output string) error {
	m, err := asm.ParseFile(llPath)
	if err != nil {
		return errors.WithStack(err)
	}
	file, err := pe.ParseFile(origPath)
	if err != nil {
		return errors.WithStack(err)
	}
	w := os.Stdout
	if len(output) > 0 {
		fd, err := os.Create(output)
		if err != nil {
			return errors.WithStack(err)
		}
		defer fd.Close()
		w = fd
	}
	const preface = `
#include "export.h"
`
	fmt.Fprintln(w, preface[1:])
	for _, f := range m.Funcs {
		if len(f.Blocks) == 0 {
			continue
		}
		locals := mdutil.LocalVars(f)
		addr, err := parseAddr(f, locals)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := printFunc(w, f, addr, locals, file); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// printFunc outputs the given function in Go syntax, writing to w. The content
// of the original binary executable is used for patches (restoring the original
// assembly instructions that were overwritten by the injected jmp instruction).
func printFunc(w io.Writer, f *ir.Func, addr uint64, locals []mdutil.Var, file *pe.File) error {
	// Get return type.
	m := make(map[string]mdutil.Var)
	for _, local := range locals {
		m[local.LLVarName] = local
	}
	retType, err := parseRetType(f)
	if err != nil {
		return errors.WithStack(err)
	}

	// Get calling convention.
	callConv := cCallConv(f.CallingConv)

	// Get function name.
	funcName := f.Name()

	// Get params.
	var params []mdutil.Var
	for _, param := range f.Params {
		// Look for store instructions in the entry basic block, used to store
		// function paramters in stack-allocated local variables.
		localName, err := findParamName(f, param.Name())
		if err != nil {
			return errors.WithStack(err)
		}
		local, ok := m[localName]
		if !ok {
			panic(fmt.Errorf("unable to locate debug info of local %q in function %q", localName, f.Name()))
		}
		params = append(params, local)
	}

	// Output using template.
	funcs := template.FuncMap{
		"verb": verbFromCType,
	}
	srcDir, err := goutil.SrcDir("github.com/mewmew/genie/cmd/genie")
	if err != nil {
		return errors.WithStack(err)
	}
	const tmplName = "export.tmpl"
	tmplPath := filepath.Join(srcDir, tmplName)
	t, err := template.New(tmplName).Funcs(funcs).ParseFiles(tmplPath)
	if err != nil {
		return errors.WithStack(err)
	}
	// Size of injected jmp instruction in number of bytes.
	const patchSize = 5
	orig := file.ReadData(addr, patchSize)
	tw := tabwriter.NewWriter(w, 1, 3, 1, ' ', tabwriter.TabIndent)
	hasRet := true
	if rt, ok := retType.(ctype.BasicType); ok && rt == ctype.BasicTypeVoid {
		hasRet = false
	}
	data := map[string]interface{}{
		"RetType":   retType,
		"CallConv":  callConv,
		"FuncName":  funcName,
		"Params":    params,
		"Orig":      orig,
		"PatchSize": patchSize,
		"Addr":      addr,
		"HasRet":    hasRet,
	}
	if err := t.Execute(tw, data); err != nil {
		return errors.WithStack(err)
	}
	if err := tw.Flush(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// verbFromCType returns the format string verb corresponding to the given C
// type.
func verbFromCType(t ctype.Type) string {
	switch t := t.(type) {
	case ctype.BasicType:
		switch t {
		case ctype.BasicTypeChar, ctype.BasicTypeSChar, ctype.BasicTypeUChar, ctype.BasicTypeShort, ctype.BasicTypeShortInt, ctype.BasicTypeSShort, ctype.BasicTypeSShortInt, ctype.BasicTypeUShort, ctype.BasicTypeUShortInt, ctype.BasicTypeInt, ctype.BasicTypeSigned, ctype.BasicTypeSInt, ctype.BasicTypeUnsigned, ctype.BasicTypeUInt, ctype.BasicTypeLong, ctype.BasicTypeLongInt, ctype.BasicTypeSLong, ctype.BasicTypeSLongInt, ctype.BasicTypeULong, ctype.BasicTypeULongInt:
			return "%d"
		case ctype.BasicTypeLongLong, ctype.BasicTypeLongLongInt, ctype.BasicTypeSLongLong, ctype.BasicTypeSLongLongInt, ctype.BasicTypeULongLong, ctype.BasicTypeULongLongInt:
			return "%ld"
		case ctype.BasicTypeFloat, ctype.BasicTypeDouble, ctype.BasicTypeLongDouble:
			return "%f"
		default:
			panic(fmt.Errorf("support for basic type %v not yet implemented", uint(t)))
		}
	case *ctype.PointerType:
		if bt, ok := t.Elem.(ctype.BasicType); ok && bt == ctype.BasicTypeChar {
			return "%s"
		}
		return "%p"
	case *ctype.Typedef:
		return verbFromCType(t.Typ)
	default:
		panic(fmt.Errorf("support for type %T not yet implemented", t))
	}
}

// findLocalVarOfParam returns the name of the stack-allocated local variable
// (alloca) corresponding to the given function parameter.
func findParamName(f *ir.Func, paramName string) (string, error) {
	// Clang allocates a local variables to store each function parameter in.
	entry := f.Blocks[0]
	for _, inst := range entry.Insts {
		storeInst, ok := inst.(*ir.InstStore)
		if !ok {
			continue
		}
		src, ok := storeInst.Src.(value.Named)
		if !ok {
			continue
		}
		if src.Name() == paramName {
			dst := storeInst.Dst.(value.Named)
			return dst.Name(), nil
		}
	}
	return "", errors.Errorf("unable to locate name of stack-allocated local variable corresponding to function parameter %q in function %q", paramName, f.Name())
}

// cCallConv returns the C calling convention corresponding to the given LLVM IR
// calling convention.
func cCallConv(callConv enum.CallingConv) string {
	switch callConv {
	case enum.CallingConvNone:
		return ""
	case enum.CallingConvX86StdCall:
		return "__stdcall"
	case enum.CallingConvX86FastCall:
		return "__fastcall"
	default:
		panic(fmt.Errorf("support for calling convention %v not yet implemented", callConv))
	}
}

// parseAddr parses the address of the given function. The address is stored in
// the 'addr' variable.
func parseAddr(f *ir.Func, locals []mdutil.Var) (uint64, error) {
	if len(f.Blocks) != 1 {
		return 0, errors.Errorf("invalid number of basic blocks in %q; expected 1, got %d", f.Name(), len(f.Blocks))
	}
	entry := f.Blocks[0]
	// Locate LLVM IR local variable corresponding to the C variable `addr`.
	// maps from C variable name to LLVM IR variable.
	cNameToVar := make(map[string]mdutil.Var)
	for _, local := range locals {
		cNameToVar[local.CVarName] = local
	}
	addrLocal, ok := cNameToVar["addr"]
	if !ok {
		return 0, errors.Errorf("unable to locate LLVM IR local variable corresponding to C variable `addr` in function %q", f.Name())
	}
	varName := addrLocal.LLVarName
	// Locate store instruction, storing the function address to the `addr`
	// variable.
	for _, inst := range entry.Insts {
		storeInst, ok := inst.(*ir.InstStore)
		if !ok {
			continue
		}
		dst, ok := storeInst.Dst.(*ir.InstAlloca)
		if !ok {
			continue
		}
		if dst.Name() != varName {
			continue
		}
		// Store instruction located, which stores the function address to the
		// `addr` variable.
		v, ok := storeInst.Src.(*constant.Int)
		if !ok {
			return 0, errors.Errorf("addr constant type mismatch; expected *constant.Int, got %T", storeInst.Src)
		}
		addr := v.X.Uint64()
		return addr, nil
	}
	return 0, errors.Errorf("unable to locate `store` instruction of `addr` variable in function %q", f.Name())
}

// parseRetType parses the return type of a given function based on its attached
// metadata.
func parseRetType(f *ir.Func) (ctype.Type, error) {
	for _, md := range f.MDAttachments() {
		diSub, ok := md.Node.(*metadata.DISubprogram)
		if !ok {
			continue
		}
		diSubType, ok := diSub.Type.(*metadata.DISubroutineType)
		if !ok {
			continue
		}
		// Parse return type.
		switch field := diSubType.Types.Fields[0].(type) {
		case *metadata.NullLit:
			return ctype.BasicTypeVoid, nil
		case metadata.Field:
			return mdutil.TypeFromField(field), nil
		default:
			panic(fmt.Errorf("support for metadata field type %T not yet implemented", field))
		}
	}
	return nil, errors.Errorf("unable to locate return type of function %q", f.Name())
}
