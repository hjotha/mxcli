// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
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

	// StartEvent produces no output. Void EndEvent emits an explicit return.
	// Each emitted activity has a @position line before it.
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d: %v", len(lines), lines)
	}
	assertContains(t, lines[0], "@position(0, 0)")
	assertContains(t, lines[1], "$Obj = create Mod.Entity;")
	assertContains(t, lines[2], "@position(0, 0)")
	assertContains(t, lines[3], "commit $Obj;")
	assertContains(t, lines[4], "@position(0, 0)")
	assertContains(t, lines[5], "return;")
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

func TestFindMergeForSplit_ChoosesNearestMergeBeforeDownstreamIf(t *testing.T) {
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("logo_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("logo_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Logo != empty"},
		},
		mkID("logo_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("logo_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "logo"}}},
		},
		mkID("logo_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("logo_merge")},
		mkID("after_logo"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after_logo")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after logo"}}},
		},
		mkID("cover_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("cover_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Cover != empty"},
		},
		mkID("cover_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("cover_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "cover"}}},
		},
		mkID("cover_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("cover_merge")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("logo_split"): {
			mkBranchFlow("logo_split", "logo_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("logo_split", "logo_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("logo_then"):  {mkFlow("logo_then", "logo_merge")},
		mkID("logo_merge"): {mkFlow("logo_merge", "after_logo")},
		mkID("after_logo"): {mkFlow("after_logo", "cover_split")},
		mkID("cover_split"): {
			mkBranchFlow("cover_split", "cover_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("cover_split", "cover_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("cover_then"): {mkFlow("cover_then", "cover_merge")},
	}

	got := findMergeForSplit(nil, mkID("logo_split"), flowsByOrigin, activityMap)
	if got != mkID("logo_merge") {
		t.Fatalf("logo split paired with %q, want nearest merge %q", got, mkID("logo_merge"))
	}
}

// TestTraverseFlow_IfWithoutElse verifies that when the FALSE branch jumps
// straight to the merge point (as emitted by the builder for `if X then ... end if`),
// the describer does not print an empty `else` — the previous behavior produced
// `if X then ... else end if;` which re-parses as an IF with an empty else body
// that is indistinguishable from the original.
func TestTraverseFlow_IfWithoutElse(t *testing.T) {
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
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "split")},
		mkID("split"): {
			mkBranchFlow("split", "true_act", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("split", "merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("true_act"): {mkFlow("true_act", "merge")},
		mkID("merge"):    {mkFlow("merge", "end")},
	}

	splitMergeMap := map[model.ID]model.ID{
		mkID("split"): mkID("merge"),
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 1, nil, 0, nil)

	for _, line := range lines {
		if contains(line, "else") {
			t.Errorf("expected no else branch in output, got: %v", lines)
		}
	}
	foundENDIF := false
	for _, line := range lines {
		if contains(line, "end if;") {
			foundENDIF = true
		}
	}
	if !foundENDIF {
		t.Errorf("expected end if in output: %v", lines)
	}
}

func TestTraverseFlow_LoopBodyMergesParentFlowsForExistingOrigin(t *testing.T) {
	e := newTestExecutor()

	logActivity := func(id, message string, x, y int) *microflows.ActionActivity {
		return &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: mkID(id)},
					Position:    model.Point{X: x, Y: y},
				},
			},
			Action: &microflows.LogMessageAction{
				LogLevel:        "Info",
				LogNodeName:     "'Synthetic'",
				MessageTemplate: &model.Text{Translations: map[string]string{"en_US": message}},
			},
		}
	}

	nestedLoop := &microflows.LoopedActivity{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: mkID("nested")},
			Position:    model.Point{X: 500, Y: 100},
		},
		LoopSource: &microflows.IterableList{
			VariableName:     "role",
			ListVariableName: "roles",
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{
			Objects: []microflows.MicroflowObject{
				logActivity("nested-log", "nested", 120, 80),
				logActivity("nested-tail", "nested-tail", 320, 80),
			},
			Flows: []*microflows.SequenceFlow{
				// Same pattern one level deeper: the loop-boundary flow is
				// local, while the real body continuation is supplied by the
				// parent graph that must be threaded into nested loops.
				mkFlow("nested-log", "nested"),
			},
		},
	}
	outerLoop := &microflows.LoopedActivity{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: mkID("loop")},
		},
		LoopSource: &microflows.IterableList{
			VariableName:     "item",
			ListVariableName: "items",
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{
			// Adversarial order: storage lists the nested loop first, but the
			// flow graph and positions define the actual body order.
			Objects: []microflows.MicroflowObject{
				nestedLoop,
				logActivity("setup", "setup", 100, 100),
				logActivity("fetch", "fetch", 300, 100),
			},
			Flows: []*microflows.SequenceFlow{
				mkFlow("setup", "fetch"),
				// This local loop-boundary flow gives "fetch" an existing
				// local origin entry. Parent-level body flows with the same
				// origin must still be merged in.
				mkFlow("fetch", "loop"),
			},
		},
	}

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("loop"):  outerLoop,
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"):  {mkFlow("start", "loop")},
		mkID("loop"):   {mkFlow("loop", "end")},
		mkID("fetch"):  {mkFlow("fetch", "nested")},
		mkID("nested"): {mkFlow("nested", "loop")},
		mkID("nested-log"): {
			mkFlow("nested-log", "nested-tail"),
		},
		mkID("nested-tail"): {mkFlow("nested-tail", "nested")},
	}

	var lines []string
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, make(map[model.ID]bool), nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	for _, want := range []string{
		"log info node 'Synthetic' 'setup';",
		"log info node 'Synthetic' 'fetch';",
		"loop $role in $roles",
		"log info node 'Synthetic' 'nested';",
		"log info node 'Synthetic' 'nested-tail';",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output:\n%s", want, out)
		}
	}
	if strings.Index(out, "setup") > strings.Index(out, "fetch") ||
		strings.Index(out, "fetch") > strings.Index(out, "loop $role in $roles") {
		t.Fatalf("loop body emitted in wrong order:\n%s", out)
	}
}

