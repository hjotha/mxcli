// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/visitor"
)

func TestValidateMicroflow_InheritanceSplitAllBranchesReturn(t *testing.T) {
	input := `create or modify microflow Test.SplitReturns ($ListContext: Test.Context)
returns List of Test.Item
begin
  split type $ListContext
  case Test.SpecialContext
    return empty;
  else
    return empty;
  end split;
end;`

	prog, errs := visitor.Build(input)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	mf, ok := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if !ok {
		t.Fatalf("Expected CreateMicroflowStmt, got %T", prog.Statements[0])
	}

	for _, violation := range ValidateMicroflow(mf) {
		if violation.RuleID == "MDL003" {
			t.Fatalf("unexpected missing-return violation: %#v", violation)
		}
	}
}

func TestValidateMicroflow_InfiniteWhileWithContinueDoesNotRequireTrailingReturn(t *testing.T) {
	input := `create or modify microflow Test.RetryUntilReturn ()
returns Boolean
begin
  declare $Iteration Integer = 1;
  while true
  begin
    if $Iteration > 5 then
      return false;
    end if;
    set $Iteration = $Iteration + 1;
    continue;
  end while;
end;`

	prog, errs := visitor.Build(input)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	mf, ok := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if !ok {
		t.Fatalf("Expected CreateMicroflowStmt, got %T", prog.Statements[0])
	}

	for _, violation := range ValidateMicroflow(mf) {
		if violation.RuleID == "MDL003" {
			t.Fatalf("unexpected missing-return violation: %#v", violation)
		}
	}
}
