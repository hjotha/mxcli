// SPDX-License-Identifier: Apache-2.0

package executor

import "github.com/mendixlabs/mxcli/mdl/ast"

// Handler registration functions — each registers handlers for its domain.
// Handlers are thin wrappers around existing Executor methods. Once handlers
// are migrated to *ExecContext signatures, these wrappers will be replaced
// by direct function references.

func registerConnectionHandlers(r *Registry) {
	r.Register(&ast.ConnectStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execConnect(ctx, stmt.(*ast.ConnectStmt))
	})
	r.Register(&ast.DisconnectStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execDisconnect(ctx)
	})
	r.Register(&ast.StatusStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execStatus(ctx)
	})
}

func registerModuleHandlers(r *Registry) {
	r.Register(&ast.CreateModuleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execCreateModule(ctx, stmt.(*ast.CreateModuleStmt))
	})
	r.Register(&ast.DropModuleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execDropModule(ctx, stmt.(*ast.DropModuleStmt))
	})
}

func registerEnumerationHandlers(r *Registry) {
	r.Register(&ast.CreateEnumerationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execCreateEnumeration(ctx, stmt.(*ast.CreateEnumerationStmt))
	})
	r.Register(&ast.AlterEnumerationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execAlterEnumeration(ctx, stmt.(*ast.AlterEnumerationStmt))
	})
	r.Register(&ast.DropEnumerationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return execDropEnumeration(ctx, stmt.(*ast.DropEnumerationStmt))
	})
}

func registerConstantHandlers(r *Registry) {
	r.Register(&ast.CreateConstantStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return createConstant(ctx, stmt.(*ast.CreateConstantStmt))
	})
	r.Register(&ast.DropConstantStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return dropConstant(ctx, stmt.(*ast.DropConstantStmt))
	})
}

func registerDatabaseConnectionHandlers(r *Registry) {
	r.Register(&ast.CreateDatabaseConnectionStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createDatabaseConnection(stmt.(*ast.CreateDatabaseConnectionStmt))
	})
}

func registerEntityHandlers(r *Registry) {
	r.Register(&ast.CreateEntityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateEntity(stmt.(*ast.CreateEntityStmt))
	})
	r.Register(&ast.CreateViewEntityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateViewEntity(stmt.(*ast.CreateViewEntityStmt))
	})
	r.Register(&ast.AlterEntityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterEntity(stmt.(*ast.AlterEntityStmt))
	})
	r.Register(&ast.DropEntityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropEntity(stmt.(*ast.DropEntityStmt))
	})
}

func registerAssociationHandlers(r *Registry) {
	r.Register(&ast.CreateAssociationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateAssociation(stmt.(*ast.CreateAssociationStmt))
	})
	r.Register(&ast.AlterAssociationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterAssociation(stmt.(*ast.AlterAssociationStmt))
	})
	r.Register(&ast.DropAssociationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropAssociation(stmt.(*ast.DropAssociationStmt))
	})
}

func registerMicroflowHandlers(r *Registry) {
	r.Register(&ast.CreateMicroflowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateMicroflow(stmt.(*ast.CreateMicroflowStmt))
	})
	r.Register(&ast.DropMicroflowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropMicroflow(stmt.(*ast.DropMicroflowStmt))
	})
}

func registerPageHandlers(r *Registry) {
	r.Register(&ast.CreatePageStmtV3{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreatePageV3(stmt.(*ast.CreatePageStmtV3))
	})
	r.Register(&ast.DropPageStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropPage(stmt.(*ast.DropPageStmt))
	})
	r.Register(&ast.CreateSnippetStmtV3{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateSnippetV3(stmt.(*ast.CreateSnippetStmtV3))
	})
	r.Register(&ast.DropSnippetStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropSnippet(stmt.(*ast.DropSnippetStmt))
	})
	r.Register(&ast.DropJavaActionStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropJavaAction(stmt.(*ast.DropJavaActionStmt))
	})
	r.Register(&ast.CreateJavaActionStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateJavaAction(stmt.(*ast.CreateJavaActionStmt))
	})
	r.Register(&ast.DropFolderStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropFolder(stmt.(*ast.DropFolderStmt))
	})
	r.Register(&ast.MoveFolderStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execMoveFolder(stmt.(*ast.MoveFolderStmt))
	})
	r.Register(&ast.MoveStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execMove(stmt.(*ast.MoveStmt))
	})
	r.Register(&ast.RenameStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRename(stmt.(*ast.RenameStmt))
	})
}