func TestTraverseFlow_LoopBodyEmitsStructuredIfElse(t *testing.T) {
	split := &microflows.ExclusiveSplit{
		BaseMicroflowObject: mkObj("split"),
		SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Item/IsActive"},
	}
	trueLog := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("true_log")},
		Action: &microflows.LogMessageAction{
			LogLevel:    "Info",
			LogNodeName: "'App'",
			MessageTemplate: &model.Text{
				Translations: map[string]string{"en_US": "active"},
			},
		},
	}
	falseLog := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("false_log")},
		Action: &microflows.LogMessageAction{
			LogLevel:    "Info",
			LogNodeName: "'App'",
			MessageTemplate: &model.Text{
				Translations: map[string]string{"en_US": "inactive"},
			},
		},
	}
	merge := &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")}
	loop := &microflows.LoopedActivity{
		BaseMicroflowObject: mkObj("loop"),
		ObjectCollection: &microflows.MicroflowObjectCollection{
			Objects: []microflows.MicroflowObject{split, trueLog, falseLog, merge},
			Flows: []*microflows.SequenceFlow{
				mkBranchFlow("split", "true_log", &microflows.ExpressionCase{Expression: "true"}),
				mkBranchFlow("split", "false_log", &microflows.ExpressionCase{Expression: "false"}),
				mkFlow("true_log", "merge"),
				mkFlow("false_log", "merge"),
			},
		},
	}

	var lines []string
	emitLoopBody(nil, loop, nil, nil, nil, nil, &lines, 0, nil, 0, nil)
	out := strings.Join(lines, "\n")
	for _, want := range []string{"if $Item/IsActive then", "else", "end if;"} {
		if !strings.Contains(out, want) {
			t.Fatalf("loop body should emit structured %q, got:\n%s", want, out)
		}
	}
}

