// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestValidateMicroflow_EnumSplitAllBranchesReturn(t *testing.T) {
	stmt := &ast.CreateMicroflowStmt{
		Name: ast.QualifiedName{Module: "Sample", Name: "Route"},
		ReturnType: &ast.MicroflowReturnType{
			Type: ast.DataType{Kind: ast.TypeBoolean},
		},
		Body: []ast.MicroflowStatement{
			&ast.EnumSplitStmt{
				Variable: "Status",
				Cases: []ast.EnumSplitCase{
					{Values: []string{"Open"}, Body: []ast.MicroflowStatement{
						&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
					}},
					{Values: []string{"Closed"}, Body: []ast.MicroflowStatement{
						&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: false}},
					}},
				},
				ElseBody: []ast.MicroflowStatement{
					&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: false}},
				},
			},
		},
	}

	violations := ValidateMicroflow(stmt)
	for _, violation := range violations {
		if violation.RuleID == "MDL003" {
			t.Fatalf("ENUM split with exhaustive returning branches must satisfy return validation: %#v", violation)
		}
	}
}

func TestValidateMicroflow_EnumSplitBranchScopedVariable(t *testing.T) {
	stmt := &ast.CreateMicroflowStmt{
		Name: ast.QualifiedName{Module: "Sample", Name: "Route"},
		Body: []ast.MicroflowStatement{
			&ast.EnumSplitStmt{
				Variable: "Status",
				Cases: []ast.EnumSplitCase{
					{Values: []string{"Open"}, Body: []ast.MicroflowStatement{
						&ast.DeclareStmt{Variable: "OnlyInsideCase", Type: ast.DataType{Kind: ast.TypeString}},
					}},
				},
			},
			&ast.MfSetStmt{
				Target: "OnlyInsideCase",
				Value:  &ast.LiteralExpr{Kind: ast.LiteralString, Value: "outside"},
			},
		},
	}

	violations := ValidateMicroflow(stmt)
	for _, violation := range violations {
		if violation.RuleID == "MDL005" {
			return
		}
	}
	t.Fatalf("expected MDL005 for variable declared inside ENUM split branch, got %#v", violations)
}

func TestValidateMicroflowBody_EnumSplitRejectsMoreThanSupportedBranches(t *testing.T) {
	stmt := &ast.CreateMicroflowStmt{
		Name: ast.QualifiedName{Module: "Sample", Name: "Route"},
		Body: []ast.MicroflowStatement{
			enumSplitWithBranchCount(maxEnumSplitBranches + 1),
		},
	}

	errors := strings.Join(ValidateMicroflowBody(stmt), "\n")
	if !strings.Contains(errors, "enum split has 17 branches; at most 16 branches are supported") {
		t.Fatalf("expected unsupported branch count error, got %q", errors)
	}
}
