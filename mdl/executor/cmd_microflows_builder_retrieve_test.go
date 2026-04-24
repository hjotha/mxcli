// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestAddRetrieveAction_AllowsAssociationPathSortAttribute(t *testing.T) {
	fb := &flowBuilder{varTypes: map[string]string{}}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable: "PrivateCloudEnvironmentList",
		Source: ast.QualifiedName{
			Module: "AppsCombinedView",
			Name:   "PrivateCloudEnvironment",
		},
		Where: &ast.SourceExpr{
			Source: "AppsCombinedView.PrivateCloudEnvironment_AppView/AppsCombinedView.AppView/AppsCombinedView.AppView_Company = $Company",
		},
		SortColumns: []ast.SortColumnDef{
			{Attribute: "AppsCombinedView.AppView.AppCreatedDate", Order: "DESC"},
			{Attribute: "AppsCombinedView.AppView.AppName", Order: "ASC"},
		},
	})

	if len(fb.errors) > 0 {
		t.Fatalf("unexpected builder errors: %v", fb.errors)
	}
	if len(fb.objects) != 1 {
		t.Fatalf("got %d objects, want 1", len(fb.objects))
	}

	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("got object %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.RetrieveAction)
	if !ok {
		t.Fatalf("got action %T, want *microflows.RetrieveAction", activity.Action)
	}
	source, ok := action.Source.(*microflows.DatabaseRetrieveSource)
	if !ok {
		t.Fatalf("got source %T, want *microflows.DatabaseRetrieveSource", action.Source)
	}
	if len(source.Sorting) != 2 {
		t.Fatalf("got %d sort items, want 2", len(source.Sorting))
	}
	if got := source.Sorting[0].AttributeQualifiedName; got != "AppsCombinedView.AppView.AppCreatedDate" {
		t.Fatalf("first sort attribute = %q", got)
	}
	if got := source.Sorting[0].Direction; got != microflows.SortDirectionDescending {
		t.Fatalf("first sort direction = %q, want %q", got, microflows.SortDirectionDescending)
	}
	if got := source.Sorting[1].AttributeQualifiedName; got != "AppsCombinedView.AppView.AppName" {
		t.Fatalf("second sort attribute = %q", got)
	}
}
