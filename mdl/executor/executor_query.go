// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func (e *Executor) execShow(s *ast.ShowStmt) error {
	if e.reader == nil && s.ObjectType != ast.ShowModules && s.ObjectType != ast.ShowFragments {
		return fmt.Errorf("not connected to a project")
	}

	switch s.ObjectType {
	case ast.ShowModules:
		return e.showModules()
	case ast.ShowEnumerations:
		return e.showEnumerations(s.InModule)
	case ast.ShowConstants:
		return e.showConstants(s.InModule)
	case ast.ShowConstantValues:
		return e.showConstantValues(s.InModule)
	case ast.ShowEntities:
		return e.showEntities(s.InModule)
	case ast.ShowEntity:
		return e.showEntity(s.Name)
	case ast.ShowAssociations:
		return e.showAssociations(s.InModule)
	case ast.ShowAssociation:
		return e.showAssociation(s.Name)
	case ast.ShowMicroflows:
		return e.showMicroflows(s.InModule)
	case ast.ShowNanoflows:
		return e.showNanoflows(s.InModule)
	case ast.ShowPages:
		return e.showPages(s.InModule)
	case ast.ShowSnippets:
		return e.showSnippets(s.InModule)
	case ast.ShowLayouts:
		return e.showLayouts(s.InModule)
	case ast.ShowJavaActions:
		return e.showJavaActions(s.InModule)
	case ast.ShowJavaScriptActions:
		return e.showJavaScriptActions(s.InModule)
	case ast.ShowVersion:
		return e.showVersion()
	case ast.ShowCatalogTables:
		return e.execShowCatalogTables()
	case ast.ShowCatalogStatus:
		return e.execShowCatalogStatus()
	case ast.ShowCallers:
		return e.execShowCallers(s)
	case ast.ShowCallees:
		return e.execShowCallees(s)
	case ast.ShowReferences:
		return e.execShowReferences(s)
	case ast.ShowImpact:
		return e.execShowImpact(s)
	case ast.ShowContext:
		return e.execShowContext(s)
	case ast.ShowProjectSecurity:
		return e.showProjectSecurity()
	case ast.ShowModuleRoles:
		return e.showModuleRoles(s.InModule)
	case ast.ShowUserRoles:
		return e.showUserRoles()
	case ast.ShowDemoUsers:
		return e.showDemoUsers()
	case ast.ShowAccessOn:
		return e.showAccessOnEntity(s.Name)
	case ast.ShowAccessOnMicroflow:
		return e.showAccessOnMicroflow(s.Name)
	case ast.ShowAccessOnPage:
		return e.showAccessOnPage(s.Name)
	case ast.ShowAccessOnWorkflow:
		return e.showAccessOnWorkflow(s.Name)
	case ast.ShowSecurityMatrix:
		return e.showSecurityMatrix(s.InModule)
	case ast.ShowODataClients:
		return e.showODataClients(s.InModule)
	case ast.ShowODataServices:
		return e.showODataServices(s.InModule)
	case ast.ShowExternalEntities:
		return e.showExternalEntities(s.InModule)
	case ast.ShowExternalActions:
		return e.showExternalActions(s.InModule)
	case ast.ShowNavigation:
		return e.showNavigation()
	case ast.ShowNavigationMenu:
		return e.showNavigationMenu(s.Name)
	case ast.ShowNavigationHomes:
		return e.showNavigationHomes()
	case ast.ShowStructure:
		return e.execShowStructure(s)
	case ast.ShowWorkflows:
		return e.showWorkflows(s.InModule)
	case ast.ShowBusinessEventServices:
		return e.showBusinessEventServices(s.InModule)
	case ast.ShowBusinessEventClients:
		return e.showBusinessEventClients(s.InModule)
	case ast.ShowBusinessEvents:
		return e.showBusinessEvents(s.InModule)
	case ast.ShowSettings:
		return e.showSettings()
	case ast.ShowFragments:
		return e.showFragments()
	case ast.ShowDatabaseConnections:
		return e.showDatabaseConnections(s.InModule)
	case ast.ShowImageCollections:
		return e.showImageCollections(s.InModule)
	case ast.ShowRestClients:
		return e.showRestClients(s.InModule)
	case ast.ShowPublishedRestServices:
		return e.showPublishedRestServices(s.InModule)
	case ast.ShowContractEntities:
		return e.showContractEntities(s.Name)
	case ast.ShowContractActions:
		return e.showContractActions(s.Name)
	case ast.ShowContractChannels:
		return e.showContractChannels(s.Name)
	case ast.ShowContractMessages:
		return e.showContractMessages(s.Name)
	case ast.ShowJsonStructures:
		return e.showJsonStructures(s.InModule)
	case ast.ShowImportMappings:
		return e.showImportMappings(s.InModule)
	case ast.ShowExportMappings:
		return e.showExportMappings(s.InModule)
	default:
		return fmt.Errorf("unknown show object type")
	}
}

