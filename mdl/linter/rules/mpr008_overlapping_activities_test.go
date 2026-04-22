// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/linter"
)

// NOTE: The full Check() logic requires ctx.Reader().GetMicroflow() to read microflow
// positions from a real MPR file. The overlap detection algorithm (collect positions →
// pairwise distance check) is inline in Check() and cannot be unit-tested without
// building a mock mpr.Reader. This rule currently lacks end-to-end coverage;
// behavioral testing requires a real .mpr project with overlapping activities.

func TestOverlappingActivitiesRule_NilReader(t *testing.T) {
	r := NewOverlappingActivitiesRule()
	ctx := linter.NewLintContextFromDB(nil)
	violations := r.Check(ctx)
	if violations != nil {
		t.Errorf("expected nil with nil reader, got %v", violations)
	}
}

func TestOverlappingActivitiesRule_Metadata(t *testing.T) {
	r := NewOverlappingActivitiesRule()
	if r.ID() != "MPR008" {
		t.Errorf("ID = %q, want MPR008", r.ID())
	}
	if r.Category() != "correctness" {
		t.Errorf("Category = %q, want correctness", r.Category())
	}
	if r.Name() != "OverlappingActivities" {
		t.Errorf("Name = %q, want OverlappingActivities", r.Name())
	}
}

// NOTE: The full Check() logic for overlapping activities requires ctx.Reader().GetMicroflow()
// which needs a real *mpr.Reader. This rule currently lacks end-to-end coverage;
// behavioral testing requires a real .mpr project. Here we verify constants used for
// overlap threshold.

func TestOverlappingActivities_Constants(t *testing.T) {
	if activityBoxWidth != 120 {
		t.Errorf("activityBoxWidth = %d, want 120", activityBoxWidth)
	}
	if activityBoxHeight != 60 {
		t.Errorf("activityBoxHeight = %d, want 60", activityBoxHeight)
	}
}

// Test the walker function indirectly: the Check method calls an inline collect function
// that processes ActionActivity, LoopedActivity, ExclusiveSplit, and ExclusiveMerge.
// Since collect is defined inline in Check(), we can only test it end-to-end via a real reader.
// For unit coverage, we verify the nil-reader guard and metadata above.
