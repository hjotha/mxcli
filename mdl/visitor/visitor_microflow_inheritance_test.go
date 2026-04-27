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

func TestMicroflowParsing_EnumSplit(t *testing.T) {
	input := `CREATE MICROFLOW MyModule.DispatchEvent ()
BEGIN
  split enum $Event/EventType
  case CREATE, UPDATE
  case DELETE
    log info 'delete'
  case (empty)
    log info 'empty'
  end split;
END;`

	prog, errs := Build(input)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	mf, ok := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if !ok {
		t.Fatalf("statement: got %T, want *ast.CreateMicroflowStmt", prog.Statements[0])
	}
	if len(mf.Body) != 1 {
		t.Fatalf("body statements: got %d, want 1", len(mf.Body))
	}
	split, ok := mf.Body[0].(*ast.EnumSplitStmt)
	if !ok {
		t.Fatalf("body statement: got %T, want *ast.EnumSplitStmt", mf.Body[0])
	}
	if split.Variable != "Event/EventType" {
		t.Fatalf("split variable: got %q, want Event/EventType", split.Variable)
	}
	if len(split.Cases) != 3 {
		t.Fatalf("case count: got %d, want 3", len(split.Cases))
	}
	if split.Cases[0].Value != "CREATE" || len(split.Cases[0].Values) != 2 || split.Cases[0].Values[1] != "UPDATE" || len(split.Cases[0].Body) != 0 {
		t.Fatalf("first case: got value=%q values=%v body=%d", split.Cases[0].Value, split.Cases[0].Values, len(split.Cases[0].Body))
	}
	if split.Cases[2].Value != "(empty)" || len(split.Cases[2].Body) != 1 {
		t.Fatalf("empty case: got value=%q body=%d", split.Cases[2].Value, len(split.Cases[2].Body))
	}
}
