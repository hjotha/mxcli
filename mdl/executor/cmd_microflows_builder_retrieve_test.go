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
	deploymentTargetID := model.ID("deployment-target-entity")
	appViewID := model.ID("application-view-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleApps" {
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
						{BaseElement: model.BaseElement{ID: deploymentTargetID}, Name: "DeploymentTarget"},
						{BaseElement: model.BaseElement{ID: appViewID}, Name: "ApplicationView"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "DeploymentTarget_ApplicationView",
							ParentID: deploymentTargetID,
							ChildID:  appViewID,
							Type:     domainmodel.AssociationTypeReference,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable: "DeploymentTargetList",
		Source: ast.QualifiedName{
			Module: "SampleApps",
			Name:   "DeploymentTarget",
		},
		Where: &ast.SourceExpr{
			Source: "SampleApps.DeploymentTarget_ApplicationView/SampleApps.ApplicationView/SampleApps.ApplicationView_Company = $Company",
		},
		SortColumns: []ast.SortColumnDef{
			{Attribute: "SampleApps.ApplicationView.CreatedAt", Order: "DESC"},
			{Attribute: "SampleApps.ApplicationView.Name", Order: "ASC"},
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
	if got := source.Sorting[0].AttributeQualifiedName; got != "SampleApps.ApplicationView.CreatedAt" {
		t.Fatalf("first sort attribute = %q", got)
	}
	if got := source.Sorting[0].EntityRefSteps; len(got) != 1 || got[0].Association != "SampleApps.DeploymentTarget_ApplicationView" || got[0].DestinationEntity != "SampleApps.ApplicationView" {
		t.Fatalf("first sort entity ref steps = %#v", got)
	}
	if got := source.Sorting[0].Direction; got != microflows.SortDirectionDescending {
		t.Fatalf("first sort direction = %q, want %q", got, microflows.SortDirectionDescending)
	}
	if got := source.Sorting[1].AttributeQualifiedName; got != "SampleApps.ApplicationView.Name" {
		t.Fatalf("second sort attribute = %q", got)
	}
}

func TestAddRetrieveAction_CompactReverseReferenceUsesAssociationSource(t *testing.T) {
	moduleID := model.ID("academy-module")
	profileID := model.ID("profile-entity")
	certificateID := model.ID("certificate-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"Iteratorcertificates": "SampleLearning.SampleItem",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleLearning" {
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
						{BaseElement: model.BaseElement{ID: certificateID}, Name: "SampleItem"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "TopSampleItem_UserProfile",
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
		Variable:      "UserToUpdateSample",
		StartVariable: "Iteratorcertificates",
		Source: ast.QualifiedName{
			Module: "SampleLearning",
			Name:   "TopSampleItem_UserProfile",
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
	if source.AssociationQualifiedName != "SampleLearning.TopSampleItem_UserProfile" {
		t.Fatalf("AssociationQualifiedName = %q", source.AssociationQualifiedName)
	}
	if got := fb.varTypes["UserToUpdateSample"]; got != "List of SampleLearning.userprofiles" {
		t.Fatalf("var type = %q, want List of SampleLearning.userprofiles", got)
	}
}

func TestAddRetrieveAction_ForwardReferenceRegistersSingleTargetEntityType(t *testing.T) {
	moduleID := model.ID("academy-module")
	profileID := model.ID("profile-entity")
	certificateID := model.ID("certificate-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"UserProfile": "SampleLearning.userprofiles",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleLearning" {
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
						{BaseElement: model.BaseElement{ID: certificateID}, Name: "SampleItem"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "TopSampleItem_UserProfile",
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
		Variable:      "SampleItem",
		StartVariable: "UserProfile",
		Source: ast.QualifiedName{
			Module: "SampleLearning",
			Name:   "TopSampleItem_UserProfile",
		},
	})

	if got := fb.varTypes["SampleItem"]; got != "SampleLearning.SampleItem" {
		t.Fatalf("var type = %q, want SampleLearning.SampleItem", got)
	}
}

func TestAddRetrieveAction_ReferenceSetRegistersTargetEntityListType(t *testing.T) {
	moduleID := model.ID("sample-events-module")
	coreEventID := model.ID("source-event-entity")
	metaDataEventID := model.ID("derived-event-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"IteratorSourceEvent": "SampleEvents.SourceEvent",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleEvents" {
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
						{BaseElement: model.BaseElement{ID: coreEventID}, Name: "SourceEvent"},
						{BaseElement: model.BaseElement{ID: metaDataEventID}, Name: "DerivedEvent"},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "SourceEvent_DerivedEvent",
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
		Variable:      "DerivedEventList",
		StartVariable: "IteratorSourceEvent",
		Source: ast.QualifiedName{
			Module: "SampleEvents",
			Name:   "SourceEvent_DerivedEvent",
		},
	})

	if got := fb.varTypes["DerivedEventList"]; got != "List of SampleEvents.DerivedEvent" {
		t.Fatalf("var type = %q, want List of SampleEvents.DerivedEvent", got)
	}
}

func TestAddRetrieveAction_CrossReferenceSetRegistersRemoteTargetEntityListType(t *testing.T) {
	moduleID := model.ID("sample-events-module")
	coreEventID := model.ID("source-event-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"IteratorSourceEvent": "SampleEvents.SourceEvent",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleEvents" {
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
						{BaseElement: model.BaseElement{ID: coreEventID}, Name: "SourceEvent"},
					},
					CrossAssociations: []*domainmodel.CrossModuleAssociation{
						{
							Name:     "SourceEvent_DerivedEvent",
							ParentID: coreEventID,
							ChildRef: "SampleEventIntegration.DerivedEvent",
							Type:     domainmodel.AssociationTypeReferenceSet,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "DerivedEvent",
		StartVariable: "IteratorSourceEvent",
		Source: ast.QualifiedName{
			Module: "SampleEvents",
			Name:   "SourceEvent_DerivedEvent",
		},
	})

	if got := fb.varTypes["DerivedEvent"]; got != "List of SampleEventIntegration.DerivedEvent" {
		t.Fatalf("var type = %q, want List of SampleEventIntegration.DerivedEvent", got)
	}
}

func TestAddRetrieveAction_ReverseReferenceFromSubtypeUsesParentListType(t *testing.T) {
	moduleID := model.ID("system-module")
	messageID := model.ID("message-entity")
	responseID := model.ID("response-entity")
	headerID := model.ID("header-entity")
	fb := &flowBuilder{
		varTypes: map[string]string{
			"Response": "SampleSystem.Response",
		},
		backend: &mock.MockBackend{
			GetModuleByNameFunc: func(name string) (*model.Module, error) {
				if name != "SampleSystem" {
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
						{BaseElement: model.BaseElement{ID: messageID}, Name: "Message"},
						{
							BaseElement:      model.BaseElement{ID: responseID},
							Name:             "Response",
							GeneralizationID: messageID,
						},
						{
							BaseElement: model.BaseElement{ID: headerID},
							Name:        "Header",
							Attributes:  []*domainmodel.Attribute{{Name: "Key"}},
						},
					},
					Associations: []*domainmodel.Association{
						{
							Name:     "Message_Headers",
							ParentID: headerID,
							ChildID:  messageID,
							Type:     domainmodel.AssociationTypeReference,
						},
					},
				}, nil
			},
		},
	}

	fb.addRetrieveAction(&ast.RetrieveStmt{
		Variable:      "HeaderList",
		StartVariable: "Response",
		Source:        ast.QualifiedName{Module: "SampleSystem", Name: "Message_Headers"},
	})

	if got := fb.varTypes["HeaderList"]; got != "List of SampleSystem.Header" {
		t.Fatalf("var type = %q, want List of SampleSystem.Header", got)
	}

	fb.addListOperationAction(&ast.ListOperationStmt{
		OutputVariable: "ContentTypeHeaders",
		Operation:      ast.ListOpFilter,
		InputVariable:  "HeaderList",
		Condition: &ast.BinaryExpr{
			Left:     &ast.IdentifierExpr{Name: "Key"},
			Operator: "=",
			Right:    &ast.LiteralExpr{Kind: ast.LiteralString, Value: "Content-Type"},
		},
	})

	activity := fb.objects[len(fb.objects)-1].(*microflows.ActionActivity)
	action := activity.Action.(*microflows.ListOperationAction)
	op, ok := action.Operation.(*microflows.FilterByAttributeOperation)
	if !ok {
		t.Fatalf("operation type = %T, want *FilterByAttributeOperation", action.Operation)
	}
	if op.Attribute != "SampleSystem.Header.Key" {
		t.Fatalf("filter attribute = %q, want SampleSystem.Header.Key", op.Attribute)
	}
}
