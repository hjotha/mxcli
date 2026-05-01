// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/visitor"
)

func TestValidateMicroflowReferencesAddExpressionValue(t *testing.T) {
	input := `create microflow Synthetic.MF_AddExpressionScope ()
returns Boolean
begin
  declare $Items List of Synthetic.Item = empty;
  if true then
    declare $SourceItems List of Synthetic.Item = empty;
  end if;
  add head($SourceItems) to $Items;
  return true;
end;`

	prog, errs := visitor.Build(input)
	if len(errs) > 0 {
		t.Fatalf("Parse error: %v", errs[0])
	}
	stmt := prog.Statements[0].(*ast.CreateMicroflowStmt)

	violations := ValidateMicroflow(stmt)
	for _, violation := range violations {
		if violation.RuleID == "MDL005" && strings.Contains(violation.Message, "$SourceItems") {
			return
		}
	}
	t.Fatalf("Expected MDL005 for add expression source variable, got %#v", violations)
}
