// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestConnect_Local(t *testing.T) {
	input := `CONNECT LOCAL '/path/to/project';`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	stmt, ok := prog.Statements[0].(*ast.ConnectStmt)
	if !ok {
		t.Fatalf("Expected ConnectStmt, got %T", prog.Statements[0])
	}
	if stmt.Path != "/path/to/project" {
		t.Errorf("Expected /path/to/project, got %q", stmt.Path)
	}
}

func TestDisconnect(t *testing.T) {
	input := `DISCONNECT;`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	_, ok := prog.Statements[0].(*ast.DisconnectStmt)
	if !ok {
		t.Fatalf("Expected DisconnectStmt, got %T", prog.Statements[0])
	}
}

func TestStatus_BareKeyword(t *testing.T) {
	prog, errs := Build(`STATUS;`)
	if len(errs) > 0 {
		t.Fatalf("Parse errors for bare STATUS: %v", errs)
	}
	if _, ok := prog.Statements[0].(*ast.StatusStmt); !ok {
		t.Fatalf("Expected StatusStmt, got %T", prog.Statements[0])
	}
}

func TestStatus_ShowStatus(t *testing.T) {
	prog, errs := Build(`SHOW STATUS;`)
	if len(errs) > 0 {
		t.Fatalf("Parse errors for SHOW STATUS: %v", errs)
	}
	if _, ok := prog.Statements[0].(*ast.StatusStmt); !ok {
		t.Fatalf("Expected StatusStmt for SHOW STATUS, got %T", prog.Statements[0])
	}
}

func TestStatus_ShowCatalogStatusUnaffected(t *testing.T) {
	prog, errs := Build(`SHOW CATALOG STATUS;`)
	if len(errs) > 0 {
		t.Fatalf("Parse errors for SHOW CATALOG STATUS: %v", errs)
	}
	stmt, ok := prog.Statements[0].(*ast.ShowStmt)
	if !ok {
		t.Fatalf("Expected ShowStmt, got %T", prog.Statements[0])
	}
	if stmt.ObjectType != ast.ShowCatalogStatus {
		t.Errorf("Expected ShowCatalogStatus, got %v", stmt.ObjectType)
	}
}
