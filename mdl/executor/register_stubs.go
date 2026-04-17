// SPDX-License-Identifier: Apache-2.0

package executor

import "github.com/mendixlabs/mxcli/mdl/ast"

// Handler registration functions — each registers handlers for its domain.
// Handlers are thin wrappers around existing Executor methods. Once handlers
// are migrated to *ExecContext signatures, these wrappers will be replaced
// by direct function references.

func registerConnectionHandlers(r *Registry) {
	r.Register(&ast.ConnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execConnect(stmt.(*ast.ConnectStmt))
	})
	r.Register(&ast.DisconnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDisconnect()
	})
	r.Register(&ast.StatusStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execStatus()
	})
}

func registerModuleHandlers(r *Registry) {
	r.Register(&ast.CreateModuleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateModule(stmt.(*ast.CreateModuleStmt))
	})
	r.Register(&ast.DropModuleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropModule(stmt.(*ast.DropModuleStmt))
	})
}

func registerEnumerationHandlers(r *Registry) {
	r.Register(&ast.CreateEnumerationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateEnumeration(stmt.(*ast.CreateEnumerationStmt))
	})
	r.Register(&ast.AlterEnumerationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterEnumeration(stmt.(*ast.AlterEnumerationStmt))
	})
	r.Register(&ast.DropEnumerationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropEnumeration(stmt.(*ast.DropEnumerationStmt))
	})
}

func registerConstantHandlers(r *Registry) {
	r.Register(&ast.CreateConstantStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createConstant(stmt.(*ast.CreateConstantStmt))
	})
	r.Register(&ast.DropConstantStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropConstant(stmt.(*ast.DropConstantStmt))
	})
}

func registerDatabaseConnectionHandlers(r *Registry) {
	r.Register(&ast.CreateDatabaseConnectionStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createDatabaseConnection(stmt.(*ast.CreateDatabaseConnectionStmt))
	})
}

func registerEntityHandlers(r *Registry) {
	r.Register(&ast.CreateEntityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateEntity(stmt.(*ast.CreateEntityStmt))
	})
	r.Register(&ast.CreateViewEntityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateViewEntity(stmt.(*ast.CreateViewEntityStmt))
	})
	r.Register(&ast.AlterEntityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterEntity(stmt.(*ast.AlterEntityStmt))
	})
	r.Register(&ast.DropEntityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropEntity(stmt.(*ast.DropEntityStmt))
	})
}

func registerAssociationHandlers(r *Registry) {
	r.Register(&ast.CreateAssociationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateAssociation(stmt.(*ast.CreateAssociationStmt))
	})
	r.Register(&ast.AlterAssociationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterAssociation(stmt.(*ast.AlterAssociationStmt))
	})
	r.Register(&ast.DropAssociationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropAssociation(stmt.(*ast.DropAssociationStmt))
	})
}

func registerMicroflowHandlers(r *Registry) {
	r.Register(&ast.CreateMicroflowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateMicroflow(stmt.(*ast.CreateMicroflowStmt))
	})
	r.Register(&ast.DropMicroflowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropMicroflow(stmt.(*ast.DropMicroflowStmt))
	})
}

func registerPageHandlers(r *Registry) {
	r.Register(&ast.CreatePageStmtV3{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreatePageV3(stmt.(*ast.CreatePageStmtV3))
	})
	r.Register(&ast.DropPageStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropPage(stmt.(*ast.DropPageStmt))
	})
	r.Register(&ast.CreateSnippetStmtV3{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateSnippetV3(stmt.(*ast.CreateSnippetStmtV3))
	})
	r.Register(&ast.DropSnippetStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropSnippet(stmt.(*ast.DropSnippetStmt))
	})
	r.Register(&ast.DropJavaActionStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropJavaAction(stmt.(*ast.DropJavaActionStmt))
	})
	r.Register(&ast.CreateJavaActionStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateJavaAction(stmt.(*ast.CreateJavaActionStmt))
	})
	r.Register(&ast.DropFolderStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropFolder(stmt.(*ast.DropFolderStmt))
	})
	r.Register(&ast.MoveFolderStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execMoveFolder(stmt.(*ast.MoveFolderStmt))
	})
	r.Register(&ast.MoveStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execMove(stmt.(*ast.MoveStmt))
	})
	r.Register(&ast.RenameStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRename(stmt.(*ast.RenameStmt))
	})
}