func registerSecurityHandlers(r *Registry) {
	r.Register(&ast.CreateModuleRoleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateModuleRole(stmt.(*ast.CreateModuleRoleStmt))
	})
	r.Register(&ast.DropModuleRoleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropModuleRole(stmt.(*ast.DropModuleRoleStmt))
	})
	r.Register(&ast.CreateUserRoleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateUserRole(stmt.(*ast.CreateUserRoleStmt))
	})
	r.Register(&ast.AlterUserRoleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterUserRole(stmt.(*ast.AlterUserRoleStmt))
	})
	r.Register(&ast.DropUserRoleStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropUserRole(stmt.(*ast.DropUserRoleStmt))
	})
	r.Register(&ast.GrantEntityAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantEntityAccess(stmt.(*ast.GrantEntityAccessStmt))
	})
	r.Register(&ast.RevokeEntityAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokeEntityAccess(stmt.(*ast.RevokeEntityAccessStmt))
	})
	r.Register(&ast.GrantMicroflowAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantMicroflowAccess(stmt.(*ast.GrantMicroflowAccessStmt))
	})
	r.Register(&ast.RevokeMicroflowAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokeMicroflowAccess(stmt.(*ast.RevokeMicroflowAccessStmt))
	})
	r.Register(&ast.GrantPageAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantPageAccess(stmt.(*ast.GrantPageAccessStmt))
	})
	r.Register(&ast.RevokePageAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokePageAccess(stmt.(*ast.RevokePageAccessStmt))
	})
	r.Register(&ast.GrantWorkflowAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantWorkflowAccess(stmt.(*ast.GrantWorkflowAccessStmt))
	})
	r.Register(&ast.RevokeWorkflowAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokeWorkflowAccess(stmt.(*ast.RevokeWorkflowAccessStmt))
	})
	r.Register(&ast.AlterProjectSecurityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterProjectSecurity(stmt.(*ast.AlterProjectSecurityStmt))
	})
	r.Register(&ast.CreateDemoUserStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateDemoUser(stmt.(*ast.CreateDemoUserStmt))
	})
	r.Register(&ast.DropDemoUserStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropDemoUser(stmt.(*ast.DropDemoUserStmt))
	})
	r.Register(&ast.UpdateSecurityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execUpdateSecurity(stmt.(*ast.UpdateSecurityStmt))
	})
}

func registerNavigationHandlers(r *Registry) {
	r.Register(&ast.AlterNavigationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterNavigation(stmt.(*ast.AlterNavigationStmt))
	})
}

func registerImageHandlers(r *Registry) {
	r.Register(&ast.CreateImageCollectionStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateImageCollection(stmt.(*ast.CreateImageCollectionStmt))
	})
	r.Register(&ast.DropImageCollectionStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropImageCollection(stmt.(*ast.DropImageCollectionStmt))
	})
}

func registerWorkflowHandlers(r *Registry) {
	r.Register(&ast.CreateWorkflowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateWorkflow(stmt.(*ast.CreateWorkflowStmt))
	})
	r.Register(&ast.DropWorkflowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropWorkflow(stmt.(*ast.DropWorkflowStmt))
	})
	r.Register(&ast.AlterWorkflowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterWorkflow(stmt.(*ast.AlterWorkflowStmt))
	})
}

func registerBusinessEventHandlers(r *Registry) {
	r.Register(&ast.CreateBusinessEventServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createBusinessEventService(stmt.(*ast.CreateBusinessEventServiceStmt))
	})
	r.Register(&ast.DropBusinessEventServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.dropBusinessEventService(stmt.(*ast.DropBusinessEventServiceStmt))
	})
}

func registerSettingsHandlers(r *Registry) {
	r.Register(&ast.AlterSettingsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.alterSettings(stmt.(*ast.AlterSettingsStmt))
	})
	r.Register(&ast.CreateConfigurationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createConfiguration(stmt.(*ast.CreateConfigurationStmt))
	})
	r.Register(&ast.DropConfigurationStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.dropConfiguration(stmt.(*ast.DropConfigurationStmt))
	})
}

