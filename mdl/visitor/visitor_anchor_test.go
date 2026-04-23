// SPDX-License-Identifier: Apache-2.0

// Tests for @anchor annotation parsing — sets OriginConnectionIndex /
// DestinationConnectionIndex hints on microflow statements so that DESCRIBE →
// re-execute round-trips preserve the visual side on which each SequenceFlow
// attaches.
package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func firstStatement(t *testing.T, src string) ast.MicroflowStatement {
	t.Helper()
	input := "create microflow MfTest.M () begin\n" + src + "\nend;"
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse: %v", e)
		}
		t.FailNow()
	}
	if len(prog.Statements) == 0 {
		t.Fatal("no statements parsed")
	}
	mf := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if len(mf.Body) == 0 {
		t.Fatal("no body statements")
	}
	return mf.Body[0]
}

func TestAnchorAnnotation_FromAndTo(t *testing.T) {
	stmt := firstStatement(t, "@anchor(from: right, to: left)\nlog info node 'App' 'hi';")

	log, ok := stmt.(*ast.LogStmt)
	if !ok {
		t.Fatalf("expected LogStmt, got %T", stmt)
	}
	if log.Annotations == nil || log.Annotations.Anchor == nil {
		t.Fatal("expected Anchor annotation to be set")
	}
	if log.Annotations.Anchor.From != ast.AnchorSideRight {
		t.Errorf("From: got %v, want Right", log.Annotations.Anchor.From)
	}
	if log.Annotations.Anchor.To != ast.AnchorSideLeft {
		t.Errorf("To: got %v, want Left", log.Annotations.Anchor.To)
	}
}

func TestAnchorAnnotation_OnlyFrom(t *testing.T) {
	stmt := firstStatement(t, "@anchor(from: bottom)\nlog info node 'App' 'hi';")
	log := stmt.(*ast.LogStmt)
	if log.Annotations.Anchor == nil {
		t.Fatal("expected Anchor")
	}
	if log.Annotations.Anchor.From != ast.AnchorSideBottom {
		t.Errorf("From: got %v, want Bottom", log.Annotations.Anchor.From)
	}
	if log.Annotations.Anchor.To != ast.AnchorSideUnset {
		t.Errorf("To: got %v, want Unset", log.Annotations.Anchor.To)
	}
}

func TestAnchorAnnotation_AllFourSides(t *testing.T) {
	cases := []struct {
		name string
		side string
		want ast.AnchorSide
	}{
		{"top", "top", ast.AnchorSideTop},
		{"right", "right", ast.AnchorSideRight},
		{"bottom", "bottom", ast.AnchorSideBottom},
		{"left", "left", ast.AnchorSideLeft},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stmt := firstStatement(t, "@anchor(from: "+tc.side+")\nlog info node 'App' 'hi';")
			log := stmt.(*ast.LogStmt)
			if log.Annotations.Anchor.From != tc.want {
				t.Errorf("From for %s: got %v, want %v", tc.side, log.Annotations.Anchor.From, tc.want)
			}
		})
	}
}

func TestAnchorAnnotation_SplitBranches(t *testing.T) {
	src := `@anchor(true: (from: right, to: left), false: (from: bottom, to: top))
if true then
  log info node 'App' 'yes';
else
  log info node 'App' 'no';
end if;`
	stmt := firstStatement(t, src)
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", stmt)
	}
	if ifStmt.Annotations == nil {
		t.Fatal("expected Annotations")
	}
	if ifStmt.Annotations.TrueBranchAnchor == nil {
		t.Fatal("expected TrueBranchAnchor")
	}
	if ifStmt.Annotations.TrueBranchAnchor.From != ast.AnchorSideRight ||
		ifStmt.Annotations.TrueBranchAnchor.To != ast.AnchorSideLeft {
		t.Errorf("true branch: got from=%v to=%v; want right/left",
			ifStmt.Annotations.TrueBranchAnchor.From, ifStmt.Annotations.TrueBranchAnchor.To)
	}
	if ifStmt.Annotations.FalseBranchAnchor == nil {
		t.Fatal("expected FalseBranchAnchor")
	}
	if ifStmt.Annotations.FalseBranchAnchor.From != ast.AnchorSideBottom ||
		ifStmt.Annotations.FalseBranchAnchor.To != ast.AnchorSideTop {
		t.Errorf("false branch: got from=%v to=%v; want bottom/top",
			ifStmt.Annotations.FalseBranchAnchor.From, ifStmt.Annotations.FalseBranchAnchor.To)
	}
}

func TestAnchorAnnotation_MissingLeavesUnset(t *testing.T) {
	// A statement without @anchor should have Annotations.Anchor == nil,
	// signalling "builder picks the default".
	stmt := firstStatement(t, "log info node 'App' 'hi';")
	log := stmt.(*ast.LogStmt)
	if log.Annotations != nil && log.Annotations.Anchor != nil {
		t.Errorf("expected Anchor nil on statement without @anchor, got %+v", log.Annotations.Anchor)
	}
}
