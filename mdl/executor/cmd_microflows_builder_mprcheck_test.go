// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildChangeObject_EmptyChangeRefreshesInClient(t *testing.T) {
	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}

	id := fb.addChangeObjectAction(&ast.ChangeObjectStmt{Variable: "Company"})
	if id == "" || len(fb.objects) != 1 {
		t.Fatalf("expected one change object activity, got id=%q objects=%d", id, len(fb.objects))
	}
	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("object type = %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.ChangeObjectAction)
	if !ok {
		t.Fatalf("action type = %T, want *microflows.ChangeObjectAction", activity.Action)
	}
	if !action.RefreshInClient {
		t.Fatal("empty change object must refresh in client; otherwise Mendix rejects it as CE0032")
	}
}

func TestBuildListFind_AttributeEqualsExpressionUsesAttributeOperation(t *testing.T) {
	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
		varTypes: map[string]string{
			"CertificateList": "List of AcademyIntegration.Certificate",
		},
	}

	id := fb.addListOperationAction(&ast.ListOperationStmt{
		OutputVariable: "ExistingCertificate",
		Operation:      ast.ListOpFind,
		InputVariable:  "CertificateList",
		Condition: &ast.BinaryExpr{
			Left:     &ast.IdentifierExpr{Name: "UUID"},
			Operator: "=",
			Right: &ast.AttributePathExpr{
				Variable: "IteratorCertificate",
				Path:     []string{"Certificate_ID"},
			},
		},
	})
	if id == "" || len(fb.objects) != 1 {
		t.Fatalf("expected one list operation activity, got id=%q objects=%d", id, len(fb.objects))
	}
	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("object type = %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.ListOperationAction)
	if !ok {
		t.Fatalf("action type = %T, want *microflows.ListOperationAction", activity.Action)
	}
	op, ok := action.Operation.(*microflows.FindByAttributeOperation)
	if !ok {
		t.Fatalf("operation type = %T, want *microflows.FindByAttributeOperation", action.Operation)
	}
	if op.Attribute != "AcademyIntegration.Certificate.UUID" {
		t.Fatalf("Attribute = %q, want AcademyIntegration.Certificate.UUID", op.Attribute)
	}
	if op.Expression != "$IteratorCertificate/Certificate_ID" {
		t.Fatalf("Expression = %q, want $IteratorCertificate/Certificate_ID", op.Expression)
	}
}
