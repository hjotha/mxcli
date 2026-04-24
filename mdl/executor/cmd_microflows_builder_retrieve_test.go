// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestAddRetrieveAction_AllowsAssociationPathSortAttribute(t *testing.T) {
	moduleID := model.ID("apps-combined-view-module")
	privateCloudEnvironmentID := model.ID("private-cloud-environment-entity")
	appViewID := model.ID("app-view-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "AppsCombinedView" {
					return nil, nil
				}
				return &model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: name}, nil
			},
			GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
				if id != moduleID {
					return nil, nil
				}
				return &domainmodel.DomainModel{
					ContainerID: moduleID,
					Entities: []*domainmodel.Entity{
						{BaseElement: model.BaseElement{ID: privateCloudEnvironmentID}, Name: "PrivateCloudEnvironment"},
						{BaseElement: model.BaseElement{ID: appViewID}, Name: "AppView"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "PrivateCloudEnvironment_AppView",
							ParentID: privateCloudEnvironmentID,
							ChildID:  appViewID,
							Type:     domainmodel.AssociationTypeReference,
						},
					},
				}, nil
			},
		},
	}

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
	if got := source.Sorting[0].EntityRefSteps; len(got) != 1 || got[0].Association != "AppsCombinedView.PrivateCloudEnvironment_AppView" || got[0].DestinationEntity != "AppsCombinedView.AppView" {
		t.Fatalf("first sort entity ref steps = %#v", got)
	}
	if got := source.Sorting[0].Direction; got != microflows.SortDirectionDescending {
		t.Fatalf("first sort direction = %q, want %q", got, microflows.SortDirectionDescending)
	}
	if got := source.Sorting[1].AttributeQualifiedName; got != "AppsCombinedView.AppView.AppName" {
		t.Fatalf("second sort attribute = %q", got)
	}
}

func TestAddRetrieveAction_CompactReverseReferenceUsesAssociationSource(t *testing.T) {
	moduleID := model.ID("academy-module")
	profileID := model.ID("profile-entity")
	certificateID := model.ID("certificate-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"Iteratorcertificates": "AcademyIntegration.Certificate",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "AcademyIntegration" {
					return nil, nil
				}
				return &model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: name}, nil
			},
			GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
				if id != moduleID {
					return nil, nil
				}
				return &domainmodel.DomainModel{
					ContainerID: moduleID,
					Entities: []*domainmodel.Entity{
						{BaseElement: model.BaseElement{ID: profileID}, Name: "userprofiles"},
						{BaseElement: model.BaseElement{ID: certificateID}, Name: "Certificate"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "HighestCertificate_UserProfile",
							ParentID: profileID,
							ChildID:  certificateID,
							Type:     domainmodel.AssociationTypeReference,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "UserToUpdateCertificate",
		StartVariable: "Iteratorcertificates",
		Source: ast.QualifiedName{
			Module: "AcademyIntegration",
			Name:   "HighestCertificate_UserProfile",
		},
	})

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
	source, ok := action.Source.(*microflows.AssociationRetrieveSource)
	if !ok {
		t.Fatalf("got source %T, want *microflows.AssociationRetrieveSource", action.Source)
	}
	if source.StartVariable != "Iteratorcertificates" {
		t.Fatalf("StartVariable = %q, want Iteratorcertificates", source.StartVariable)
	}
	if source.AssociationQualifiedName != "AcademyIntegration.HighestCertificate_UserProfile" {
		t.Fatalf("AssociationQualifiedName = %q", source.AssociationQualifiedName)
	}
	if got := fb.varTypes["UserToUpdateCertificate"]; got != "List of AcademyIntegration.userprofiles" {
		t.Fatalf("var type = %q, want List of AcademyIntegration.userprofiles", got)
	}
}

