// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// =============================================================================
// traverseFlow — simple linear flow
// =============================================================================

func TestTraverseFlow_LinearSequence(t *testing.T) {
	e := newTestExecutor()

	// start -> create -> commit -> end
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("create"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("create")},
			Action: &microflows.CreateObjectAction{
				EntityQualifiedName: "Mod.Entity",
				OutputVariable:      "Obj",
			},
		},
		mkID("commit"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("commit")},
			Action:       &microflows.CommitObjectsAction{CommitVariable: "Obj"},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"):  {mkFlow("start", "create")},
		mkID("create"): {mkFlow("create", "commit")},
		mkID("commit"): {mkFlow("commit", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 1, nil, 0, nil)

	// StartEvent produces no output, EndEvent with no return produces no output.
	// Each activity now has a @position line before it.
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
	}
	assertContains(t, lines[0], "@position(0, 0)")
	assertContains(t, lines[1], "$Obj = create Mod.Entity;")
	assertContains(t, lines[2], "@position(0, 0)")
	assertContains(t, lines[3], "commit $Obj;")
}

// =============================================================================
// traverseFlow — IF/ELSE branching
// =============================================================================

func TestTraverseFlow_IfElse(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$x > 0"},
		},
		mkID("true_act"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("true_act")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "positive"}}},
		},
		mkID("false_act"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("false_act")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "negative"}}},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "split")},
		mkID("split"): {
			mkBranchFlow("split", "true_act", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("split", "false_act", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("true_act"):  {mkFlow("true_act", "merge")},
		mkID("false_act"): {mkFlow("false_act", "merge")},
		mkID("merge"):     {mkFlow("merge", "end")},
	}

	splitMergeMap := map[model.ID]model.ID{
		mkID("split"): mkID("merge"),
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 1, nil, 0, nil)

	// Should produce: IF, true-body, ELSE, false-body, END IF
	foundIF := false
	foundELSE := false
	foundENDIF := false
	for _, line := range lines {
		if contains(line, "if $x > 0 then") {
			foundIF = true
		}
		if contains(line, "else") {
			foundELSE = true
		}
		if contains(line, "end if;") {
			foundENDIF = true
		}
	}
	if !foundIF {
		t.Errorf("expected if statement in output: %v", lines)
	}
	if !foundELSE {
		t.Errorf("expected else in output: %v", lines)
	}
	if !foundENDIF {
		t.Errorf("expected end if in output: %v", lines)
	}
}

// =============================================================================
// collectErrorHandlerStatements
// =============================================================================

func TestCollectErrorHandlerStatements_Simple(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("err_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("err_log")},
			Action: &microflows.LogMessageAction{
				LogLevel:    "Error",
				LogNodeName: "'App'",
				MessageTemplate: &model.Text{
					Translations: map[string]string{"en_US": "Something failed"},
				},
			},
		},
		mkID("err_end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("err_end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("err_log"): {mkFlow("err_log", "err_end")},
	}

	stmts := e.collectErrorHandlerStatements(mkID("err_log"), activityMap, flowsByOrigin, nil, nil)
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d: %v", len(stmts), stmts)
	}
	assertContains(t, stmts[0], "log error")
	assertContains(t, stmts[0], "Something failed")
}

func TestCollectErrorHandlerStatements_StopsAtMerge(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("err_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("err_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Error", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "err"}}},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("after"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after"}}},
		},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("err_log"): {mkFlow("err_log", "merge")},
		mkID("merge"):   {mkFlow("merge", "after")},
	}

	stmts := e.collectErrorHandlerStatements(mkID("err_log"), activityMap, flowsByOrigin, nil, nil)
	// Should stop at merge, not include "after"
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement (stop at merge), got %d: %v", len(stmts), stmts)
	}
}

func TestCollectErrorHandlerStatements_EmptyID(t *testing.T) {
	e := newTestExecutor()
	stmts := e.collectErrorHandlerStatements("", nil, nil, nil, nil)
	if len(stmts) != 0 {
		t.Errorf("expected 0 statements for empty ID, got %d", len(stmts))
	}
}

// =============================================================================
// traverseFlow — skips merge points
// =============================================================================

func TestTraverseFlow_SkipsMergePoint(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("merge"), activityMap, nil, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	if len(lines) != 0 {
		t.Errorf("expected no output for merge point, got %v", lines)
	}
}

func TestTraverseFlow_EmptyID(t *testing.T) {
	e := newTestExecutor()
	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow("", nil, nil, nil, visited, nil, nil, &lines, 0, nil, 0, nil)
	if len(lines) != 0 {
		t.Errorf("expected no output for empty ID")
	}
}

func TestTraverseFlow_AlreadyVisited(t *testing.T) {
	e := newTestExecutor()
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("a"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("a")},
			Action:       &microflows.DeleteObjectAction{DeleteVariable: "X"},
		},
	}
	var lines []string
	visited := map[model.ID]bool{mkID("a"): true}
	e.traverseFlow(mkID("a"), activityMap, nil, nil, visited, nil, nil, &lines, 0, nil, 0, nil)
	if len(lines) != 0 {
		t.Errorf("expected no output for already visited node")
	}
}

// =============================================================================
// traverseFlowWithSourceMap — verifies source map recording
// =============================================================================

func TestTraverseFlowWithSourceMap_RecordsRange(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("act"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("act")},
			Action:       &microflows.DeleteObjectAction{DeleteVariable: "X"},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("act"): {mkFlow("act", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	sourceMap := make(map[string]elkSourceRange)

	e.traverseFlow(mkID("act"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, sourceMap, 5, nil)

	entry, ok := sourceMap["node-act"]
	if !ok {
		t.Fatal("expected source map entry for node-act")
	}
	if entry.StartLine != 5 {
		t.Errorf("expected StartLine=5, got %d", entry.StartLine)
	}
	if entry.EndLine != 6 {
		t.Errorf("expected EndLine=6, got %d", entry.EndLine)
	}
}
