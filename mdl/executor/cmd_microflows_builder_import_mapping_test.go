// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	mdltypes "github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
)

func TestAddImportFromMappingRegistersSingleResultType(t *testing.T) {
	fb := importMappingFlowBuilder(t, "Object")

	fb.addImportFromMappingAction(&ast.ImportFromMappingStmt{
		OutputVariable: "ImportedOrder",
		SourceVariable: "Payload",
		Mapping:        ast.QualifiedName{Module: "Integration", Name: "ImportOrder"},
	})

	if got := fb.varTypes["ImportedOrder"]; got != "Sales.Order" {
		t.Fatalf("ImportedOrder type = %q, want Sales.Order", got)
	}
}

func TestAddImportFromMappingRegistersListResultType(t *testing.T) {
	fb := importMappingFlowBuilder(t, "Array")

	fb.addImportFromMappingAction(&ast.ImportFromMappingStmt{
		OutputVariable: "ImportedOrders",
		SourceVariable: "Payload",
		Mapping:        ast.QualifiedName{Module: "Integration", Name: "ImportOrderList"},
	})

	if got := fb.varTypes["ImportedOrders"]; got != "List of Sales.Order" {
		t.Fatalf("ImportedOrders type = %q, want list of Sales.Order", got)
	}
}

func importMappingFlowBuilder(t *testing.T, rootElementType string) *flowBuilder {
	t.Helper()

	return &flowBuilder{
		varTypes: map[string]string{},
		backend: &mock.MockBackend{
			GetImportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ImportMapping, error) {
				if moduleName != "Integration" {
					return nil, fmt.Errorf("unexpected module %q", moduleName)
				}
				return &model.ImportMapping{
					JsonStructure: "Integration.OrderPayload",
					Elements: []*model.ImportMappingElement{
						{Entity: "Sales.Order"},
					},
				}, nil
			},
			GetJsonStructureByQualifiedNameFunc: func(moduleName, name string) (*mdltypes.JsonStructure, error) {
				if moduleName != "Integration" || name != "OrderPayload" {
					return nil, fmt.Errorf("unexpected json structure %s.%s", moduleName, name)
				}
				return &mdltypes.JsonStructure{
					Elements: []*mdltypes.JsonElement{
						{ElementType: rootElementType},
					},
				}, nil
			},
		},
	}
}
