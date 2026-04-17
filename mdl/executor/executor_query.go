// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
)

func execShow(ctx *ExecContext, s *ast.ShowStmt) error {
	if !ctx.Connected() && s.ObjectType != ast.ShowModules && s.ObjectType != ast.ShowFragments {
		return mdlerrors.NewNotConnected()
	}

	switch s.ObjectType {
	case ast.ShowModules:
		return showModules(ctx)
	case ast.ShowEnumerations:
		return showEnumerations(ctx, s.InModule)
	case ast.ShowConstants:
		return showConstants(ctx, s.InModule)
	case ast.ShowConstantValues:
		return showConstantValues(ctx, s.InModule)
	case ast.ShowEntities:
		return showEntities(ctx, s.InModule)
	case ast.ShowEntity:
		return showEntity(ctx, s.Name)
	case ast.ShowAssociations:
		return showAssociations(ctx, s.InModule)
	case ast.ShowAssociation:
		return showAssociation(ctx, s.Name)
	case ast.ShowMicroflows:
		return showMicroflows(ctx, s.InModule)
	case ast.ShowNanoflows:
		return showNanoflows(ctx, s.InModule)
	case ast.ShowPages:
		return showPages(ctx, s.InModule)
	case ast.ShowSnippets:
		return showSnippets(ctx, s.InModule)
	case ast.ShowLayouts:
		return showLayouts(ctx, s.InModule)
	case ast.ShowJavaActions:
		return showJavaActions(ctx, s.InModule)
	case ast.ShowJavaScriptActions:
		return showJavaScriptActions(ctx, s.InModule)
	case ast.ShowVersion:
		return showVersion(ctx)
	case ast.ShowCatalogTables:
		return execShowCatalogTables(ctx)
	case ast.ShowCatalogStatus:
		return execShowCatalogStatus(ctx)
	case ast.ShowCallers:
		return execShowCallers(ctx, s)
	case ast.ShowCallees:
		return execShowCallees(ctx, s)
	case ast.ShowReferences:
		return execShowReferences(ctx, s)
	case ast.ShowImpact:
		return execShowImpact(ctx, s)
	case ast.ShowContext:
		return execShowContext(ctx, s)
	case ast.ShowProjectSecurity:
		return showProjectSecurity(ctx)
	case ast.ShowModuleRoles:
		return showModuleRoles(ctx, s.InModule)
	case ast.ShowUserRoles:
		return showUserRoles(ctx)
	case ast.ShowDemoUsers:
		return showDemoUsers(ctx)
	case ast.ShowAccessOn:
		return showAccessOnEntity(ctx, s.Name)
	case ast.ShowAccessOnMicroflow:
		return showAccessOnMicroflow(ctx, s.Name)
	case ast.ShowAccessOnPage:
		return showAccessOnPage(ctx, s.Name)
	case ast.ShowAccessOnWorkflow:
		return showAccessOnWorkflow(ctx, s.Name)
	case ast.ShowSecurityMatrix:
		return showSecurityMatrix(ctx, s.InModule)
	case ast.ShowODataClients:
		return showODataClients(ctx, s.InModule)
	case ast.ShowODataServices:
		return showODataServices(ctx, s.InModule)
	case ast.ShowExternalEntities:
		return showExternalEntities(ctx, s.InModule)
	case ast.ShowExternalActions:
		return showExternalActions(ctx, s.InModule)
	case ast.ShowNavigation:
		return showNavigation(ctx)
	case ast.ShowNavigationMenu:
		return showNavigationMenu(ctx, s.Name)
	case ast.ShowNavigationHomes:
		return showNavigationHomes(ctx)
	case ast.ShowStructure:
		return execShowStructure(ctx, s)
	case ast.ShowWorkflows:
		return showWorkflows(ctx, s.InModule)
	case ast.ShowBusinessEventServices:
		return showBusinessEventServices(ctx, s.InModule)
	case ast.ShowBusinessEventClients:
		return showBusinessEventClients(ctx, s.InModule)
	case ast.ShowBusinessEvents:
		return showBusinessEvents(ctx, s.InModule)
	case ast.ShowSettings:
		return showSettings(ctx)
	case ast.ShowFragments:
		return showFragments(ctx)
	case ast.ShowDatabaseConnections:
		return showDatabaseConnections(ctx, s.InModule)
	case ast.ShowImageCollections:
		return showImageCollections(ctx, s.InModule)
	case ast.ShowModels:
		return showAgentEditorModels(ctx, s.InModule)
	case ast.ShowAgents:
		return showAgentEditorAgents(ctx, s.InModule)
	case ast.ShowKnowledgeBases:
		return showAgentEditorKnowledgeBases(ctx, s.InModule)
	case ast.ShowConsumedMCPServices:
		return showAgentEditorConsumedMCPServices(ctx, s.InModule)
	case ast.ShowRestClients:
		return showRestClients(ctx, s.InModule)
	case ast.ShowPublishedRestServices:
		return showPublishedRestServices(ctx, s.InModule)
	case ast.ShowDataTransformers:
		return listDataTransformers(ctx, s.InModule)
	case ast.ShowContractEntities:
		return showContractEntities(ctx, s.Name)
	case ast.ShowContractActions:
		return showContractActions(ctx, s.Name)
	case ast.ShowContractChannels:
		return showContractChannels(ctx, s.Name)
	case ast.ShowContractMessages:
		return showContractMessages(ctx, s.Name)
	case ast.ShowJsonStructures:
		return showJsonStructures(ctx, s.InModule)
	case ast.ShowImportMappings:
		return showImportMappings(ctx, s.InModule)
	case ast.ShowExportMappings:
		return showExportMappings(ctx, s.InModule)
	default:
		return mdlerrors.NewUnsupported("unknown show object type")
	}
}