// TestTraverseFlow_UnsupportedActivitySkipsAnnotations verifies that when the
// describer falls back to a `-- Unsupported action type: ...` placeholder, it
// does NOT also emit @position / @anchor before the comment. Annotations are
// only valid as a prefix of real MDL statements; orphaning them above a pure
// line comment triggers a parse error during `exec`.
func TestTraverseFlow_UnsupportedActivitySkipsAnnotations(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("soap"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("soap")},
			Action:       &microflows.UnknownAction{TypeName: "Microflows$CallWebServiceAction"},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "soap")},
		mkID("soap"):  {mkFlow("soap", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	found := false
	for idx, line := range lines {
		if contains(line, "Unsupported action type") {
			found = true
			if idx > 0 && contains(lines[idx-1], "@") {
				t.Errorf("expected no annotation prefix before unsupported-action comment, got: %v", lines)
			}
		}
	}
	if !found {
		t.Errorf("expected unsupported-action comment, got: %v", lines)
	}
}

func TestTraverseFlow_CommonActivityJoinKeepsTailOutsideBranches(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("outer_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("outer_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Allowed"},
		},
		mkID("allowed_act"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("allowed_act")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "allowed branch"}}},
		},
		mkID("denied_act"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("denied_act")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "denied branch"}}},
		},
		mkID("tail_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("tail_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$FollowUp"},
		},
		mkID("tail_true"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("tail_true")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "follow true"}}},
		},
		mkID("tail_false"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("tail_false")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "follow false"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "outer_split")},
		mkID("outer_split"): {
			mkBranchFlow("outer_split", "allowed_act", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("outer_split", "denied_act", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("allowed_act"): {mkFlow("allowed_act", "tail_split")},
		mkID("denied_act"):  {mkFlow("denied_act", "tail_split")},
		mkID("tail_split"): {
			mkBranchFlow("tail_split", "tail_true", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("tail_split", "tail_false", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("tail_true"):  {mkFlow("tail_true", "end")},
		mkID("tail_false"): {mkFlow("tail_false", "end")},
	}

	joinID := findMergeForSplit(nil, mkID("outer_split"), flowsByOrigin, activityMap)
	if joinID != mkID("tail_split") {
		t.Fatalf("outer split paired with %q, want common tail activity %q", joinID, mkID("tail_split"))
	}

	splitMergeMap := map[model.ID]model.ID{mkID("outer_split"): joinID}
	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	firstEndIf := strings.Index(out, "end if;")
	tailIf := strings.Index(out, "if $FollowUp then")
	if firstEndIf == -1 || tailIf == -1 || firstEndIf > tailIf {
		t.Fatalf("shared tail must be emitted after the outer IF closes:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "if $FollowUp then") && strings.HasPrefix(line, "  ") {
			t.Fatalf("shared tail must be top-level, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_SequentialIfWithoutElseKeepsContinuationOutsideFirstIf(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("logo_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("logo_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Logo != empty"},
		},
		mkID("logo_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("logo_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "logo"}}},
		},
		mkID("logo_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("logo_merge")},
		mkID("after_logo"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after_logo")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after logo"}}},
		},
		mkID("cover_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("cover_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Cover != empty"},
		},
		mkID("cover_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("cover_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "cover"}}},
		},
		mkID("cover_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("cover_merge")},
		mkID("end"):         &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "logo_split")},
		mkID("logo_split"): {
			mkBranchFlow("logo_split", "logo_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("logo_split", "logo_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("logo_then"):  {mkFlow("logo_then", "logo_merge")},
		mkID("logo_merge"): {mkFlow("logo_merge", "after_logo")},
		mkID("after_logo"): {mkFlow("after_logo", "cover_split")},
		mkID("cover_split"): {
			mkBranchFlow("cover_split", "cover_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("cover_split", "cover_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("cover_then"):  {mkFlow("cover_then", "cover_merge")},
		mkID("cover_merge"): {mkFlow("cover_merge", "end")},
	}

	splitMergeMap := map[model.ID]model.ID{
		mkID("logo_split"):  findMergeForSplit(nil, mkID("logo_split"), flowsByOrigin, activityMap),
		mkID("cover_split"): findMergeForSplit(nil, mkID("cover_split"), flowsByOrigin, activityMap),
	}
	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	firstEndIf := strings.Index(out, "end if;")
	afterLogo := strings.Index(out, "after logo")
	if firstEndIf == -1 || afterLogo == -1 || firstEndIf > afterLogo {
		t.Fatalf("continuation after first IF was emitted inside the IF:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "after logo") && strings.HasPrefix(line, "  ") {
			t.Fatalf("continuation after first IF must be top-level, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_GuardBranchWithMultipleActivitiesKeepsContinuationOutsideElse(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$HasExistingData"},
		},
		mkID("then_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("then_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "terminal branch"}}},
		},
		mkID("then_return"): &microflows.EndEvent{BaseMicroflowObject: mkObj("then_return")},
		mkID("tail_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("tail_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "shared tail"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "split")},
		mkID("split"): {
			mkBranchFlow("split", "then_log", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("split", "tail_log", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("then_log"): {mkFlow("then_log", "then_return")},
		mkID("tail_log"): {mkFlow("tail_log", "end")},
	}
	if !branchFlowTerminatesBeforeMerge(flowsByOrigin[mkID("split")][0], "", activityMap, flowsByOrigin, nil) {
		t.Fatal("expected multi-activity true branch to be terminal")
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if strings.Contains(out, "\nelse\n") {
		t.Fatalf("terminal guard branch should not wrap continuation in ELSE:\n%s", out)
	}
	if strings.Index(out, "end if;") > strings.Index(out, "shared tail") {
		t.Fatalf("shared tail must be emitted after the guard IF closes:\n%s", out)
	}
}

func TestTraverseFlow_NestedTerminalGuardBranchSuppressesEmptyOuterElse(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("outer_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("outer_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$NeedsValidation"},
		},
		mkID("inner_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("inner_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$IsValid"},
		},
		mkID("valid_return"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("valid_return")},
		mkID("invalid_return"): &microflows.EndEvent{BaseMicroflowObject: mkObj("invalid_return")},
		mkID("tail_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("tail_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "fallback tail"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "outer_split")},
		mkID("outer_split"): {
			mkBranchFlow("outer_split", "inner_split", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("outer_split", "tail_log", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("inner_split"): {
			mkBranchFlow("inner_split", "valid_return", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("inner_split", "invalid_return", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("tail_log"): {mkFlow("tail_log", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if got := strings.Count(out, "\n  else\n"); got != 1 {
		t.Fatalf("expected only the nested IF ELSE, got %d ELSE blocks:\n%s", got, out)
	}
	if strings.Index(out, "end if;") > strings.Index(out, "fallback tail") {
		t.Fatalf("fallback tail must be emitted after the outer guard IF closes:\n%s", out)
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
	if len(stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d: %v", len(stmts), stmts)
	}
	assertContains(t, stmts[0], "log error")
	assertContains(t, stmts[0], "Something failed")
	assertContains(t, stmts[1], "return;")
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

func TestCollectErrorHandlerStatements_StructuredIfEmitsEndIf(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$latestHttpResponse != empty"},
		},
		mkID("return_error"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("return_error"),
			ReturnValue:         "latestHttpResponse",
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("after"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'Synthetic'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after"}}},
		},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("split"): {
			mkBranchFlow("split", "return_error", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("split", "merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("merge"): {mkFlow("merge", "after")},
	}

	stmts := e.collectErrorHandlerStatements(mkID("split"), activityMap, flowsByOrigin, nil, nil)
	got := strings.Join(stmts, "\n")

	assertContains(t, got, "if $latestHttpResponse != empty then")
	assertContains(t, got, "return $latestHttpResponse;")
	assertContains(t, got, "else")
	assertContains(t, got, "end if;")
	if strings.Contains(got, "after") {
		t.Fatalf("error handler traversal crossed the rejoin merge: %s", got)
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

func TestBranchFlowTerminatesBeforeMerge_InheritanceSplitFallsThroughParentMerge(t *testing.T) {
	entityID := mkID("entity-specialized")
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("type_split"): &microflows.InheritanceSplit{
			BaseMicroflowObject: mkObj("type_split"),
			VariableName:        "Input",
		},
		mkID("set_value"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("set_value")},
			Action:       &microflows.ChangeVariableAction{VariableName: "Value", Value: "$Input/Value"},
		},
		mkID("error_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("error_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "no matching type"}}},
		},
		mkID("error_return"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("error_return"),
			ReturnValue:         "empty",
		},
		mkID("parent_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("parent_merge")},
		mkID("tail_return"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("tail_return"),
			ReturnValue:         "'ok'",
		},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("type_split"): {
			mkBranchFlow("type_split", "set_value", &microflows.InheritanceCase{EntityID: entityID}),
			mkBranchFlow("type_split", "error_log", &microflows.InheritanceCase{}),
		},
		mkID("set_value"):    {mkFlow("set_value", "parent_merge")},
		mkID("error_log"):    {mkFlow("error_log", "error_return")},
		mkID("parent_merge"): {mkFlow("parent_merge", "tail_return")},
	}
	parentBranch := mkFlow("outer_split", "type_split")

	if branchFlowTerminatesBeforeMerge(parentBranch, mkID("parent_merge"), activityMap, flowsByOrigin, nil) {
		t.Fatal("inheritance split branch that falls through the parent merge must not be classified as terminal")
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

// =============================================================================
// negateIfCondition
// =============================================================================

func TestNegateIfCondition(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"if $Active then", "if not($Active) then"},
		{"if $X = 42 then", "if not($X = 42) then"},
		{"if not($Done) then", "if $Done then"},                           // double-negation removal
		{"if not($A and $B) then", "if $A and $B then"},                   // unwrap not()
		{"if true then", "if not(true) then"},                             // literal
		{"something else", "something else"},                              // no match — passthrough
		{"if find($S,'{{') >= 0 then", "if not(find($S,'{{') >= 0) then"}, // complex expression
		{"if not($A) or $B) then", "if not(not($A) or $B)) then"},         // unbalanced — do NOT unwrap
	}
	for _, tt := range tests {
		got := negateIfCondition(tt.in)
		if got != tt.want {
			t.Errorf("negateIfCondition(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// =============================================================================
// Empty-then swap: true branch → merge produces negated condition
// =============================================================================

func TestTraverseFlow_EmptyThenSwap(t *testing.T) {
	e := newTestExecutor()

	// Graph: start → split → (true) → merge → end
	//                      → (false) → log → merge
	// Without swap: if cond then else log end if;
	// With swap: if not(cond) then log end if;
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Active"},
		},
		mkID("log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("log")},
			Action: &microflows.LogMessageAction{
				LogLevel:    "Info",
				LogNodeName: "'Test'",
			},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {{OriginID: mkID("start"), DestinationID: mkID("split")}},
		mkID("split"): {
			{OriginID: mkID("split"), DestinationID: mkID("merge"), CaseValue: microflows.EnumerationCase{Value: "true"}},
			{OriginID: mkID("split"), DestinationID: mkID("log"), CaseValue: microflows.EnumerationCase{Value: "false"}},
		},
		mkID("log"):   {{OriginID: mkID("log"), DestinationID: mkID("merge")}},
		mkID("merge"): {{OriginID: mkID("merge"), DestinationID: mkID("end")}},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}

	var lines []string
	visited := map[model.ID]bool{}

	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	output := ""
	for _, l := range lines {
		output += l + "\n"
	}

	if !strings.Contains(output, "not($Active)") {
		t.Errorf("expected negated condition 'not($Active)', got:\n%s", output)
	}
	if !strings.Contains(output, "log info") {
		t.Errorf("expected 'log info' in output, got:\n%s", output)
	}
	if strings.Contains(output, "else") {
		t.Errorf("expected no empty else block, got:\n%s", output)
	}
}

func TestTraverseFlow_BothBranchesToMerge_NoSwap(t *testing.T) {
	e := newTestExecutor()

	// Graph: start → split → (true) → merge → end
	//                      → (false) → merge
	// Both branches empty — no swap should occur, condition stays positive.
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Flag"},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {{OriginID: mkID("start"), DestinationID: mkID("split")}},
		mkID("split"): {
			{OriginID: mkID("split"), DestinationID: mkID("merge"), CaseValue: microflows.EnumerationCase{Value: "true"}},
			{OriginID: mkID("split"), DestinationID: mkID("merge"), CaseValue: microflows.EnumerationCase{Value: "false"}},
		},
		mkID("merge"): {{OriginID: mkID("merge"), DestinationID: mkID("end")}},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}

	var lines []string
	visited := map[model.ID]bool{}

	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	output := ""
	for _, l := range lines {
		output += l + "\n"
	}

	// Condition should NOT be negated — both branches are empty
	if strings.Contains(output, "not($Flag)") {
		t.Errorf("expected no negation when both branches go to merge, got:\n%s", output)
	}
}
