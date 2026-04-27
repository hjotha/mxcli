// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestTraverseFlow_EnumSplit(t *testing.T) {
	e := newTestExecutor()

	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("split"): &microflows.ExclusiveSplit{
			BaseMicroflowObject: mkObj("split"),
			SplitCondition:      &microflows.ExpressionSplitCondition{Expression: "$Status"},
		},
		mkID("open"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("open")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "open"}}},
		},
		mkID("closed"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("closed")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "closed"}}},
		},
		mkID("other"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("other")},
			Action:       &microflows.LogMessageAction{LogLevel: "Info", MessageTemplate: &model.Text{Translations: map[string]string{"en_US": "other"}}},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("split"): {
			mkBranchFlow("split", "open", microflows.EnumerationCase{Value: "Open"}),
			mkBranchFlow("split", "closed", microflows.EnumerationCase{Value: "Closed"}),
			mkFlow("split", "other"),
		},
		mkID("open"):   {mkFlow("open", "merge")},
		mkID("closed"): {mkFlow("closed", "merge")},
		mkID("other"):  {mkFlow("other", "merge")},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("split"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 1, nil, 0, nil)

	assertContainsAny(t, lines, "split enum $Status")
	assertContainsAny(t, lines, "case Open")
	assertContainsAny(t, lines, "case Closed")
	assertContainsAny(t, lines, "else")
	assertContainsAny(t, lines, "end split;")
}

func assertContainsAny(t *testing.T, lines []string, want string) {
	t.Helper()
	for _, line := range lines {
		if contains(line, want) {
			return
		}
	}
	t.Fatalf("Expected output to contain %q, got %v", want, lines)
}