func execDescribe(ctx *ExecContext, s *ast.DescribeStmt) error {
	if !ctx.Connected() && s.ObjectType != ast.DescribeFragment {
		return mdlerrors.NewNotConnected()
	}

	// Determine the object type label and name for JSON wrapping.
	objectType := describeObjectTypeLabel(s.ObjectType)
	name := s.Name.String()

	return writeDescribeJSON(ctx, name, objectType, func() error {
		switch s.ObjectType {
		case ast.DescribeEnumeration:
			return describeEnumeration(ctx, s.Name)
		case ast.DescribeEntity:
			return describeEntity(ctx, s.Name)
		case ast.DescribeAssociation:
			return describeAssociation(ctx, s.Name)
		case ast.DescribeMicroflow:
			return describeMicroflow(ctx, s.Name)
		case ast.DescribeNanoflow:
			return describeNanoflow(ctx, s.Name)
		case ast.DescribeModule:
			return describeModule(ctx, s.Name.Module, s.WithAll)
		case ast.DescribePage:
			return describePage(ctx, s.Name)
		case ast.DescribeSnippet:
			return describeSnippet(ctx, s.Name)
		case ast.DescribeLayout:
			return describeLayout(ctx, s.Name)
		case ast.DescribeConstant:
			return describeConstant(ctx, s.Name)
		case ast.DescribeJavaAction:
			return describeJavaAction(ctx, s.Name)
		case ast.DescribeJavaScriptAction:
			return describeJavaScriptAction(ctx, s.Name)
		case ast.DescribeModuleRole:
			return describeModuleRole(ctx, s.Name)
		case ast.DescribeUserRole:
			return describeUserRole(ctx, s.Name)
		case ast.DescribeDemoUser:
			return describeDemoUser(ctx, s.Name.Name)
		case ast.DescribeODataClient:
			return describeODataClient(ctx, s.Name)
		case ast.DescribeODataService:
			return describeODataService(ctx, s.Name)
		case ast.DescribeExternalEntity:
			return describeExternalEntity(ctx, s.Name)
		case ast.DescribeNavigation:
			return describeNavigation(ctx, s.Name)
		case ast.DescribeWorkflow:
			return describeWorkflow(ctx, s.Name)
		case ast.DescribeBusinessEventService:
			return describeBusinessEventService(ctx, s.Name)
		case ast.DescribeDatabaseConnection:
			return describeDatabaseConnection(ctx, s.Name)
		case ast.DescribeSettings:
			return describeSettings(ctx)
		case ast.DescribeFragment:
			return describeFragment(ctx, s.Name)
		case ast.DescribeImageCollection:
			return describeImageCollection(ctx, s.Name)
		case ast.DescribeModel:
			return describeAgentEditorModel(ctx, s.Name)
		case ast.DescribeAgent:
			return describeAgentEditorAgent(ctx, s.Name)
		case ast.DescribeKnowledgeBase:
			return describeAgentEditorKnowledgeBase(ctx, s.Name)
		case ast.DescribeConsumedMCPService:
			return describeAgentEditorConsumedMCPService(ctx, s.Name)
		case ast.DescribeRestClient:
			return describeRestClient(ctx, s.Name)
		case ast.DescribePublishedRestService:
			return describePublishedRestService(ctx, s.Name)
		case ast.DescribeDataTransformer:
			return describeDataTransformer(ctx, s.Name)
		case ast.DescribeContractEntity:
			return describeContractEntity(ctx, s.Name, s.Format)
		case ast.DescribeContractAction:
			return describeContractAction(ctx, s.Name, s.Format)
		case ast.DescribeContractMessage:
			return describeContractMessage(ctx, s.Name)
		case ast.DescribeJsonStructure:
			return describeJsonStructure(ctx, s.Name)
		case ast.DescribeImportMapping:
			return describeImportMapping(ctx, s.Name)
		case ast.DescribeExportMapping:
			return describeExportMapping(ctx, s.Name)
		default:
			return mdlerrors.NewUnsupported("unknown describe object type")
		}
	})
}

