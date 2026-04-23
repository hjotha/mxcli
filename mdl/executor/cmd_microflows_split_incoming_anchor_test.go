// SPDX-License-Identifier: Apache-2.0

// Regression tests for the split-form @anchor emission — the incoming anchor
// that lands on an ExclusiveSplit / InheritanceSplit was previously lost at
// describe time because emitAnchorAnnotation early-returned for splits. The
// builder happily consumed @anchor(to: X) and set DestinationConnectionIndex
// on the flow entering the split, but the describer never read it back.
//
// The fix moves split anchors into a dedicated emitSplitAnchorAnnotation path
// that emits `@anchor(to: X, true: (...), false: (...))` whenever any of the
// three has a non-default value.
package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestEmitSplitAnchor_EmitsIncomingToSide(t *testing.T) {
	splitID := model.ID("split-1")
	split := &microflows.ExclusiveSplit{}
	split.ID = splitID

	incoming := &microflows.SequenceFlow{
		DestinationID:              splitID,
		DestinationConnectionIndex: AnchorTop,
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{}
	flowsByDest := map[model.ID][]*microflows.SequenceFlow{
		splitID: {incoming},
	}

	var lines []string
	emitAnchorAnnotation(split, flowsByOrigin, flowsByDest, &lines, "")

	if len(lines) != 1 {
		t.Fatalf("expected 1 anchor line, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "to: top") {
		t.Errorf("expected split anchor to include `to: top`, got %q", lines[0])
	}
	// The split has no outgoing flows, so true/false fragments must be absent.
	if strings.Contains(lines[0], "true:") || strings.Contains(lines[0], "false:") {
		t.Errorf("no outgoing flows configured, but output contains branch fragment: %q", lines[0])
	}
}

func TestEmitSplitAnchor_EmitsBranchAnchors(t *testing.T) {
	splitID := model.ID("split-2")
	split := &microflows.ExclusiveSplit{}
	split.ID = splitID

	trueFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorRight,
		DestinationConnectionIndex: AnchorLeft,
		CaseValue: microflows.EnumerationCase{
			Value: "true",
		},
	}
	falseFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorTop,
		CaseValue: microflows.EnumerationCase{
			Value: "false",
		},
	}

	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		splitID: {trueFlow, falseFlow},
	}

	var lines []string
	emitAnchorAnnotation(split, flowsByOrigin, nil, &lines, "")

	if len(lines) != 1 {
		t.Fatalf("expected 1 anchor line, got %d: %v", len(lines), lines)
	}
	out := lines[0]

	for _, want := range []string{
		"true: (from: right, to: left)",
		"false: (from: bottom, to: top)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull: %s", want, out)
		}
	}
}

func TestEmitSplitAnchor_SupportsExpressionCase(t *testing.T) {
	// Mendix splits often use ExpressionCase (Expression == "true" / "false")
	// instead of EnumerationCase. The anchor emission must identify the
	// branches via findBranchFlows, which handles every CaseValue variant the
	// parser produces — otherwise a real project with expression-based
	// branches loses its anchor information on describe.
	splitID := model.ID("split-expr")
	split := &microflows.ExclusiveSplit{}
	split.ID = splitID

	trueFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorRight,
		DestinationConnectionIndex: AnchorLeft,
		CaseValue: &microflows.ExpressionCase{
			Expression: "true",
		},
	}
	falseFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorTop,
		CaseValue: &microflows.ExpressionCase{
			Expression: "false",
		},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		splitID: {trueFlow, falseFlow},
	}

	var lines []string
	emitAnchorAnnotation(split, flowsByOrigin, nil, &lines, "")

	if len(lines) != 1 {
		t.Fatalf("expected 1 anchor line, got %d: %v", len(lines), lines)
	}
	out := lines[0]
	for _, want := range []string{
		"true: (from: right, to: left)",
		"false: (from: bottom, to: top)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull: %s", want, out)
		}
	}
}

func TestEmitSplitAnchor_SupportsBooleanCase(t *testing.T) {
	// BooleanCase uses the literal bool field Value — findBranchFlows maps
	// Value=true → trueFlow, Value=false → falseFlow.
	splitID := model.ID("split-bool")
	split := &microflows.ExclusiveSplit{}
	split.ID = splitID

	trueFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorTop,
		DestinationConnectionIndex: AnchorLeft,
		CaseValue:                  &microflows.BooleanCase{Value: true},
	}
	falseFlow := &microflows.SequenceFlow{
		OriginID:                   splitID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorTop,
		CaseValue:                  &microflows.BooleanCase{Value: false},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		splitID: {trueFlow, falseFlow},
	}

	var lines []string
	emitAnchorAnnotation(split, flowsByOrigin, nil, &lines, "")

	if len(lines) != 1 {
		t.Fatalf("expected 1 anchor line, got %d: %v", len(lines), lines)
	}
	out := lines[0]
	for _, want := range []string{
		"true: (from: top, to: left)",
		"false: (from: bottom, to: top)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull: %s", want, out)
		}
	}
}

func TestEmitSplitAnchor_NoEmissionWhenAllDefaultsAbsent(t *testing.T) {
	// A split with no flows yet (defensive: shouldn't happen in real
	// microflows, but the emission path must not panic or emit an
	// empty parenthesised line).
	splitID := model.ID("split-3")
	split := &microflows.ExclusiveSplit{}
	split.ID = splitID

	var lines []string
	emitAnchorAnnotation(split, map[model.ID][]*microflows.SequenceFlow{}, nil, &lines, "")

	if len(lines) != 0 {
		t.Errorf("expected no anchor line, got %v", lines)
	}
}