func registerSecurityHandlers(r *Registry) {
	r.Register(&ast.CreateModuleRoleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateModuleRole(stmt.(*ast.CreateModuleRoleStmt))
	})
	r.Register(&ast.DropModuleRoleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropModuleRole(stmt.(*ast.DropModuleRoleStmt))
	})
	r.Register(&ast.CreateUserRoleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateUserRole(stmt.(*ast.CreateUserRoleStmt))
	})
	r.Register(&ast.AlterUserRoleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterUserRole(stmt.(*ast.AlterUserRoleStmt))
	})
	r.Register(&ast.DropUserRoleStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropUserRole(stmt.(*ast.DropUserRoleStmt))
	})
	r.Register(&ast.GrantEntityAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantEntityAccess(stmt.(*ast.GrantEntityAccessStmt))
	})
	r.Register(&ast.RevokeEntityAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokeEntityAccess(stmt.(*ast.RevokeEntityAccessStmt))
	})
	r.Register(&ast.GrantMicroflowAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantMicroflowAccess(stmt.(*ast.GrantMicroflowAccessStmt))
	})
	r.Register(&ast.RevokeMicroflowAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokeMicroflowAccess(stmt.(*ast.RevokeMicroflowAccessStmt))
	})
	r.Register(&ast.GrantPageAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantPageAccess(stmt.(*ast.GrantPageAccessStmt))
	})
	r.Register(&ast.RevokePageAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokePageAccess(stmt.(*ast.RevokePageAccessStmt))
	})
	r.Register(&ast.GrantWorkflowAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantWorkflowAccess(stmt.(*ast.GrantWorkflowAccessStmt))
	})
	r.Register(&ast.RevokeWorkflowAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokeWorkflowAccess(stmt.(*ast.RevokeWorkflowAccessStmt))
	})
	r.Register(&ast.AlterProjectSecurityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterProjectSecurity(stmt.(*ast.AlterProjectSecurityStmt))
	})
	r.Register(&ast.CreateDemoUserStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateDemoUser(stmt.(*ast.CreateDemoUserStmt))
	})
	r.Register(&ast.DropDemoUserStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropDemoUser(stmt.(*ast.DropDemoUserStmt))
	})
	r.Register(&ast.UpdateSecurityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execUpdateSecurity(stmt.(*ast.UpdateSecurityStmt))
	})
}

func registerNavigationHandlers(r *Registry) {
	r.Register(&ast.AlterNavigationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterNavigation(stmt.(*ast.AlterNavigationStmt))
	})
}

func registerImageHandlers(r *Registry) {
	r.Register(&ast.CreateImageCollectionStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateImageCollection(stmt.(*ast.CreateImageCollectionStmt))
	})
	r.Register(&ast.DropImageCollectionStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropImageCollection(stmt.(*ast.DropImageCollectionStmt))
	})
}

func registerWorkflowHandlers(r *Registry) {
	r.Register(&ast.CreateWorkflowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateWorkflow(stmt.(*ast.CreateWorkflowStmt))
	})
	r.Register(&ast.DropWorkflowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropWorkflow(stmt.(*ast.DropWorkflowStmt))
	})
	r.Register(&ast.AlterWorkflowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterWorkflow(stmt.(*ast.AlterWorkflowStmt))
	})
}

func registerBusinessEventHandlers(r *Registry) {
	r.Register(&ast.CreateBusinessEventServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createBusinessEventService(stmt.(*ast.CreateBusinessEventServiceStmt))
	})
	r.Register(&ast.DropBusinessEventServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropBusinessEventService(stmt.(*ast.DropBusinessEventServiceStmt))
	})
}

func registerSettingsHandlers(r *Registry) {
	r.Register(&ast.AlterSettingsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.alterSettings(stmt.(*ast.AlterSettingsStmt))
	})
	r.Register(&ast.CreateConfigurationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createConfiguration(stmt.(*ast.CreateConfigurationStmt))
	})
	r.Register(&ast.DropConfigurationStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropConfiguration(stmt.(*ast.DropConfigurationStmt))
	})
}