func (e *Executor) execDescribe(s *ast.DescribeStmt) error {
	if e.reader == nil && s.ObjectType != ast.DescribeFragment {
		return fmt.Errorf("not connected to a project")
	}

	switch s.ObjectType {
	case ast.DescribeEnumeration:
		return e.describeEnumeration(s.Name)
	case ast.DescribeEntity:
		return e.describeEntity(s.Name)
	case ast.DescribeAssociation:
		return e.describeAssociation(s.Name)
	case ast.DescribeMicroflow:
		return e.describeMicroflow(s.Name)
	case ast.DescribeNanoflow:
		return e.describeNanoflow(s.Name)
	case ast.DescribeModule:
		return e.describeModule(s.Name.Module, s.WithAll)
	case ast.DescribePage:
		return e.describePage(s.Name)
	case ast.DescribeSnippet:
		return e.describeSnippet(s.Name)
	case ast.DescribeLayout:
		return e.describeLayout(s.Name)
	case ast.DescribeConstant:
		return e.describeConstant(s.Name)
	case ast.DescribeJavaAction:
		return e.describeJavaAction(s.Name)
	case ast.DescribeJavaScriptAction:
		return e.describeJavaScriptAction(s.Name)
	case ast.DescribeModuleRole:
		return e.describeModuleRole(s.Name)
	case ast.DescribeUserRole:
		return e.describeUserRole(s.Name)
	case ast.DescribeDemoUser:
		return e.describeDemoUser(s.Name.Name)
	case ast.DescribeODataClient:
		return e.describeODataClient(s.Name)
	case ast.DescribeODataService:
		return e.describeODataService(s.Name)
	case ast.DescribeExternalEntity:
		return e.describeExternalEntity(s.Name)
	case ast.DescribeNavigation:
		return e.describeNavigation(s.Name)
	case ast.DescribeWorkflow:
		return e.describeWorkflow(s.Name)
	case ast.DescribeBusinessEventService:
		return e.describeBusinessEventService(s.Name)
	case ast.DescribeDatabaseConnection:
		return e.describeDatabaseConnection(s.Name)
	case ast.DescribeSettings:
		return e.describeSettings()
	case ast.DescribeFragment:
		return e.describeFragment(s.Name)
	case ast.DescribeImageCollection:
		return e.describeImageCollection(s.Name)
	case ast.DescribeRestClient:
		return e.describeRestClient(s.Name)
	case ast.DescribePublishedRestService:
		return e.describePublishedRestService(s.Name)
	case ast.DescribeContractEntity:
		return e.describeContractEntity(s.Name, s.Format)
	case ast.DescribeContractAction:
		return e.describeContractAction(s.Name, s.Format)
	case ast.DescribeContractMessage:
		return e.describeContractMessage(s.Name)
	case ast.DescribeJsonStructure:
		return e.describeJsonStructure(s.Name)
	case ast.DescribeImportMapping:
		return e.describeImportMapping(s.Name)
	case ast.DescribeExportMapping:
		return e.describeExportMapping(s.Name)
	default:
		return fmt.Errorf("unknown describe object type")
	}
}