func registerODataHandlers(r *Registry) {
	r.Register(&ast.CreateODataClientStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createODataClient(stmt.(*ast.CreateODataClientStmt))
	})
	r.Register(&ast.AlterODataClientStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.alterODataClient(stmt.(*ast.AlterODataClientStmt))
	})
	r.Register(&ast.DropODataClientStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.dropODataClient(stmt.(*ast.DropODataClientStmt))
	})
	r.Register(&ast.CreateODataServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createODataService(stmt.(*ast.CreateODataServiceStmt))
	})
	r.Register(&ast.AlterODataServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.alterODataService(stmt.(*ast.AlterODataServiceStmt))
	})
	r.Register(&ast.DropODataServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.dropODataService(stmt.(*ast.DropODataServiceStmt))
	})
}

func registerJSONStructureHandlers(r *Registry) {
	r.Register(&ast.CreateJsonStructureStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateJsonStructure(stmt.(*ast.CreateJsonStructureStmt))
	})
	r.Register(&ast.DropJsonStructureStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropJsonStructure(stmt.(*ast.DropJsonStructureStmt))
	})
}

func registerMappingHandlers(r *Registry) {
	r.Register(&ast.CreateImportMappingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateImportMapping(stmt.(*ast.CreateImportMappingStmt))
	})
	r.Register(&ast.DropImportMappingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropImportMapping(stmt.(*ast.DropImportMappingStmt))
	})
	r.Register(&ast.CreateExportMappingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateExportMapping(stmt.(*ast.CreateExportMappingStmt))
	})
	r.Register(&ast.DropExportMappingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropExportMapping(stmt.(*ast.DropExportMappingStmt))
	})
}

func registerRESTHandlers(r *Registry) {
	r.Register(&ast.CreateRestClientStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createRestClient(stmt.(*ast.CreateRestClientStmt))
	})
	r.Register(&ast.DropRestClientStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.dropRestClient(stmt.(*ast.DropRestClientStmt))
	})
	r.Register(&ast.CreatePublishedRestServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreatePublishedRestService(stmt.(*ast.CreatePublishedRestServiceStmt))
	})
	r.Register(&ast.DropPublishedRestServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropPublishedRestService(stmt.(*ast.DropPublishedRestServiceStmt))
	})
	r.Register(&ast.AlterPublishedRestServiceStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterPublishedRestService(stmt.(*ast.AlterPublishedRestServiceStmt))
	})
	r.Register(&ast.CreateExternalEntityStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateExternalEntity(stmt.(*ast.CreateExternalEntityStmt))
	})
	r.Register(&ast.CreateExternalEntitiesStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.createExternalEntities(stmt.(*ast.CreateExternalEntitiesStmt))
	})
	r.Register(&ast.GrantODataServiceAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantODataServiceAccess(stmt.(*ast.GrantODataServiceAccessStmt))
	})
	r.Register(&ast.RevokeODataServiceAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokeODataServiceAccess(stmt.(*ast.RevokeODataServiceAccessStmt))
	})
	r.Register(&ast.GrantPublishedRestServiceAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execGrantPublishedRestServiceAccess(stmt.(*ast.GrantPublishedRestServiceAccessStmt))
	})
	r.Register(&ast.RevokePublishedRestServiceAccessStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRevokePublishedRestServiceAccess(stmt.(*ast.RevokePublishedRestServiceAccessStmt))
	})
}

func registerDataTransformerHandlers(r *Registry) {
	r.Register(&ast.CreateDataTransformerStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCreateDataTransformer(stmt.(*ast.CreateDataTransformerStmt))
	})
	r.Register(&ast.DropDataTransformerStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDropDataTransformer(stmt.(*ast.DropDataTransformerStmt))
	})
}