func registerODataHandlers(r *Registry) {
	r.Register(&ast.CreateODataClientStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createODataClient(stmt.(*ast.CreateODataClientStmt))
	})
	r.Register(&ast.AlterODataClientStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.alterODataClient(stmt.(*ast.AlterODataClientStmt))
	})
	r.Register(&ast.DropODataClientStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropODataClient(stmt.(*ast.DropODataClientStmt))
	})
	r.Register(&ast.CreateODataServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createODataService(stmt.(*ast.CreateODataServiceStmt))
	})
	r.Register(&ast.AlterODataServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.alterODataService(stmt.(*ast.AlterODataServiceStmt))
	})
	r.Register(&ast.DropODataServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropODataService(stmt.(*ast.DropODataServiceStmt))
	})
}

func registerJSONStructureHandlers(r *Registry) {
	r.Register(&ast.CreateJsonStructureStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateJsonStructure(stmt.(*ast.CreateJsonStructureStmt))
	})
	r.Register(&ast.DropJsonStructureStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropJsonStructure(stmt.(*ast.DropJsonStructureStmt))
	})
}

func registerMappingHandlers(r *Registry) {
	r.Register(&ast.CreateImportMappingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateImportMapping(stmt.(*ast.CreateImportMappingStmt))
	})
	r.Register(&ast.DropImportMappingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropImportMapping(stmt.(*ast.DropImportMappingStmt))
	})
	r.Register(&ast.CreateExportMappingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateExportMapping(stmt.(*ast.CreateExportMappingStmt))
	})
	r.Register(&ast.DropExportMappingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropExportMapping(stmt.(*ast.DropExportMappingStmt))
	})
}

func registerRESTHandlers(r *Registry) {
	r.Register(&ast.CreateRestClientStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createRestClient(stmt.(*ast.CreateRestClientStmt))
	})
	r.Register(&ast.DropRestClientStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.dropRestClient(stmt.(*ast.DropRestClientStmt))
	})
	r.Register(&ast.CreatePublishedRestServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreatePublishedRestService(stmt.(*ast.CreatePublishedRestServiceStmt))
	})
	r.Register(&ast.DropPublishedRestServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropPublishedRestService(stmt.(*ast.DropPublishedRestServiceStmt))
	})
	r.Register(&ast.AlterPublishedRestServiceStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterPublishedRestService(stmt.(*ast.AlterPublishedRestServiceStmt))
	})
	r.Register(&ast.CreateExternalEntityStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateExternalEntity(stmt.(*ast.CreateExternalEntityStmt))
	})
	r.Register(&ast.CreateExternalEntitiesStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.createExternalEntities(stmt.(*ast.CreateExternalEntitiesStmt))
	})
	r.Register(&ast.GrantODataServiceAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantODataServiceAccess(stmt.(*ast.GrantODataServiceAccessStmt))
	})
	r.Register(&ast.RevokeODataServiceAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokeODataServiceAccess(stmt.(*ast.RevokeODataServiceAccessStmt))
	})
	r.Register(&ast.GrantPublishedRestServiceAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execGrantPublishedRestServiceAccess(stmt.(*ast.GrantPublishedRestServiceAccessStmt))
	})
	r.Register(&ast.RevokePublishedRestServiceAccessStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRevokePublishedRestServiceAccess(stmt.(*ast.RevokePublishedRestServiceAccessStmt))
	})
}

func registerDataTransformerHandlers(r *Registry) {
	r.Register(&ast.CreateDataTransformerStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCreateDataTransformer(stmt.(*ast.CreateDataTransformerStmt))
	})
	r.Register(&ast.DropDataTransformerStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDropDataTransformer(stmt.(*ast.DropDataTransformerStmt))
	})
}

