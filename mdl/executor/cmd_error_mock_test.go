// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"github.com/mendixlabs/mxcli/sdk/security"
	"github.com/mendixlabs/mxcli/sdk/workflows"
)

// errBackend is a sentinel used in backend-error tests.
var errBackend = fmt.Errorf("backend failure")

func TestShowEnumerations_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showEnumerations(ctx, ""))
}

func TestShowConstants_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		ListConstantsFunc: func() ([]*model.Constant, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showConstants(ctx, ""))
}

func TestShowMicroflows_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:    func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showMicroflows(ctx, ""))
}

func TestShowNanoflows_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		ListNanoflowsFunc: func() ([]*microflows.Nanoflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showNanoflows(ctx, ""))
}

func TestShowPages_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showPages(ctx, ""))
}

func TestShowSnippets_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:  func() bool { return true },
		ListSnippetsFunc: func() ([]*pages.Snippet, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showSnippets(ctx, ""))
}

func TestShowLayouts_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListLayoutsFunc: func() ([]*pages.Layout, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showLayouts(ctx, ""))
}

func TestShowWorkflows_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		ListWorkflowsFunc: func() ([]*workflows.Workflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showWorkflows(ctx, ""))
}

func TestShowODataClients_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:               func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showODataClients(ctx, ""))
}

func TestShowODataServices_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:                func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showODataServices(ctx, ""))
}

func TestShowRestClients_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:              func() bool { return true },
		ListConsumedRestServicesFunc: func() ([]*model.ConsumedRestService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showRestClients(ctx, ""))
}

func TestShowPublishedRestServices_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:               func() bool { return true },
		ListPublishedRestServicesFunc: func() ([]*model.PublishedRestService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showPublishedRestServices(ctx, ""))
}

func TestShowJavaActions_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:     func() bool { return true },
		ListJavaActionsFunc: func() ([]*types.JavaAction, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showJavaActions(ctx, ""))
}

func TestShowJavaScriptActions_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListJavaScriptActionsFunc: func() ([]*types.JavaScriptAction, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showJavaScriptActions(ctx, ""))
}

func TestShowDatabaseConnections_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:             func() bool { return true },
		ListDatabaseConnectionsFunc: func() ([]*model.DatabaseConnection, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showDatabaseConnections(ctx, ""))
}

func TestShowImageCollections_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showImageCollections(ctx, ""))
}

func TestShowJsonStructures_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListJsonStructuresFunc: func() ([]*types.JsonStructure, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showJsonStructures(ctx, ""))
}

func TestShowNavigation_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		GetNavigationFunc: func() (*types.NavigationDocument, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showNavigation(ctx))
}

func TestShowProjectSecurity_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showProjectSecurity(ctx))
}

func TestShowModuleRoles_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showModuleRoles(ctx, ""))
}

func TestShowUserRoles_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showUserRoles(ctx))
}

func TestShowDemoUsers_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showDemoUsers(ctx))
}

func TestShowBusinessEventServices_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:               func() bool { return true },
		ListBusinessEventServicesFunc: func() ([]*model.BusinessEventService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showBusinessEventServices(ctx, ""))
}

func TestShowAgentEditorModels_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorModelsFunc: func() ([]*agenteditor.Model, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showAgentEditorModels(ctx, ""))
}

func TestShowAgentEditorAgents_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorAgentsFunc: func() ([]*agenteditor.Agent, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showAgentEditorAgents(ctx, ""))
}

func TestShowAgentEditorKnowledgeBases_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:                   func() bool { return true },
		ListAgentEditorKnowledgeBasesFunc: func() ([]*agenteditor.KnowledgeBase, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showAgentEditorKnowledgeBases(ctx, ""))
}

func TestShowAgentEditorMCPServices_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:                        func() bool { return true },
		ListAgentEditorConsumedMCPServicesFunc: func() ([]*agenteditor.ConsumedMCPService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showAgentEditorConsumedMCPServices(ctx, ""))
}

func TestListDataTransformers_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListDataTransformersFunc: func() ([]*model.DataTransformer, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listDataTransformers(ctx, ""))
}

func TestShowExportMappings_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListExportMappingsFunc: func() ([]*model.ExportMapping, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showExportMappings(ctx, ""))
}

func TestShowImportMappings_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListImportMappingsFunc: func() ([]*model.ImportMapping, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showImportMappings(ctx, ""))
}

func TestShowSettings_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSettingsFunc: func() (*model.ProjectSettings, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, showSettings(ctx))
}

// Describe handler backend errors

func TestDescribeEnumeration_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeEnumeration(ctx, ast.QualifiedName{Module: "M", Name: "E"}))
}

func TestDescribeConstant_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		ListConstantsFunc: func() ([]*model.Constant, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeConstant(ctx, ast.QualifiedName{Module: "M", Name: "C"}))
}

func TestDescribeMicroflow_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:    func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeMicroflow(ctx, ast.QualifiedName{Module: "M", Name: "F"}))
}

func TestDescribeWorkflow_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		ListWorkflowsFunc: func() ([]*workflows.Workflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeWorkflow(ctx, ast.QualifiedName{Module: "M", Name: "W"}))
}

func TestDescribeNavigation_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:   func() bool { return true },
		GetNavigationFunc: func() (*types.NavigationDocument, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeNavigation(ctx, ast.QualifiedName{Module: "M", Name: "N"}))
}

func TestDescribeODataClient_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:               func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeODataClient(ctx, ast.QualifiedName{Module: "M", Name: "C"}))
}

func TestDescribeODataService_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:                func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeODataService(ctx, ast.QualifiedName{Module: "M", Name: "S"}))
}

func TestDescribeRestClient_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:              func() bool { return true },
		ListConsumedRestServicesFunc: func() ([]*model.ConsumedRestService, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeRestClient(ctx, ast.QualifiedName{Module: "M", Name: "R"}))
}

func TestDescribeImageCollection_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeImageCollection(ctx, ast.QualifiedName{Module: "M", Name: "I"}))
}

func TestDescribeDatabaseConnection_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:             func() bool { return true },
		ListDatabaseConnectionsFunc: func() ([]*model.DatabaseConnection, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeDatabaseConnection(ctx, ast.QualifiedName{Module: "M", Name: "D"}))
}

func TestDescribeModuleRole_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeModuleRole(ctx, ast.QualifiedName{Module: "M", Name: "R"}))
}

func TestDescribeUserRole_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeUserRole(ctx, ast.QualifiedName{Module: "", Name: "Admin"}))
}

func TestDescribeDemoUser_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeDemoUser(ctx, "demo"))
}

// Write handler backend errors

func TestExecCreateModule_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execCreateModule(ctx, &ast.CreateModuleStmt{Name: "M"}))
}

func TestExecCreateEnumeration_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execCreateEnumeration(ctx, &ast.CreateEnumerationStmt{
		Name: ast.QualifiedName{Module: "M", Name: "E"},
	}))
}

func TestExecDropMicroflow_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:    func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropMicroflow(ctx, &ast.DropMicroflowStmt{
		Name: ast.QualifiedName{Module: "M", Name: "F"},
	}))
}

func TestExecDropPage_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropPage(ctx, &ast.DropPageStmt{
		Name: ast.QualifiedName{Module: "M", Name: "P"},
	}))
}

func TestExecDropSnippet_Mock_BackendError(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc:  func() bool { return true },
		ListSnippetsFunc: func() ([]*pages.Snippet, error) { return nil, errBackend },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropSnippet(ctx, &ast.DropSnippetStmt{
		Name: ast.QualifiedName{Module: "M", Name: "S"},
	}))
}