func registerQueryHandlers(r *Registry) {
	r.Register(&ast.ShowStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execShow(stmt.(*ast.ShowStmt))
	})
	r.Register(&ast.ShowWidgetsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execShowWidgets(stmt.(*ast.ShowWidgetsStmt))
	})
	r.Register(&ast.UpdateWidgetsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execUpdateWidgets(stmt.(*ast.UpdateWidgetsStmt))
	})
	r.Register(&ast.SelectStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execCatalogQuery(stmt.(*ast.SelectStmt).Query)
	})
	r.Register(&ast.DescribeStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDescribe(stmt.(*ast.DescribeStmt))
	})
	r.Register(&ast.DescribeCatalogTableStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDescribeCatalogTable(stmt.(*ast.DescribeCatalogTableStmt))
	})
	// NOTE: ShowFeaturesStmt was missing from the original type-switch
	// (pre-existing bug). Adding it here to fix the dead-code path.
	r.Register(&ast.ShowFeaturesStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execShowFeatures(stmt.(*ast.ShowFeaturesStmt))
	})
}

func registerStylingHandlers(r *Registry) {
	r.Register(&ast.ShowDesignPropertiesStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execShowDesignProperties(stmt.(*ast.ShowDesignPropertiesStmt))
	})
	r.Register(&ast.DescribeStylingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDescribeStyling(stmt.(*ast.DescribeStylingStmt))
	})
	r.Register(&ast.AlterStylingStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterStyling(stmt.(*ast.AlterStylingStmt))
	})
}

func registerRepositoryHandlers(r *Registry) {
	r.Register(&ast.UpdateStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execUpdate()
	})
	r.Register(&ast.RefreshStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRefresh()
	})
	r.Register(&ast.RefreshCatalogStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execRefreshCatalogStmt(stmt.(*ast.RefreshCatalogStmt))
	})
	r.Register(&ast.SearchStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSearch(stmt.(*ast.SearchStmt))
	})
}

func registerSessionHandlers(r *Registry) {
	r.Register(&ast.SetStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSet(stmt.(*ast.SetStmt))
	})
	r.Register(&ast.HelpStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execHelp()
	})
	r.Register(&ast.ExitStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execExit()
	})
	r.Register(&ast.ExecuteScriptStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execExecuteScript(stmt.(*ast.ExecuteScriptStmt))
	})
}

func registerLintHandlers(r *Registry) {
	r.Register(&ast.LintStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execLint(stmt.(*ast.LintStmt))
	})
}

func registerAlterPageHandlers(r *Registry) {
	r.Register(&ast.AlterPageStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execAlterPage(stmt.(*ast.AlterPageStmt))
	})
}

func registerFragmentHandlers(r *Registry) {
	r.Register(&ast.DefineFragmentStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execDefineFragment(stmt.(*ast.DefineFragmentStmt))
	})
	r.Register(&ast.DescribeFragmentFromStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.describeFragmentFrom(stmt.(*ast.DescribeFragmentFromStmt))
	})
}

func registerSQLHandlers(r *Registry) {
	r.Register(&ast.SQLConnectStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLConnect(stmt.(*ast.SQLConnectStmt))
	})
	r.Register(&ast.SQLDisconnectStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLDisconnect(stmt.(*ast.SQLDisconnectStmt))
	})
	r.Register(&ast.SQLConnectionsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLConnections()
	})
	r.Register(&ast.SQLQueryStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLQuery(stmt.(*ast.SQLQueryStmt))
	})
	r.Register(&ast.SQLShowTablesStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLShowTables(stmt.(*ast.SQLShowTablesStmt))
	})
	r.Register(&ast.SQLShowViewsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLShowViews(stmt.(*ast.SQLShowViewsStmt))
	})
	r.Register(&ast.SQLShowFunctionsStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLShowFunctions(stmt.(*ast.SQLShowFunctionsStmt))
	})
	r.Register(&ast.SQLDescribeTableStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLDescribeTable(stmt.(*ast.SQLDescribeTableStmt))
	})
	r.Register(&ast.SQLGenerateConnectorStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execSQLGenerateConnector(stmt.(*ast.SQLGenerateConnectorStmt))
	})
}

func registerImportHandlers(r *Registry) {
	r.Register(&ast.ImportStmt{}, func(ctx *ExecContext, stmt ast.Statement) error {
		return ctx.executor.execImport(stmt.(*ast.ImportStmt))
	})
}