func registerQueryHandlers(r *Registry) {
	r.Register(&ast.ShowStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execShow(stmt.(*ast.ShowStmt))
	})
	r.Register(&ast.ShowWidgetsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execShowWidgets(stmt.(*ast.ShowWidgetsStmt))
	})
	r.Register(&ast.UpdateWidgetsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execUpdateWidgets(stmt.(*ast.UpdateWidgetsStmt))
	})
	r.Register(&ast.SelectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execCatalogQuery(stmt.(*ast.SelectStmt).Query)
	})
	r.Register(&ast.DescribeStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDescribe(stmt.(*ast.DescribeStmt))
	})
	r.Register(&ast.DescribeCatalogTableStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDescribeCatalogTable(stmt.(*ast.DescribeCatalogTableStmt))
	})
	// NOTE: ShowFeaturesStmt was missing from the original type-switch
	// (pre-existing bug). Adding it here to fix the dead-code path.
	r.Register(&ast.ShowFeaturesStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execShowFeatures(stmt.(*ast.ShowFeaturesStmt))
	})
}

func registerStylingHandlers(r *Registry) {
	r.Register(&ast.ShowDesignPropertiesStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execShowDesignProperties(stmt.(*ast.ShowDesignPropertiesStmt))
	})
	r.Register(&ast.DescribeStylingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDescribeStyling(stmt.(*ast.DescribeStylingStmt))
	})
	r.Register(&ast.AlterStylingStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterStyling(stmt.(*ast.AlterStylingStmt))
	})
}

func registerRepositoryHandlers(r *Registry) {
	r.Register(&ast.UpdateStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execUpdate()
	})
	r.Register(&ast.RefreshStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRefresh()
	})
	r.Register(&ast.RefreshCatalogStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execRefreshCatalogStmt(stmt.(*ast.RefreshCatalogStmt))
	})
	r.Register(&ast.SearchStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSearch(stmt.(*ast.SearchStmt))
	})
}

func registerSessionHandlers(r *Registry) {
	r.Register(&ast.SetStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSet(stmt.(*ast.SetStmt))
	})
	r.Register(&ast.HelpStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execHelp()
	})
	r.Register(&ast.ExitStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execExit()
	})
	r.Register(&ast.ExecuteScriptStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execExecuteScript(stmt.(*ast.ExecuteScriptStmt))
	})
}

func registerLintHandlers(r *Registry) {
	r.Register(&ast.LintStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execLint(stmt.(*ast.LintStmt))
	})
}

func registerAlterPageHandlers(r *Registry) {
	r.Register(&ast.AlterPageStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execAlterPage(stmt.(*ast.AlterPageStmt))
	})
}

func registerFragmentHandlers(r *Registry) {
	r.Register(&ast.DefineFragmentStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execDefineFragment(stmt.(*ast.DefineFragmentStmt))
	})
	r.Register(&ast.DescribeFragmentFromStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.describeFragmentFrom(stmt.(*ast.DescribeFragmentFromStmt))
	})
}

func registerSQLHandlers(r *Registry) {
	r.Register(&ast.SQLConnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLConnect(stmt.(*ast.SQLConnectStmt))
	})
	r.Register(&ast.SQLDisconnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLDisconnect(stmt.(*ast.SQLDisconnectStmt))
	})
	r.Register(&ast.SQLConnectionsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLConnections()
	})
	r.Register(&ast.SQLQueryStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLQuery(stmt.(*ast.SQLQueryStmt))
	})
	r.Register(&ast.SQLShowTablesStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLShowTables(stmt.(*ast.SQLShowTablesStmt))
	})
	r.Register(&ast.SQLShowViewsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLShowViews(stmt.(*ast.SQLShowViewsStmt))
	})
	r.Register(&ast.SQLShowFunctionsStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLShowFunctions(stmt.(*ast.SQLShowFunctionsStmt))
	})
	r.Register(&ast.SQLDescribeTableStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLDescribeTable(stmt.(*ast.SQLDescribeTableStmt))
	})
	r.Register(&ast.SQLGenerateConnectorStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execSQLGenerateConnector(stmt.(*ast.SQLGenerateConnectorStmt))
	})
}

func registerImportHandlers(r *Registry) {
	r.Register(&ast.ImportStmt{}, func(e *Executor, stmt ast.Statement) error {
		return e.execImport(stmt.(*ast.ImportStmt))
	})
}
