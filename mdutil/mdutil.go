// Package mdutil provides LLVM IR metadata utility functions.
package mdutil

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/metadata"
	"github.com/llir/llvm/ir/value"
	"github.com/mewmew/genie/ctype"
)

// Var maps an LLVM IR local variable to its corresponding C variable and type
// information, as based on metadata.
type Var struct {
	// LLVM IR variable name.
	LLVarName string
	// C variable name.
	CVarName string
	// C type information.
	CType ctype.Type
}

// LocalVars returns the mapping between LLVM IR local variables and their
// corresponding C variables and type information, as based on the metadata of
// the given function.
func LocalVars(f *ir.Func) []Var {
	var locals []Var
	for _, block := range f.Blocks {
		for _, inst := range block.Insts {
			// Locate calls to @llvm.dbg.declare.
			callInst, ok := inst.(*ir.InstCall)
			if !ok {
				continue
			}
			callee, ok := callInst.Callee.(*ir.Func)
			if !ok {
				continue
			}
			if callee.Name() != "llvm.dbg.declare" {
				continue
			}
			// Locate LLVM IR variable name.
			mdVal0, ok := callInst.Args[0].(*metadata.Value)
			if !ok {
				continue
			}
			v, ok := mdVal0.Value.(value.Named)
			if !ok {
				continue
			}
			llVarName := v.Name()
			// Locate C variable name.
			mdVal1, ok := callInst.Args[1].(*metadata.Value)
			if !ok {
				continue
			}
			diVar, ok := mdVal1.Value.(*metadata.DILocalVariable)
			if !ok {
				continue
			}
			cVarName := diVar.Name
			// Locate C type information.
			cType := TypeFromField(diVar.Type)
			// Record local variable.
			local := Var{
				LLVarName: llVarName,
				CVarName:  cVarName,
				CType:     cType,
			}
			locals = append(locals, local)
		}
	}
	return locals
}
