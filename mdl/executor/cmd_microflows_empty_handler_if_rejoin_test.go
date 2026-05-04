// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// TestEmptyErrorHandlerFollowedByBothReturnIfRejoinsAtIfSplit reproduces the
// shape where a call action with an empty custom error handler (
// `$Var = call microflow ... on error without rollback { };`) is followed
// by an IF that tests $Var and whose branches both RETURN. Before the fix
// the builder produced a phantom EndEvent at the microflow tail because
// the pending error-handler rejoin skipped past the IF (the IF references
// the skipVar so the heuristic refused to reuse it), then
// terminatePendingErrorHandlersAtEnd created its own terminator.
//
// The fix: when the statement that references the skipVar is itself an
// IfStmt, rejoin the error-handler edge directly at the IF's split. The
// IF's condition already handles the empty/error case in its fallback
// branch, so re-using the split matches the Studio Pro source graph.
func TestEmptyErrorHandlerFollowedByBothReturnIfRejoinsAtIfSplit(t *testing.T) {
	fb := &flowBuilder{
		spacing:      HorizontalSpacing,
		measurer:     &layoutMeasurer{},
		declaredVars: map[string]string{"R": "String"},
	}

	oc := fb.buildFlowGraph([]ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName:  ast.QualifiedName{Module: "M", Name: "Helper"},
			OutputVariable: "R",
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: nil,
			},
		},
		&ast.IfStmt{
			Condition: &ast.VariableExpr{Name: "R"},
			ThenBody:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}, nil)

	endEvents := 0
	for _, obj := range oc.Objects {
		if _, ok := obj.(*microflows.EndEvent); ok {
			endEvents++
		}
	}
	// Expected: 2 EndEvents (one per IF branch). The phantom
	// terminatePendingErrorHandlersAtEnd EndEvent should NOT be created.
	if endEvents != 2 {
		t.Errorf("got %d EndEvents, want 2 — phantom EndEvent from empty handler at microflow end", endEvents)
	}

	// The error-handler flow from the call activity must reach the IF
	// split (possibly through a rejoin merge), not a standalone EndEvent.
	var callID, ifSplitID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			if _, ok := o.Action.(*microflows.MicroflowCallAction); ok {
				callID = o.ID
			}
		case *microflows.ExclusiveSplit:
			ifSplitID = o.ID
		}
	}
	if callID == "" || ifSplitID == "" {
		t.Fatal("missing call activity or if split")
	}
	// Walk the error flow from callID; it should reach ifSplitID through
	// zero or more ExclusiveMerge hops.
	errorFlows := map[model.ID][]*microflows.SequenceFlow{}
	for _, f := range oc.Flows {
		errorFlows[f.OriginID] = append(errorFlows[f.OriginID], f)
	}
	reachesIf := false
	var walk func(id model.ID, depth int)
	walk = func(id model.ID, depth int) {
		if depth > 5 {
			return
		}
		if id == ifSplitID {
			reachesIf = true
			return
		}
		for _, f := range errorFlows[id] {
			walk(f.DestinationID, depth+1)
		}
	}
	for _, f := range errorFlows[callID] {
		if !f.IsErrorHandler {
			continue
		}
		walk(f.DestinationID, 0)
	}
	if !reachesIf {
		t.Error("error-handler edge from call did not reach the IF split")
	}
}