func TestAddRetrieveAction_ForwardReferenceRegistersSingleTargetEntityType(t *testing.T) {
	moduleID := model.ID("academy-module")
	profileID := model.ID("profile-entity")
	certificateID := model.ID("certificate-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"UserProfile": "AcademyIntegration.userprofiles",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "AcademyIntegration" {
					return nil, nil
				}
				return &model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: name}, nil
			},
			GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
				if id != moduleID {
					return nil, nil
				}
				return &domainmodel.DomainModel{
					ContainerID: moduleID,
					Entities: []*domainmodel.Entity{
						{BaseElement: model.BaseElement{ID: profileID}, Name: "userprofiles"},
						{BaseElement: model.BaseElement{ID: certificateID}, Name: "Certificate"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "HighestCertificate_UserProfile",
							ParentID: profileID,
							ChildID:  certificateID,
							Type:     domainmodel.AssociationTypeReference,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "Certificate",
		StartVariable: "UserProfile",
		Source: ast.QualifiedName{
			Module: "AcademyIntegration",
			Name:   "HighestCertificate_UserProfile",
		},
	})

	if got := fb.varTypes["Certificate"]; got != "AcademyIntegration.Certificate" {
		t.Fatalf("var type = %q, want AcademyIntegration.Certificate", got)
	}
}

func TestAddRetrieveAction_ReferenceSetRegistersTargetEntityListType(t *testing.T) {
	moduleID := model.ID("data-lake-module")
	coreEventID := model.ID("core-event-entity")
	metaDataEventID := model.ID("metadata-event-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"IteratorCoreEvent": "DataLake.CoreEvent",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "DataLake" {
					return nil, nil
				}
				return &model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: name}, nil
			},
			GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
				if id != moduleID {
					return nil, nil
				}
				return &domainmodel.DomainModel{
					ContainerID: moduleID,
					Entities: []*domainmodel.Entity{
						{BaseElement: model.BaseElement{ID: coreEventID}, Name: "CoreEvent"},
						{BaseElement: model.BaseElement{ID: metaDataEventID}, Name: "MetaDataEvent"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "CoreEvent_MetaDataEvent",
							ParentID: coreEventID,
							ChildID:  metaDataEventID,
							Type:     domainmodel.AssociationTypeReferenceSet,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "MetaDataEventList",
		StartVariable: "IteratorCoreEvent",
		Source: ast.QualifiedName{
			Module: "DataLake",
			Name:   "CoreEvent_MetaDataEvent",
		},
	})

	if got := fb.varTypes["MetaDataEventList"]; got != "List of DataLake.MetaDataEvent" {
		t.Fatalf("var type = %q, want List of DataLake.MetaDataEvent", got)
	}
}

func TestAddRetrieveAction_CrossReferenceSetRegistersRemoteTargetEntityListType(t *testing.T) {
	moduleID := model.ID("data-lake-module")
	coreEventID := model.ID("core-event-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"IteratorCoreEvent": "DataLake.CoreEvent",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "DataLake" {
					return nil, nil
				}
				return &model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: name}, nil
			},
			GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
				if id != moduleID {
					return nil, nil
				}
				return &domainmodel.DomainModel{
					ContainerID: moduleID,
					Entities: []*domainmodel.Entity{
						{BaseElement: model.BaseElement{ID: coreEventID}, Name: "CoreEvent"},
					},
					CrossAssociations: []*domainmodel.CrossModuleAssociation{
						{
							Name:     "CoreEvent_MetaDataEvent",
							ParentID: coreEventID,
							ChildRef: "DatalakeIntegration.MetaDataEvent",
							Type:     domainmodel.AssociationTypeReferenceSet,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "MetaDataEvent",
		StartVariable: "IteratorCoreEvent",
		Source: ast.QualifiedName{
			Module: "DataLake",
			Name:   "CoreEvent_MetaDataEvent",
		},
	})

	if got := fb.varTypes["MetaDataEvent"]; got != "List of DatalakeIntegration.MetaDataEvent" {
		t.Fatalf("var type = %q, want List of DatalakeIntegration.MetaDataEvent", got)
	}
}
