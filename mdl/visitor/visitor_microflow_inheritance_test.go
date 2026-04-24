// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestMicroflowParsing_InheritanceSplitAndCastAction(t *testing.T) {
	input := `CREATE MICROFLOW MyModule.CastObject ()
BEGIN
  split type $Input;
  cast $Specific;
END;`

	prog, errs := Build(input)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("statements: got %d, want 1", len(prog.Statements))
	}
	mf, ok := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if !ok {
		t.Fatalf("statement: got %T, want *ast.CreateMicroflowStmt", prog.Statements[0])
	}
	if len(mf.Body) != 2 {
		t.Fatalf("body statements: got %d, want 2", len(mf.Body))
	}

	split, ok := mf.Body[0].(*ast.InheritanceSplitStmt)
	if !ok {
		t.Fatalf("first body statement: got %T, want *ast.InheritanceSplitStmt", mf.Body[0])
	}
	if split.Variable != "Input" {
		t.Fatalf("split variable: got %q, want Input", split.Variable)
	}

	cast, ok := mf.Body[1].(*ast.CastObjectStmt)
	if !ok {
		t.Fatalf("second body statement: got %T, want *ast.CastObjectStmt", mf.Body[1])
	}
	if cast.OutputVariable != "Specific" || cast.ObjectVariable != "" {
		t.Fatalf("cast vars: got output=%q object=%q", cast.OutputVariable, cast.ObjectVariable)
	}
}
