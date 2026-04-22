// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// TestNestedIfPreservesCaptions is a regression test for the bug where
// the outer IF's @caption would be overwritten by the inner IF's @caption
// because pendingAnnotations is shared mutable state across recursive
// addStatement calls.
//
// Before the fix:
//   - outer ExclusiveSplit received caption "Right format?" (from inner IF)
//   - inner ExclusiveSplit kept its condition expression as caption
//   - inner IF's @annotation got attached to the outer split
//
// After the fix:
//   - addIfStatement consumes its own pendingAnnotations right after
//     creating its split, so outer and inner captions stay bound to the
//     correct splits.
func TestNestedIfPreservesCaptions(t *testing.T) {
	// Build an AST equivalent to:
	//   if $S != empty            @caption 'String not empty?'
	//     if isMatch($S, 'x')     @caption 'Right format?'
	//       return true
	//     else
	//       return false
	//   else
	//     return false
	innerIf := &ast.IfStmt{
		Condition: &ast.FunctionCallExpr{
			Name:      "isMatch",
			Arguments: []ast.Expression{&ast.VariableExpr{Name: "S"}, &ast.LiteralExpr{Value: "x", Kind: ast.LiteralString}},
		},
		ThenBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: true, Kind: ast.LiteralBoolean}},
		},
		ElseBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: false, Kind: ast.LiteralBoolean}},
		},
		Annotations: &ast.ActivityAnnotations{Caption: "Right format?"},
	}
	outerIf := &ast.IfStmt{
		Condition: &ast.BinaryExpr{
			Left:     &ast.VariableExpr{Name: "S"},
			Operator: "!=",
			Right:    &ast.LiteralExpr{Value: nil, Kind: ast.LiteralNull},
		},
		ThenBody: []ast.MicroflowStatement{innerIf},
		ElseBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: false, Kind: ast.LiteralBoolean}},
		},
		Annotations: &ast.ActivityAnnotations{Caption: "String not empty?"},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{"S": "String"},
		declaredVars: map[string]string{"S": "String"},
	}
	fb.buildFlowGraph([]ast.MicroflowStatement{outerIf}, nil)

	// Collect ExclusiveSplits with their captions. The outer split is created
	// first, so objects[1] is the outer split (objects[0] is the StartEvent).
	var splits []*microflows.ExclusiveSplit
	for _, obj := range fb.objects {
		if sp, ok := obj.(*microflows.ExclusiveSplit); ok {
			splits = append(splits, sp)
		}
	}

	if len(splits) != 2 {
		t.Fatalf("expected 2 ExclusiveSplits, got %d", len(splits))
	}

	// Splits are appended in creation order: outer first (from outerIf),
	// then inner (when recursion into ThenBody creates the nested IF's split).
	outerSplit, innerSplit := splits[0], splits[1]

	if outerSplit.Caption != "String not empty?" {
		t.Errorf("outer split caption: got %q, want %q", outerSplit.Caption, "String not empty?")
	}
	if innerSplit.Caption != "Right format?" {
		t.Errorf("inner split caption: got %q, want %q", innerSplit.Caption, "Right format?")
	}
}

// TestIfCaptionWithoutNesting confirms a single IF with @caption still gets
// the right caption after the fix (baseline sanity check).
func TestIfCaptionWithoutNesting(t *testing.T) {
	ifStmt := &ast.IfStmt{
		Condition: &ast.BinaryExpr{
			Left:     &ast.VariableExpr{Name: "S"},
			Operator: "!=",
			Right:    &ast.LiteralExpr{Value: nil, Kind: ast.LiteralNull},
		},
		ThenBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: true, Kind: ast.LiteralBoolean}},
		},
		ElseBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: false, Kind: ast.LiteralBoolean}},
		},
		Annotations: &ast.ActivityAnnotations{Caption: "String not empty?"},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{"S": "String"},
		declaredVars: map[string]string{"S": "String"},
	}
	fb.buildFlowGraph([]ast.MicroflowStatement{ifStmt}, nil)

	for _, obj := range fb.objects {
		if sp, ok := obj.(*microflows.ExclusiveSplit); ok {
			if sp.Caption != "String not empty?" {
				t.Errorf("split caption: got %q, want %q", sp.Caption, "String not empty?")
			}
			return
		}
	}
	t.Fatal("no ExclusiveSplit found")
}

// TestIfAnnotationStaysWithCorrectSplit confirms @annotation on a nested IF
// attaches to that IF's split, not to the outer one.
func TestIfAnnotationStaysWithCorrectSplit(t *testing.T) {
	innerIf := &ast.IfStmt{
		Condition: &ast.FunctionCallExpr{
			Name:      "isMatch",
			Arguments: []ast.Expression{&ast.VariableExpr{Name: "S"}, &ast.LiteralExpr{Value: "x", Kind: ast.LiteralString}},
		},
		ThenBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: true, Kind: ast.LiteralBoolean}},
		},
		ElseBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: false, Kind: ast.LiteralBoolean}},
		},
		Annotations: &ast.ActivityAnnotations{
			Caption:        "Right format?",
			AnnotationText: "Inner IF note",
		},
	}
	outerIf := &ast.IfStmt{
		Condition: &ast.BinaryExpr{
			Left:     &ast.VariableExpr{Name: "S"},
			Operator: "!=",
			Right:    &ast.LiteralExpr{Value: nil, Kind: ast.LiteralNull},
		},
		ThenBody: []ast.MicroflowStatement{innerIf},
		ElseBody: []ast.MicroflowStatement{
			&ast.ReturnStmt{Value: &ast.LiteralExpr{Value: false, Kind: ast.LiteralBoolean}},
		},
		Annotations: &ast.ActivityAnnotations{
			Caption:        "String not empty?",
			AnnotationText: "Outer IF note",
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{"S": "String"},
		declaredVars: map[string]string{"S": "String"},
	}
	fb.buildFlowGraph([]ast.MicroflowStatement{outerIf}, nil)

	var splits []*microflows.ExclusiveSplit
	var annotations []*microflows.Annotation
	for _, obj := range fb.objects {
		switch o := obj.(type) {
		case *microflows.ExclusiveSplit:
			splits = append(splits, o)
		case *microflows.Annotation:
			annotations = append(annotations, o)
		}
	}

	if len(splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(splits))
	}
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	outerSplit, innerSplit := splits[0], splits[1]

	// AnnotationFlow links Annotation -> activity. Verify each flow points
	// from the annotation with the expected text to the expected split.
	var outerNoteDestID, innerNoteDestID string
	for _, af := range fb.annotationFlows {
		// Find the Annotation referenced by OriginID
		for _, ann := range annotations {
			if ann.ID != af.OriginID {
				continue
			}
			switch ann.Caption {
			case "Outer IF note":
				outerNoteDestID = string(af.DestinationID)
			case "Inner IF note":
				innerNoteDestID = string(af.DestinationID)
			}
		}
	}

	if outerNoteDestID != string(outerSplit.ID) {
		t.Errorf("outer note destination: got %q, want %q (outer split)", outerNoteDestID, outerSplit.ID)
	}
	if innerNoteDestID != string(innerSplit.ID) {
		t.Errorf("inner note destination: got %q, want %q (inner split)", innerNoteDestID, innerSplit.ID)
	}
}
