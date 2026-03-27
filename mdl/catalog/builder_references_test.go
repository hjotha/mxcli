// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func newAction(id string, action microflows.MicroflowAction) *microflows.ActionActivity {
	return &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(id)},
			},
		},
		Action: action,
	}
}

func newLoop(id string, children ...microflows.MicroflowObject) *microflows.LoopedActivity {
	return &microflows.LoopedActivity{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(id)},
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{Objects: children},
	}
}

func TestCollectActionActivities_TopLevelOnly(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			newAction("a1", &microflows.MicroflowCallAction{}),
			newAction("a2", &microflows.CreateObjectAction{}),
		},
	}
	result := collectActionActivities(oc)
	if len(result) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(result))
	}
}

func TestCollectActionActivities_InsideLoop(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			newLoop("loop1",
				newAction("inner1", &microflows.MicroflowCallAction{}),
				newAction("inner2", &microflows.ShowPageAction{}),
			),
			newAction("outer1", &microflows.RetrieveAction{}),
		},
	}
	result := collectActionActivities(oc)
	if len(result) != 3 {
		t.Fatalf("expected 3 activities (2 inside loop + 1 outside), got %d", len(result))
	}
}

func TestCollectActionActivities_NestedLoops(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			newLoop("outer-loop",
				newLoop("inner-loop",
					newAction("deep", &microflows.MicroflowCallAction{}),
				),
			),
		},
	}
	result := collectActionActivities(oc)
	if len(result) != 1 {
		t.Fatalf("expected 1 deeply nested activity, got %d", len(result))
	}
	if result[0].ID != "deep" {
		t.Errorf("expected activity ID 'deep', got %q", result[0].ID)
	}
}

func TestCollectActionActivities_NilCollection(t *testing.T) {
	result := collectActionActivities(nil)
	if result != nil {
		t.Fatalf("expected nil for nil collection, got %v", result)
	}
}

func TestCollectActionActivities_SkipsNilActions(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			newAction("no-action", nil),
			newAction("has-action", &microflows.MicroflowCallAction{}),
		},
	}
	result := collectActionActivities(oc)
	if len(result) != 1 {
		t.Fatalf("expected 1 activity (skipping nil action), got %d", len(result))
	}
}
