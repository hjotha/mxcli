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

	// StartEvent produces no output. Empty EndEvents are emitted as bare
	// returns so terminal void branches survive roundtrip instead of falling
	// through into later statements.
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

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
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

func TestTraverseFlow_IfWithoutExplicitMergeKeepsSharedContinuationOutside(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("first_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("first_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$FirstFlag"},
		},
		mkID("first_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("first_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "first branch"}}},
		},
		mkID("shared_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("shared_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$SecondFlag"},
		},
		mkID("shared_then"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("shared_then")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "shared branch"}}},
		},
		mkID("shared_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("shared_merge")},
		mkID("end"):          &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "first_split")},
		mkID("first_split"): {
			mkBranchFlow("first_split", "first_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("first_split", "shared_split", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("first_then"): {mkFlow("first_then", "shared_split")},
		mkID("shared_split"): {
			mkBranchFlow("shared_split", "shared_then", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("shared_split", "shared_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("shared_then"):  {mkFlow("shared_then", "shared_merge")},
		mkID("shared_merge"): {mkFlow("shared_merge", "end")},
	}

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
	if got := splitMergeMap[mkID("first_split")]; got != mkID("shared_split") {
		t.Fatalf("first split paired with %q, want shared continuation %q", got, mkID("shared_split"))
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	firstEndIf := strings.Index(out, "end if;")
	secondIf := strings.Index(out, "if $SecondFlag then")
	if firstEndIf == -1 || secondIf == -1 || firstEndIf > secondIf {
		t.Fatalf("shared continuation was emitted inside first IF:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "if $SecondFlag then") && strings.HasPrefix(line, "  ") {
			t.Fatalf("shared continuation must be top-level, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_NestedTerminalBranchUsesParentMerge(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("support_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("support_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "SampleAuth.CurrentUserHasSupportRole()"},
		},
		mkID("admin_check"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("admin_check")},
			Action: &microflows.MicroflowCallAction{
				ResultVariableName: "UserHasAdminRole",
				MicroflowCall:      &microflows.MicroflowCall{Microflow: "SampleAuth.UserHasAdminRole"},
			},
		},
		mkID("admin_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("admin_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$UserHasAdminRole"},
		},
		mkID("denied"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("denied")},
			Action:       &microflows.ShowMessageAction{Template: &model.Text{Translations: map[string]string{"en_US": "denied"}}},
		},
		mkID("denied_end"):  &microflows.EndEvent{BaseMicroflowObject: mkObj("denied_end")},
		mkID("outer_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("outer_merge")},
		mkID("shared_tail"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("shared_tail")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "shared tail"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "support_split")},
		mkID("support_split"): {
			mkBranchFlow("support_split", "outer_merge", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("support_split", "admin_check", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("admin_check"): {mkFlow("admin_check", "admin_split")},
		mkID("admin_split"): {
			mkBranchFlow("admin_split", "outer_merge", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("admin_split", "denied", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("denied"):      {mkFlow("denied", "denied_end")},
		mkID("outer_merge"): {mkFlow("outer_merge", "shared_tail")},
		mkID("shared_tail"): {mkFlow("shared_tail", "end")},
	}

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if got := strings.Count(out, "shared tail"); got != 1 {
		t.Fatalf("shared continuation must be emitted once, got %d:\n%s", got, out)
	}
	sharedTail := strings.Index(out, "shared tail")
	denied := strings.Index(out, "denied")
	if denied == -1 || sharedTail == -1 || denied > sharedTail {
		t.Fatalf("terminal denial branch should stay before shared continuation:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "shared tail") && strings.HasPrefix(line, "  ") {
			t.Fatalf("shared continuation must be outside the guard branch, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_NestedSplitSharedTailStaysOutsideParentIf(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("outer_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("outer_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Existing != empty"},
		},
		mkID("inner_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("inner_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Existing/IsReusable"},
		},
		mkID("reuse_end"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("reuse_end"),
			ReturnValue:         "$Existing",
		},
		mkID("discard"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("discard")},
			Action:       &microflows.DeleteObjectAction{DeleteVariable: "Existing"},
		},
		mkID("shared_fetch"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("shared_fetch")},
			Action: &microflows.MicroflowCallAction{
				ResultVariableName: "Fetched",
				MicroflowCall:      &microflows.MicroflowCall{Microflow: "Synthetic.Fetch"},
			},
		},
		mkID("shared_success"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("shared_success"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Fetched != empty"},
		},
		mkID("success_end"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("success_end"),
			ReturnValue:         "$Fetched",
		},
		mkID("empty_end"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("empty_end"),
			ReturnValue:         "empty",
		},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "outer_split")},
		mkID("outer_split"): {
			mkBranchFlow("outer_split", "inner_split", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("outer_split", "shared_fetch", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("inner_split"): {
			mkBranchFlow("inner_split", "reuse_end", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("inner_split", "discard", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("discard"):      {mkFlow("discard", "shared_fetch")},
		mkID("shared_fetch"): {mkFlow("shared_fetch", "shared_success")},
		mkID("shared_success"): {
			mkBranchFlow("shared_success", "success_end", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("shared_success", "empty_end", &microflows.ExpressionCase{Expression: "false"}),
		},
	}

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
	if got := splitMergeMap[mkID("outer_split")]; got != mkID("shared_fetch") {
		t.Fatalf("outer split paired with %q, want shared continuation %q", got, mkID("shared_fetch"))
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if got := strings.Count(out, "Synthetic.Fetch"); got != 1 {
		t.Fatalf("shared tail must be emitted once, got %d:\n%s", got, out)
	}
	outerEndIf := strings.Index(out, "end if;")
	sharedFetch := strings.Index(out, "Synthetic.Fetch")
	if outerEndIf == -1 || sharedFetch == -1 || outerEndIf > sharedFetch {
		t.Fatalf("shared tail was emitted inside parent IF:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "Synthetic.Fetch") && strings.HasPrefix(line, "  ") {
			t.Fatalf("shared tail must be top-level, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_NestedSplitStopsBeforeParentSharedFailureTail(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("outer_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("outer_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$OuterIsValid"},
		},
		mkID("inner_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("inner_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$InnerIsValid"},
		},
		mkID("success"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("success")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "success branch"}}},
		},
		mkID("parent_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("parent_merge")},
		mkID("failure"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("failure")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "shared failure tail"}}},
		},
		mkID("final_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("final_merge")},
		mkID("after"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after shared join"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "outer_split")},
		mkID("outer_split"): {
			mkBranchFlow("outer_split", "inner_split", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("outer_split", "parent_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("inner_split"): {
			mkBranchFlow("inner_split", "success", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("inner_split", "parent_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("success"):      {mkFlow("success", "final_merge")},
		mkID("parent_merge"): {mkFlow("parent_merge", "failure")},
		mkID("failure"):      {mkFlow("failure", "final_merge")},
		mkID("final_merge"):  {mkFlow("final_merge", "after")},
		mkID("after"):        {mkFlow("after", "end")},
	}

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
	if got := splitMergeMap[mkID("outer_split")]; got != mkID("parent_merge") {
		t.Fatalf("outer split paired with %q, want parent shared-tail merge %q", got, mkID("parent_merge"))
	}
	if got := splitMergeMap[mkID("inner_split")]; got != mkID("final_merge") {
		t.Fatalf("inner split paired with %q, want final join %q", got, mkID("final_merge"))
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if got := strings.Count(out, "shared failure tail"); got != 1 {
		t.Fatalf("shared failure tail must be emitted once, got %d:\n%s", got, out)
	}
	for _, line := range lines {
		if strings.Contains(line, "shared failure tail") && strings.HasPrefix(line, "    ") {
			t.Fatalf("shared failure tail leaked inside nested branch, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_NestedGuardLocalMergeDoesNotConsumeParentTail(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("outer_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("outer_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Existing != empty"},
		},
		mkID("inner_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("inner_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Existing/ExpiresAt != empty"},
		},
		mkID("guard_split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("guard_split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Existing/ExpiresAt > [%CurrentDateTime%]"},
		},
		mkID("reuse_end"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("reuse_end"),
			ReturnValue:         "$Existing",
		},
		mkID("local_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("local_merge")},
		mkID("discard"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("discard")},
			Action:       &microflows.DeleteObjectAction{DeleteVariable: "Existing"},
		},
		mkID("parent_merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("parent_merge")},
		mkID("shared_fetch"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("shared_fetch")},
			Action: &microflows.MicroflowCallAction{
				ResultVariableName: "Fetched",
				MicroflowCall:      &microflows.MicroflowCall{Microflow: "Synthetic.Fetch"},
			},
		},
		mkID("success_end"): &microflows.EndEvent{
			BaseMicroflowObject: mkObj("success_end"),
			ReturnValue:         "$Fetched",
		},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "outer_split")},
		mkID("outer_split"): {
			mkBranchFlow("outer_split", "inner_split", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("outer_split", "parent_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("inner_split"): {
			mkBranchFlow("inner_split", "guard_split", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("inner_split", "local_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("guard_split"): {
			mkBranchFlow("guard_split", "reuse_end", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("guard_split", "local_merge", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("local_merge"):  {mkFlow("local_merge", "discard")},
		mkID("discard"):      {mkFlow("discard", "parent_merge")},
		mkID("parent_merge"): {mkFlow("parent_merge", "shared_fetch")},
		mkID("shared_fetch"): {mkFlow("shared_fetch", "success_end")},
	}

	splitMergeMap := findSplitMergePointsForGraph(nil, activityMap, flowsByOrigin)
	if got := splitMergeMap[mkID("outer_split")]; got != mkID("parent_merge") {
		t.Fatalf("outer split paired with %q, want parent merge %q", got, mkID("parent_merge"))
	}
	if got := splitMergeMap[mkID("inner_split")]; got != mkID("local_merge") {
		t.Fatalf("inner split paired with %q, want local merge %q", got, mkID("local_merge"))
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 0, nil, 0, nil)

	out := strings.Join(lines, "\n")
	if got := strings.Count(out, "Synthetic.Fetch"); got != 1 {
		t.Fatalf("shared tail must be emitted once, got %d:\n%s", got, out)
	}
	outerEndIf := strings.LastIndex(out[:strings.Index(out, "Synthetic.Fetch")], "end if;")
	if outerEndIf == -1 {
		t.Fatalf("expected outer IF to close before shared tail:\n%s", out)
	}
	for _, line := range lines {
		if strings.Contains(line, "Synthetic.Fetch") && strings.HasPrefix(line, "  ") {
			t.Fatalf("shared tail must be outside parent IF, got %q in:\n%s", line, out)
		}
	}
}

func TestTraverseFlow_IfInsideLoop(t *testing.T) {
	e := newTestExecutor()

	loop := &microflows.LoopedActivity{
		BaseMicroflowObject: mkObj("loop"),
		LoopSource: &microflows.IterableList{
			BaseElement:      model.BaseElement{ID: mkID("src")},
			VariableName:     "item",
			ListVariableName: "items",
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{
			BaseElement: model.BaseElement{ID: mkID("loop-oc")},
			Objects: []microflows.MicroflowObject{
				&microflows.ExclusiveSplit{
					BaseMicroflowObject: mkObj("split"),
					SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$item/Flag"},
				},
				&microflows.ActionActivity{
					BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("then")},
					Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "then"}}},
				},
				&microflows.ActionActivity{
					BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("else")},
					Action:       &microflows.LogMessageAction{LogLevel: "Info", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "else"}}},
				},
				&microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
				&microflows.EndEvent{BaseMicroflowObject: mkObj("loop-end")},
			},
			Flows: []*microflows.SequenceFlow{
				mkBranchFlow("split", "then", &microflows.ExpressionCase{Expression: "true"}),
				mkBranchFlow("split", "else", &microflows.ExpressionCase{Expression: "false"}),
				mkFlow("then", "merge"),
				mkFlow("else", "merge"),
				mkFlow("merge", "loop-end"),
			},
		},
	}

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("loop"):  loop,
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "loop")},
		mkID("loop"):  {mkFlow("loop", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	foundIf := false
	foundElse := false
	foundEndIf := false
	for _, line := range lines {
		if contains(line, "if $item/Flag then") {
			foundIf = true
		}
		if contains(line, "else") {
			foundElse = true
		}
		if contains(line, "end if;") {
			foundEndIf = true
		}
	}
	if !foundIf || !foundElse || !foundEndIf {
		t.Fatalf("expected structured if inside loop, got: %v", lines)
	}
}

// TestTraverseFlow_UnsupportedActivitySkipsAnnotations verifies that when the
// describer falls back to a `-- Unsupported action type: ...` placeholder, it
// does NOT also emit @position / @anchor before the comment. Annotations are
// only valid as a prefix of real MDL statements; orphaning them above a pure
// line comment triggers `no viable alternative at input '@position...end'`
// during `exec`. Covers audit cluster A6 (ANNOT_GLUED).
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
	for i, line := range lines {
		if contains(line, "Unsupported action type") {
			if i > 0 && contains(lines[i-1], "@position") {
				t.Errorf("expected no @position prefix before unsupported-action comment, got: %v", lines)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected unsupported-action comment, got: %v", lines)
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

func TestCollectErrorHandlerStatements_TraverseLocalMerge(t *testing.T) {
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
	if len(stmts) != 2 {
		t.Fatalf("expected local merge tail to be included, got %d statements: %v", len(stmts), stmts)
	}
	assertContains(t, stmts[1], "after")
}

func TestCollectErrorHandlerStatements_StopsAtSharedContinuation(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("source"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("source")},
			Action:       &microflows.RestCallAction{ErrorHandlingType: microflows.ErrorHandlingTypeCustomWithoutRollback},
		},
		mkID("err_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("err_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Error", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "err"}}},
		},
		mkID("after"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("after")},
			Action:       &microflows.LogMessageAction{LogLevel: "Debug", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "after"}}},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end"), ReturnValue: "$Response"},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("source"):  {mkFlow("source", "after"), mkErrorFlow("source", "err_log")},
		mkID("err_log"): {mkFlow("err_log", "after")},
		mkID("after"):   {mkFlow("after", "end")},
	}

	stmts := e.collectErrorHandlerStatements(mkID("err_log"), activityMap, flowsByOrigin, nil, nil)
	if len(stmts) != 1 {
		t.Fatalf("expected only the error-handler statement, got %d: %v", len(stmts), stmts)
	}
	assertContains(t, stmts[0], "log error")
	if strings.Contains(strings.Join(stmts, "\n"), "after") {
		t.Fatalf("shared continuation leaked into error handler: %v", stmts)
	}
}

func TestCollectErrorHandlerStatements_IncludesTailAfterEmptyElseMerge(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("err_log"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("err_log")},
			Action:       &microflows.LogMessageAction{LogLevel: "Error", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "request failed"}}},
		},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$latestHttpResponse != empty"},
		},
		mkID("import_error"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("import_error")},
			Action: &microflows.ImportXmlAction{
				ErrorHandlingType: microflows.ErrorHandlingTypeCustomWithoutRollback,
				ResultHandling:    &microflows.ResultHandlingMapping{ResultVariable: "ErrorPayload"},
			},
		},
		mkID("payload_end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("payload_end"), ReturnValue: "$ErrorPayload"},
		mkID("merge"):       &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("fallback"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("fallback")},
			Action:       &microflows.CreateObjectAction{OutputVariable: "GenericError", EntityQualifiedName: "SampleErrors.GenericError"},
		},
		mkID("fallback_end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("fallback_end"), ReturnValue: "$GenericError"},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("err_log"):      {mkFlow("err_log", "split")},
		mkID("split"):        {mkBranchFlow("split", "import_error", &microflows.ExpressionCase{Expression: "true"}), mkBranchFlow("split", "merge", &microflows.ExpressionCase{Expression: "false"})},
		mkID("import_error"): {mkFlow("import_error", "payload_end"), mkErrorFlow("import_error", "merge")},
		mkID("merge"):        {mkFlow("merge", "fallback")},
		mkID("fallback"):     {mkFlow("fallback", "fallback_end")},
	}

	stmts := e.collectErrorHandlerStatements(mkID("err_log"), activityMap, flowsByOrigin, nil, nil)
	out := strings.Join(stmts, "\n")
	assertContains(t, out, "if $latestHttpResponse != empty then")
	assertContains(t, out, "else")
	assertContains(t, out, "end if;")
	assertContains(t, out, "$GenericError = create SampleErrors.GenericError")
	if strings.Index(out, "$GenericError = create") < strings.Index(out, "else") {
		t.Fatalf("fallback tail must be emitted on the fallback branch, got:\n%s", out)
	}
}

func TestTraverseFlow_CustomErrorHandlerWithIfElse(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("commit"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("commit")},
			Action: &microflows.CommitObjectsAction{
				CommitVariable:    "Obj",
				ErrorHandlingType: microflows.ErrorHandlingTypeCustomWithoutRollback,
			},
		},
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$HasFallback"},
		},
		mkID("warn"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("warn")},
			Action:       &microflows.LogMessageAction{LogLevel: "Warning", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "warn"}}},
		},
		mkID("err"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("err")},
			Action:       &microflows.LogMessageAction{LogLevel: "Error", LogNodeName: "'App'", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "err"}}},
		},
		mkID("merge"):   &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("err-end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("err-end")},
		mkID("end"):     &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"):  {mkFlow("start", "commit")},
		mkID("commit"): {mkFlow("commit", "end"), mkErrorFlow("commit", "split")},
		mkID("split"): {
			mkBranchFlow("split", "warn", &microflows.ExpressionCase{Expression: "true"}),
			mkBranchFlow("split", "err", &microflows.ExpressionCase{Expression: "false"}),
		},
		mkID("warn"):  {mkFlow("warn", "merge")},
		mkID("err"):   {mkFlow("err", "merge")},
		mkID("merge"): {mkFlow("merge", "err-end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	joined := strings.Join(lines, "\n")
	assertContains(t, joined, "commit $Obj on error without rollback {")
	assertContains(t, joined, "if $HasFallback then")
	assertContains(t, joined, "else")
	assertContains(t, joined, "end if;")
	assertContains(t, joined, "};")
}

func TestTraverseFlow_CustomErrorHandlerSharedContinuationEmitsEmptyBlock(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("commit"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("commit")},
			Action: &microflows.CommitObjectsAction{
				CommitVariable:    "Order",
				ErrorHandlingType: microflows.ErrorHandlingTypeCustomWithoutRollback,
			},
		},
		mkID("shared_tail"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("shared_tail")},
			Action: &microflows.LogMessageAction{
				LogLevel:        "Info",
				LogNodeName:     "'App'",
				MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "shared tail"}},
			},
		},
		mkID("end"): &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"):       {mkFlow("start", "commit")},
		mkID("commit"):      {mkFlow("commit", "shared_tail"), mkErrorFlow("commit", "shared_tail")},
		mkID("shared_tail"): {mkFlow("shared_tail", "end")},
	}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, nil, visited, nil, nil, &lines, 0, nil, 0, nil)

	joined := strings.Join(lines, "\n")
	assertContains(t, joined, "commit $Order on error without rollback { };")
	if got := strings.Count(joined, "shared tail"); got != 1 {
		t.Fatalf("shared continuation must be described once, got %d occurrences:\n%s", got, joined)
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
