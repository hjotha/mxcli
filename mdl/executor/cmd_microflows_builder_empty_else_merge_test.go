// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// TestIfEmptyElseBodyWithContinuingThenEmitsFalseFlowToMerge is a regression
// guard for CE0079 ("The 'false' condition value should be configured in
// properties for an outgoing flow").
//
// Pattern: an IF with THEN that continues (non-terminating) and an explicitly
// empty ELSE body. Both branches must feed a merge — but when the ELSE body is
// empty, there is no lastElseID to wire up, so the false outgoing flow was
// dropped entirely. The fix is to emit a direct split→merge flow with case
// "false" (analogous to the empty-THEN handling).
func TestIfEmptyElseBodyWithContinuingThenEmitsFalseFlowToMerge(t *testing.T) {
	fb := &flowBuilder{
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}

	fb.addIfStatement(&ast.IfStmt{
		Condition: &ast.VariableExpr{Name: "Flag"},
		ThenBody: []ast.MicroflowStatement{
			&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "in-then"}},
		},
		// ElseBody deliberately empty (0 statements). Len-based hasElseBody is
		// false in this case, so this narrow test drives the "non-hasElseBody"
		// path. For the hasElseBody + empty body case, see below.
	})

	var split *microflows.ExclusiveSplit
	var merge *microflows.ExclusiveMerge
	for _, obj := range fb.objects {
		switch o := obj.(type) {
		case *microflows.ExclusiveSplit:
			split = o
		case *microflows.ExclusiveMerge:
			merge = o
		}
	}
	if split == nil {
		t.Fatal("expected ExclusiveSplit to be created")
	}
	if merge == nil {
		t.Fatal("expected ExclusiveMerge to be created (THEN continues → false path must converge)")
	}

	var hasFalseFlow bool
	for _, flow := range fb.flows {
		if flow.OriginID != split.ID || flow.DestinationID != merge.ID {
			continue
		}
		if ec, ok := flow.CaseValue.(microflows.EnumerationCase); ok && ec.Value == "false" {
			hasFalseFlow = true
		}
	}
	if !hasFalseFlow {
		t.Fatal("missing split→merge flow with case \"false\" — CE0079 regression")
	}
}