// describeObjectTypeLabel returns a human-readable label for a describe object type.
func describeObjectTypeLabel(t ast.DescribeObjectType) string {
	switch t {
	case ast.DescribeEnumeration:
		return "enumeration"
	case ast.DescribeEntity:
		return "entity"
	case ast.DescribeAssociation:
		return "association"
	case ast.DescribeMicroflow:
		return "microflow"
	case ast.DescribeNanoflow:
		return "nanoflow"
	case ast.DescribeModule:
		return "module"
	case ast.DescribePage:
		return "page"
	case ast.DescribeSnippet:
		return "snippet"
	case ast.DescribeLayout:
		return "layout"
	case ast.DescribeConstant:
		return "constant"
	case ast.DescribeJavaAction:
		return "javaaction"
	case ast.DescribeJavaScriptAction:
		return "javascriptaction"
	case ast.DescribeModuleRole:
		return "modulerole"
	case ast.DescribeUserRole:
		return "userrole"
	case ast.DescribeDemoUser:
		return "demouser"
	case ast.DescribeODataClient:
		return "odataclient"
	case ast.DescribeODataService:
		return "odataservice"
	case ast.DescribeExternalEntity:
		return "externalentity"
	case ast.DescribeNavigation:
		return "navigation"
	case ast.DescribeWorkflow:
		return "workflow"
	case ast.DescribeBusinessEventService:
		return "businesseventservice"
	case ast.DescribeDatabaseConnection:
		return "databaseconnection"
	case ast.DescribeSettings:
		return "settings"
	case ast.DescribeFragment:
		return "fragment"
	case ast.DescribeImageCollection:
		return "imagecollection"
	case ast.DescribeModel:
		return "model"
	case ast.DescribeAgent:
		return "agent"
	case ast.DescribeKnowledgeBase:
		return "knowledgebase"
	case ast.DescribeConsumedMCPService:
		return "consumedmcpservice"
	case ast.DescribeRestClient:
		return "restclient"
	case ast.DescribePublishedRestService:
		return "publishedrestservice"
	case ast.DescribeDataTransformer:
		return "datatransformer"
	case ast.DescribeContractEntity:
		return "contractentity"
	case ast.DescribeContractAction:
		return "contractaction"
	case ast.DescribeContractMessage:
		return "contractmessage"
	case ast.DescribeJsonStructure:
		return "jsonstructure"
	case ast.DescribeImportMapping:
		return "importmapping"
	case ast.DescribeExportMapping:
		return "exportmapping"
	default:
		return "unknown"
	}
}
