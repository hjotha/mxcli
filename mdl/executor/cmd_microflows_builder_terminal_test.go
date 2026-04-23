// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestLastStmtIsReturn_EmptyBody(t *testing.T) {
	if lastStmtIsReturn(nil) {
		t.Error("empty body must not be terminal")
	}
}

func TestLastStmtIsReturn_PlainReturn(t *testing.T) {
	body := []ast.MicroflowStatement{&ast.ReturnStmt{}}
	if !lastStmtIsReturn(body) {
		t.Error("body ending in ReturnStmt must be terminal")
	}
}

func TestLastStmtIsReturn_RaiseError(t *testing.T) {
	body := []ast.MicroflowStatement{&ast.RaiseErrorStmt{}}
	if !lastStmtIsReturn(body) {
		t.Error("body ending in RaiseErrorStmt must be terminal")
	}
}

func TestLastStmtIsReturn_IfWithoutElse_NotTerminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
	}
	if lastStmtIsReturn(body) {
		t.Error("IF without ELSE must not be terminal (false path falls through)")
	}
}

func TestLastStmtIsReturn_IfBothBranchesReturn_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("IF/ELSE where both branches return must be terminal")
	}
}

func TestLastStmtIsReturn_IfOnlyThenReturns_NotTerminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.LogStmt{}}, // non-terminal
		},
	}
	if lastStmtIsReturn(body) {
		t.Error("IF/ELSE where only THEN terminates must not be terminal")
	}
}

func TestLastStmtIsReturn_NestedIfChain_Terminal(t *testing.T) {
	// if { return } else if { return } else { return }
	inner := &ast.IfStmt{
		ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
	}
	outer := &ast.IfStmt{
		ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		ElseBody: []ast.MicroflowStatement{inner},
	}
	if !lastStmtIsReturn([]ast.MicroflowStatement{outer}) {
		t.Error("else-if chain where every terminal branch returns must be terminal")
	}
}

func TestLastStmtIsReturn_RaiseErrorMixed_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.RaiseErrorStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("IF/ELSE with return on one side and raise error on the other must be terminal")
	}
}

func TestLastStmtIsReturn_LoopNotTerminal(t *testing.T) {
	// A LOOP whose body only returns is still non-terminal — BREAK can exit.
	body := []ast.MicroflowStatement{
		&ast.LoopStmt{Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
	}
	if lastStmtIsReturn(body) {
		t.Error("LOOP must never be terminal (BREAK path)")
	}
}
