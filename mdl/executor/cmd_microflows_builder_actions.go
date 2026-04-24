// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow builder: CRUD & data actions
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// addCreateVariableAction creates a DECLARE statement as a CreateVariableAction.
func (fb *flowBuilder) addCreateVariableAction(s *ast.DeclareStmt) model.ID {
	// Resolve TypeEnumeration → TypeEntity ambiguity using the domain model
	declType := s.Type
	if declType.Kind == ast.TypeEnumeration && declType.EnumRef != nil && fb.backend != nil {
		if fb.isEntity(declType.EnumRef.Module, declType.EnumRef.Name) {
			declType = ast.DataType{Kind: ast.TypeEntity, EntityRef: declType.EnumRef}
		}
	}

	// Register the variable as declared
	typeName := declType.Kind.String()
	fb.declaredVars[s.Variable] = typeName

	action := &microflows.CreateVariableAction{
		BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
		VariableName: s.Variable,
		DataType:     convertASTToMicroflowDataType(declType, nil),
		InitialValue: fb.exprToString(s.InitialValue),
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addChangeVariableAction creates a SET statement as a ChangeVariableAction.
func (fb *flowBuilder) addChangeVariableAction(s *ast.MfSetStmt) model.ID {
	// Validate that the variable has been declared
	if !fb.isVariableDeclared(s.Target) {
		fb.addErrorWithExample(
			fmt.Sprintf("variable '%s' is not declared", s.Target),
			errorExampleDeclareVariable(s.Target))
	}

	action := &microflows.ChangeVariableAction{
		BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
		VariableName: s.Target,
		Value:        fb.exprToString(s.Value),
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

func (fb *flowBuilder) addInheritanceSplit(s *ast.InheritanceSplitStmt) model.ID {
	if len(s.Cases) > 0 || len(s.ElseBody) > 0 {
		return fb.addStructuredInheritanceSplit(s)
	}

	split := &microflows.InheritanceSplit{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX, Y: fb.posY},
			Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
		},
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
		VariableName:      s.Variable,
	}

	fb.objects = append(fb.objects, split)
	fb.posX += fb.spacing
	return split.ID
}

func (fb *flowBuilder) addStructuredInheritanceSplit(s *ast.InheritanceSplitStmt) model.ID {
	splitX := fb.posX
	centerY := fb.posY
	split := &microflows.InheritanceSplit{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: splitX, Y: centerY},
			Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
		},
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
		VariableName:      s.Variable,
	}
	fb.objects = append(fb.objects, split)

	splitID := split.ID
	if fb.pendingAnnotations != nil {
		fb.applyAnnotations(splitID, fb.pendingAnnotations)
		fb.pendingAnnotations = nil
	}

	savedEndsWithReturn := fb.endsWithReturn
	fb.endsWithReturn = false
	allBranchesReturn := true
	branchStartX := splitX + ActivityWidth + HorizontalSpacing/2
	branchIndex := 0

	addBranch := func(caseValue string, body []ast.MicroflowStatement) {
		branchY := centerY + branchIndex*VerticalSpacing
		branchIndex++
		if len(body) == 0 {
			allBranchesReturn = false
			return
		}

		fb.posX = branchStartX
		fb.posY = branchY
		fb.endsWithReturn = false

		var lastID model.ID
		var prevAnchor *ast.FlowAnchors
		for _, stmt := range body {
			thisAnchor := stmtOwnAnchor(stmt)
			actID := fb.addStatement(stmt)
			if actID == "" {
				continue
			}
			fb.applyPendingAnnotations(actID)
			if lastID == "" {
				var flow *microflows.SequenceFlow
				if branchIndex == 1 {
					flow = newHorizontalFlowWithInheritanceCase(splitID, actID, caseValue)
				} else {
					flow = newDownwardFlowWithInheritanceCase(splitID, actID, caseValue)
				}
				if thisAnchor != nil && thisAnchor.To != ast.AnchorSideUnset {
					flow.DestinationConnectionIndex = int(thisAnchor.To)
				}
				fb.flows = append(fb.flows, flow)
			} else {
				flow := newHorizontalFlow(lastID, actID)
				applyUserAnchors(flow, prevAnchor, thisAnchor)
				fb.flows = append(fb.flows, flow)
			}
			prevAnchor = thisAnchor
			if fb.nextConnectionPoint != "" {
				lastID = fb.nextConnectionPoint
				fb.nextConnectionPoint = ""
			} else {
				lastID = actID
			}
		}
		if !lastStmtIsReturn(body) {
			allBranchesReturn = false
		}
	}

	for _, c := range s.Cases {
		addBranch(qualifiedNameString(c.Entity), c.Body)
	}
	if len(s.ElseBody) > 0 {
		addBranch("", s.ElseBody)
	}

	fb.posX = branchStartX + fb.measurer.measureStatements(appendInheritanceBodies(s)).Width + HorizontalSpacing/2
	fb.posY = centerY
	fb.endsWithReturn = savedEndsWithReturn
	if allBranchesReturn {
		fb.endsWithReturn = true
	}
	return splitID
}

func appendInheritanceBodies(s *ast.InheritanceSplitStmt) []ast.MicroflowStatement {
	var stmts []ast.MicroflowStatement
	for _, c := range s.Cases {
		stmts = append(stmts, c.Body...)
	}
	stmts = append(stmts, s.ElseBody...)
	return stmts
}

func qualifiedNameString(qn ast.QualifiedName) string {
	if qn.Module == "" {
		return qn.Name
	}
	return qn.Module + "." + qn.Name
}

func (fb *flowBuilder) addCastAction(s *ast.CastObjectStmt) model.ID {
	action := &microflows.CastAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		ObjectVariable: s.ObjectVariable,
		OutputVariable: s.OutputVariable,
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addCreateObjectAction creates a CREATE OBJECT statement.
func (fb *flowBuilder) addCreateObjectAction(s *ast.CreateObjectStmt) model.ID {
	action := &microflows.CreateObjectAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		OutputVariable: s.Variable,
		Commit:         microflows.CommitTypeNo,
	}
	// Set entity reference as qualified name (BY_NAME_REFERENCE)
	entityQN := ""
	if s.EntityType.Module != "" && s.EntityType.Name != "" {
		entityQN = s.EntityType.Module + "." + s.EntityType.Name
		action.EntityQualifiedName = entityQN
	}

	// Register variable type for CHANGE statements
	if fb.varTypes != nil && entityQN != "" {
		fb.varTypes[s.Variable] = entityQN
	}

	// Build InitialMembers for each SET assignment
	for _, change := range s.Changes {
		memberChange := &microflows.MemberChange{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Type:        microflows.MemberChangeTypeSet,
			Value:       fb.memberExpressionToString(change.Value, entityQN, change.Attribute),
		}
		fb.resolveMemberChange(memberChange, change.Attribute, entityQN)
		action.InitialMembers = append(action.InitialMembers, memberChange)
	}

	activityX := fb.posX
	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
			ErrorHandlingType:   convertErrorHandlingType(s.ErrorHandling),
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing

	// Build custom error handler flow if present
	if s.ErrorHandling != nil && len(s.ErrorHandling.Body) > 0 {
		errorY := fb.posY + VerticalSpacing
		mergeID := fb.addErrorHandlerFlow(activity.ID, activityX, s.ErrorHandling.Body)
		fb.handleErrorHandlerMerge(mergeID, activity.ID, errorY)
	}

	return activity.ID
}

// addCommitAction creates a COMMIT statement.
func (fb *flowBuilder) addCommitAction(s *ast.MfCommitStmt) model.ID {
	action := &microflows.CommitObjectsAction{
		BaseElement:       model.BaseElement{ID: model.ID(types.GenerateID())},
		ErrorHandlingType: convertErrorHandlingType(s.ErrorHandling),
		CommitVariable:    s.Variable,
		WithEvents:        s.WithEvents,
		RefreshInClient:   s.RefreshInClient,
	}

	activityX := fb.posX
	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing

	// Build custom error handler flow if present
	if s.ErrorHandling != nil && len(s.ErrorHandling.Body) > 0 {
		errorY := fb.posY + VerticalSpacing
		mergeID := fb.addErrorHandlerFlow(activity.ID, activityX, s.ErrorHandling.Body)
		fb.handleErrorHandlerMerge(mergeID, activity.ID, errorY)
	}

	return activity.ID
}

// addDeleteAction creates a DELETE statement.
func (fb *flowBuilder) addDeleteAction(s *ast.DeleteObjectStmt) model.ID {
	action := &microflows.DeleteObjectAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		DeleteVariable: s.Variable,
	}

	activityX := fb.posX
	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
			ErrorHandlingType:   convertErrorHandlingType(s.ErrorHandling),
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing

	// Build custom error handler flow if present
	if s.ErrorHandling != nil && len(s.ErrorHandling.Body) > 0 {
		errorY := fb.posY + VerticalSpacing
		mergeID := fb.addErrorHandlerFlow(activity.ID, activityX, s.ErrorHandling.Body)
		fb.handleErrorHandlerMerge(mergeID, activity.ID, errorY)
	}

	return activity.ID
}

// addRollbackAction creates a ROLLBACK statement.
func (fb *flowBuilder) addRollbackAction(s *ast.RollbackStmt) model.ID {
	action := &microflows.RollbackObjectAction{
		BaseElement:      model.BaseElement{ID: model.ID(types.GenerateID())},
		RollbackVariable: s.Variable,
		RefreshInClient:  s.RefreshInClient,
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing

	return activity.ID
}

// addChangeObjectAction creates a CHANGE statement.
func (fb *flowBuilder) addChangeObjectAction(s *ast.ChangeObjectStmt) model.ID {
	action := &microflows.ChangeObjectAction{
		BaseElement:     model.BaseElement{ID: model.ID(types.GenerateID())},
		ChangeVariable:  s.Variable,
		Commit:          microflows.CommitTypeNo,
		RefreshInClient: false,
	}

	// Look up entity type from variable scope
	entityQN := ""
	if fb.varTypes != nil {
		entityQN = fb.varTypes[s.Variable]
	}

	// Build MemberChange items for each SET assignment
	for _, change := range s.Changes {
		memberChange := &microflows.MemberChange{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Type:        microflows.MemberChangeTypeSet,
			Value:       fb.memberExpressionToString(change.Value, entityQN, change.Attribute),
		}
		fb.resolveMemberChange(memberChange, change.Attribute, entityQN)
		action.Changes = append(action.Changes, memberChange)
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addRetrieveAction creates a RETRIEVE statement.
func (fb *flowBuilder) addRetrieveAction(s *ast.RetrieveStmt) model.ID {
	var source microflows.RetrieveSource

	if s.StartVariable != "" {
		// Association retrieve: RETRIEVE $List FROM $Parent/Module.AssocName
		assocQN := s.Source.Module + "." + s.Source.Name

		// Look up association to determine type and direction.
		// For Reference associations, AssociationRetrieveSource always returns a single
		// object (the entity on the other end). When the user navigates from the child
		// (non-owner) side, the intent is to get a list of parent entities — we must use
		// a DatabaseRetrieveSource with XPath constraint instead.
		assocInfo := fb.lookupAssociation(s.Source.Module, s.Source.Name)
		startVarType := ""
		if fb.varTypes != nil {
			startVarType = fb.varTypes[s.StartVariable]
		}

		if assocInfo != nil && assocInfo.Type == domainmodel.AssociationTypeReference &&
			assocInfo.childEntityQN != "" && startVarType == assocInfo.childEntityQN {
			// Reverse traversal on Reference: child → parent (one-to-many)
			// Use DatabaseRetrieveSource with XPath to get a list of parent entities
			dbSource := &microflows.DatabaseRetrieveSource{
				BaseElement:         model.BaseElement{ID: model.ID(types.GenerateID())},
				EntityQualifiedName: assocInfo.parentEntityQN,
				XPathConstraint:     "[" + assocQN + " = $" + s.StartVariable + "]",
			}
			source = dbSource
			if fb.varTypes != nil {
				fb.varTypes[s.Variable] = "List of " + assocInfo.parentEntityQN
			}
		} else {
			// Forward traversal or ReferenceSet: use AssociationRetrieveSource
			source = &microflows.AssociationRetrieveSource{
				BaseElement:              model.BaseElement{ID: model.ID(types.GenerateID())},
				StartVariable:            s.StartVariable,
				AssociationQualifiedName: assocQN,
			}
			if fb.varTypes != nil {
				if assocInfo != nil && assocInfo.Type == domainmodel.AssociationTypeReference {
					// Reference forward traversal: returns single object
					otherEntity := assocInfo.childEntityQN
					if startVarType == assocInfo.childEntityQN {
						otherEntity = assocInfo.parentEntityQN
					}
					fb.varTypes[s.Variable] = otherEntity
				} else {
					// ReferenceSet or unknown: returns a list
					fb.varTypes[s.Variable] = "List of " + assocQN
				}
			}
		}
	} else {
		// Database retrieve: RETRIEVE $List FROM Module.Entity WHERE ...
		entityQN := s.Source.Module + "." + s.Source.Name
		dbSource := &microflows.DatabaseRetrieveSource{
			BaseElement:         model.BaseElement{ID: model.ID(types.GenerateID())},
			EntityQualifiedName: entityQN,
		}

		// Set range if LIMIT is specified
		if s.Limit != "" {
			rangeType := microflows.RangeTypeCustom
			// LIMIT 1 with no offset uses RangeTypeFirst for single object retrieval
			if s.Limit == "1" && s.Offset == "" {
				rangeType = microflows.RangeTypeFirst
			}
			dbSource.Range = &microflows.Range{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				RangeType:   rangeType,
				Limit:       s.Limit,
				Offset:      s.Offset,
			}
		}

		// Convert WHERE expression if present
		// XPath constraints are stored with square brackets in BSON: [expression]
		if s.Where != nil {
			dbSource.XPathConstraint = "[" + expressionToXPath(s.Where) + "]"
		}

		// Convert SORT BY columns if present
		if len(s.SortColumns) > 0 {
			for _, col := range s.SortColumns {
				// Resolve attribute path - if just a simple name, prefix with entity
				attrPath := col.Attribute
				if !strings.Contains(attrPath, ".") {
					attrPath = entityQN + "." + attrPath
				}

				direction := microflows.SortDirectionAscending
				if strings.EqualFold(col.Order, "desc") {
					direction = microflows.SortDirectionDescending
				}

				dbSource.Sorting = append(dbSource.Sorting, &microflows.SortItem{
					BaseElement:            model.BaseElement{ID: model.ID(types.GenerateID())},
					AttributeQualifiedName: attrPath,
					Direction:              direction,
				})
			}
		}

		source = dbSource

		// Register variable type for CHANGE statements
		// RETRIEVE with LIMIT 1 returns a single entity, otherwise returns a List
		if fb.varTypes != nil {
			if s.Limit == "1" {
				// LIMIT 1 returns a single entity
				fb.varTypes[s.Variable] = entityQN
			} else {
				// No LIMIT or LIMIT > 1 returns a list
				fb.varTypes[s.Variable] = "List of " + entityQN
			}
		}
	}

	action := &microflows.RetrieveAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		OutputVariable: s.Variable,
		Source:         source,
	}

	activityX := fb.posX
	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
			ErrorHandlingType:   convertErrorHandlingType(s.ErrorHandling),
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing

	// Build custom error handler flow if present
	if s.ErrorHandling != nil && len(s.ErrorHandling.Body) > 0 {
		errorY := fb.posY + VerticalSpacing
		mergeID := fb.addErrorHandlerFlow(activity.ID, activityX, s.ErrorHandling.Body)
		fb.handleErrorHandlerMerge(mergeID, activity.ID, errorY)
	}

	return activity.ID
}

// addListOperationAction creates list operations like HEAD, TAIL, FIND, etc.
func (fb *flowBuilder) addListOperationAction(s *ast.ListOperationStmt) model.ID {
	var operation microflows.ListOperation

	switch s.Operation {
	case ast.ListOpHead:
		operation = &microflows.HeadOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
		}
	case ast.ListOpTail:
		operation = &microflows.TailOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
		}
	case ast.ListOpFind:
		operation = &microflows.FindOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
			Expression:   fb.exprToString(s.Condition),
		}
	case ast.ListOpFilter:
		operation = &microflows.FilterOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
			Expression:   fb.exprToString(s.Condition),
		}
	case ast.ListOpSort:
		// Resolve entity type from input variable for qualified attribute names
		entityType := ""
		if fb.varTypes != nil {
			listType := fb.varTypes[s.InputVariable]
			if after, ok := strings.CutPrefix(listType, "List of "); ok {
				entityType = after
			}
		}

		// Build sort items from SortSpecs
		var sortItems []*microflows.SortItem
		for _, spec := range s.SortSpecs {
			direction := microflows.SortDirectionAscending
			if !spec.Ascending {
				direction = microflows.SortDirectionDescending
			}
			// Build fully qualified attribute name: Entity.Attribute
			attrQN := spec.Attribute
			if entityType != "" && !strings.Contains(spec.Attribute, ".") {
				attrQN = entityType + "." + spec.Attribute
			}
			sortItems = append(sortItems, &microflows.SortItem{
				BaseElement:            model.BaseElement{ID: model.ID(types.GenerateID())},
				AttributeQualifiedName: attrQN,
				Direction:              direction,
			})
		}
		operation = &microflows.SortOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
			Sorting:      sortItems,
		}
	case ast.ListOpUnion:
		operation = &microflows.UnionOperation{
			BaseElement:   model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable1: s.InputVariable,
			ListVariable2: s.SecondVariable,
		}
	case ast.ListOpIntersect:
		operation = &microflows.IntersectOperation{
			BaseElement:   model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable1: s.InputVariable,
			ListVariable2: s.SecondVariable,
		}
	case ast.ListOpSubtract:
		operation = &microflows.SubtractOperation{
			BaseElement:   model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable1: s.InputVariable,
			ListVariable2: s.SecondVariable,
		}
	case ast.ListOpContains:
		operation = &microflows.ContainsOperation{
			BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable:   s.InputVariable,
			ObjectVariable: s.SecondVariable, // The item to check
		}
	case ast.ListOpEquals:
		operation = &microflows.EqualsOperation{
			BaseElement:   model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable1: s.InputVariable,
			ListVariable2: s.SecondVariable,
		}
	case ast.ListOpRange:
		rangeOp := &microflows.ListRangeOperation{
			BaseElement:  model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariable: s.InputVariable,
		}
		if s.OffsetExpr != nil {
			rangeOp.OffsetExpression = fb.exprToString(s.OffsetExpr)
		}
		if s.LimitExpr != nil {
			rangeOp.LimitExpression = fb.exprToString(s.LimitExpr)
		}
		operation = rangeOp
	default:
		return ""
	}

	action := &microflows.ListOperationAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		Operation:      operation,
		OutputVariable: s.OutputVariable,
	}

	// Track output variable type for operations that preserve/produce list types
	if fb.varTypes != nil && s.OutputVariable != "" && s.InputVariable != "" {
		inputType := fb.varTypes[s.InputVariable]
		switch s.Operation {
		case ast.ListOpFilter, ast.ListOpSort, ast.ListOpTail, ast.ListOpUnion, ast.ListOpIntersect, ast.ListOpSubtract, ast.ListOpRange:
			// These operations preserve the list type
			if inputType != "" {
				fb.varTypes[s.OutputVariable] = inputType
			}
		case ast.ListOpHead, ast.ListOpFind:
			// These return a single element (remove "List of " prefix)
			if after, ok := strings.CutPrefix(inputType, "List of "); ok {
				fb.varTypes[s.OutputVariable] = after
			}
			// CONTAINS and EQUALS return Boolean, no need to track
		}
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addAggregateListAction creates aggregate operations like COUNT, SUM, AVERAGE, etc.
func (fb *flowBuilder) addAggregateListAction(s *ast.AggregateListStmt) model.ID {
	var function microflows.AggregateFunction
	switch s.Operation {
	case ast.AggregateCount:
		function = microflows.AggregateFunctionCount
	case ast.AggregateSum:
		function = microflows.AggregateFunctionSum
	case ast.AggregateAverage:
		function = microflows.AggregateFunctionAverage
	case ast.AggregateMinimum:
		function = microflows.AggregateFunctionMin
	case ast.AggregateMaximum:
		function = microflows.AggregateFunctionMax
	case ast.AggregateReduce:
		function = microflows.AggregateFunctionReduce
	default:
		return ""
	}

	action := &microflows.AggregateListAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		InputVariable:  s.InputVariable,
		OutputVariable: s.OutputVariable,
		Function:       function,
	}

	if s.IsExpression && s.Expression != nil {
		action.UseExpression = true
		action.Expression = expressionToString(s.Expression)
	} else if s.Attribute != "" {
		// For SUM/AVG/MIN/MAX, build qualified attribute name from variable type
		if fb.varTypes != nil {
			listType := fb.varTypes[s.InputVariable]
			if after, ok := strings.CutPrefix(listType, "List of "); ok {
				entityType := after
				action.AttributeQualifiedName = entityType + "." + s.Attribute
			}
		}
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addCreateListAction creates a CREATE LIST OF statement.
func (fb *flowBuilder) addCreateListAction(s *ast.CreateListStmt) model.ID {
	entityQN := ""
	if s.EntityType.Module != "" && s.EntityType.Name != "" {
		entityQN = s.EntityType.Module + "." + s.EntityType.Name
	}

	action := &microflows.CreateListAction{
		BaseElement:         model.BaseElement{ID: model.ID(types.GenerateID())},
		OutputVariable:      s.Variable,
		EntityQualifiedName: entityQN,
	}

	// Register variable type as list
	if fb.varTypes != nil && entityQN != "" {
		fb.varTypes[s.Variable] = "List of " + entityQN
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addAddToListAction creates an ADD TO list statement.
func (fb *flowBuilder) addAddToListAction(s *ast.AddToListStmt) model.ID {
	action := &microflows.ChangeListAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		Type:           microflows.ChangeListTypeAdd,
		ChangeVariable: s.List,
		Value:          "$" + s.Item,
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// addRemoveFromListAction creates a REMOVE FROM list statement.
func (fb *flowBuilder) addRemoveFromListAction(s *ast.RemoveFromListStmt) model.ID {
	action := &microflows.ChangeListAction{
		BaseElement:    model.BaseElement{ID: model.ID(types.GenerateID())},
		Type:           microflows.ChangeListTypeRemove,
		ChangeVariable: s.List,
		Value:          "$" + s.Item,
	}

	activity := &microflows.ActionActivity{
		BaseActivity: microflows.BaseActivity{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: ActivityWidth, Height: ActivityHeight},
			},
			AutoGenerateCaption: true,
		},
		Action: action,
	}

	fb.objects = append(fb.objects, activity)
	fb.posX += fb.spacing
	return activity.ID
}

// isEntity checks whether a qualified name refers to an entity in the domain model.
func (fb *flowBuilder) isEntity(moduleName, entityName string) bool {
	if fb.backend == nil {
		return false
	}
	mod, err := fb.backend.GetModuleByName(moduleName)
	if err != nil || mod == nil {
		return false
	}
	dm, err := fb.backend.GetDomainModel(mod.ID)
	if err != nil || dm == nil {
		return false
	}
	for _, e := range dm.Entities {
		if e.Name == entityName {
			return true
		}
	}
	return false
}

type resolvedMemberKind int

const (
	resolvedMemberUnknown resolvedMemberKind = iota
	resolvedMemberAttribute
	resolvedMemberAssociation
)

type resolvedMember struct {
	kind          resolvedMemberKind
	qualifiedName string
}

type domainEntityRef struct {
	moduleName string
	dm         *domainmodel.DomainModel
	entity     *domainmodel.Entity
}

// resolveMemberChange determines whether a member name is an association or attribute
// and sets the appropriate field on the MemberChange.
//
// memberName can be either bare ("Order_Customer") or qualified ("MfTest.Order_Customer").
func (fb *flowBuilder) resolveMemberChange(mc *microflows.MemberChange, memberName string, entityQN string) {
	if entityQN == "" {
		// Entity type of $variable is unknown (e.g., the variable comes from a
		// java action whose return type isn't registered, or from the iterator
		// of an untyped loop). Without the entity we cannot query the domain
		// model — but we must NOT silently drop the member name, otherwise
		// `change $x (Module.Assoc = $y)` would round-trip as `change $x ( = $y)`
		// which is invalid MDL. Fall back to a shape heuristic:
		//
		//   * no dot                 -> bare attribute name
		//   * exactly one dot        -> `Module.Assoc` (association)
		//   * two or more dots       -> `Module.Entity.Attribute` (qualified attribute)
		//
		// Two-dot names are never associations in MDL (association names carry a
		// single qualifier — the module), so they must stay on AttributeQualified-
		// Name even when the entity type is unknown. This avoids miscategorising
		// something like `change $x (MyModule.MyEntity.Offset = 1)` as an
		// association change.
		resolveMemberChangeFallback(mc, memberName, "")
		return
	}

	switch strings.Count(memberName, ".") {
	case 0:
		resolved := fb.resolveBareMember(memberName, entityQN)
		if resolved.kind == resolvedMemberAssociation {
			mc.AssociationQualifiedName = resolved.qualifiedName
			if mc.AssociationQualifiedName == "" {
				moduleName := entityQN
				if dot := strings.Index(moduleName, "."); dot >= 0 {
					moduleName = moduleName[:dot]
				}
				mc.AssociationQualifiedName = moduleName + "." + memberName
			}
			return
		}
		if resolved.kind == resolvedMemberAttribute && resolved.qualifiedName != "" {
			mc.AttributeQualifiedName = resolved.qualifiedName
		} else {
			mc.AttributeQualifiedName = entityQN + "." + memberName
		}
		return
	case 1:
		mc.AssociationQualifiedName = memberName
		return
	default:
		mc.AttributeQualifiedName = memberName
		return
	}
}

func (fb *flowBuilder) resolveBareMember(memberName string, entityQN string) resolvedMember {
	if memberName == "" || entityQN == "" || fb.backend == nil {
		return resolvedMember{kind: resolvedMemberUnknown}
	}
	cacheKey := entityQN + "." + memberName
	if fb.memberResolutionCache != nil {
		if cached, ok := fb.memberResolutionCache[cacheKey]; ok {
			return cached
		}
	} else {
		fb.memberResolutionCache = make(map[string]resolvedMember)
	}

	result := fb.lookupBareMember(memberName, entityQN)
	fb.memberResolutionCache[cacheKey] = result
	return result
}

func (fb *flowBuilder) lookupBareMember(memberName string, entityQN string) resolvedMember {
	entityRef, ok := fb.lookupDomainEntity(entityQN)
	if !ok {
		return resolvedMember{kind: resolvedMemberUnknown}
	}

	if attrQN, ok := fb.lookupAttributeQualifiedName(memberName, entityRef, map[string]bool{}); ok {
		return resolvedMember{kind: resolvedMemberAttribute, qualifiedName: attrQN}
	}

	hierarchy := fb.collectEntityHierarchy(entityRef, map[string]bool{})
	entityIDs := make(map[model.ID]bool, len(hierarchy))
	for _, ref := range hierarchy {
		if ref.entity != nil && ref.entity.ID != "" {
			entityIDs[ref.entity.ID] = true
		}
	}
	for _, ref := range hierarchy {
		if ref.dm == nil {
			continue
		}
		for _, assoc := range ref.dm.Associations {
			if assoc.Name == memberName && (entityIDs[assoc.ParentID] || entityIDs[assoc.ChildID]) {
				return resolvedMember{
					kind:          resolvedMemberAssociation,
					qualifiedName: ref.moduleName + "." + assoc.Name,
				}
			}
		}
		for _, assoc := range ref.dm.CrossAssociations {
			if assoc.Name == memberName && entityIDs[assoc.ParentID] {
				return resolvedMember{
					kind:          resolvedMemberAssociation,
					qualifiedName: ref.moduleName + "." + assoc.Name,
				}
			}
		}
	}

	return resolvedMember{kind: resolvedMemberUnknown}
}

func (fb *flowBuilder) lookupDomainEntity(entityQN string) (domainEntityRef, bool) {
	parts := strings.SplitN(entityQN, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return domainEntityRef{}, false
	}
	moduleName, entityName := parts[0], parts[1]
	mod, err := fb.backend.GetModuleByName(moduleName)
	if err != nil || mod == nil {
		return domainEntityRef{}, false
	}
	dm, err := fb.backend.GetDomainModel(mod.ID)
	if err != nil || dm == nil {
		return domainEntityRef{}, false
	}

	entity := dm.FindEntityByName(entityName)
	if entity == nil {
		return domainEntityRef{}, false
	}
	return domainEntityRef{moduleName: moduleName, dm: dm, entity: entity}, true
}

func (fb *flowBuilder) lookupAttributeQualifiedName(memberName string, ref domainEntityRef, visited map[string]bool) (string, bool) {
	if ref.entity == nil {
		return "", false
	}
	key := ref.moduleName + "." + ref.entity.Name
	if visited[key] {
		return "", false
	}
	visited[key] = true

	for _, attr := range ref.entity.Attributes {
		if attr.Name == memberName {
			return key + "." + memberName, true
		}
	}

	if parentRef, ok := fb.lookupGeneralization(ref); ok {
		return fb.lookupAttributeQualifiedName(memberName, parentRef, visited)
	}
	return "", false
}

func (fb *flowBuilder) collectEntityHierarchy(ref domainEntityRef, visited map[string]bool) []domainEntityRef {
	if ref.entity == nil {
		return nil
	}
	key := ref.moduleName + "." + ref.entity.Name
	if visited[key] {
		return nil
	}
	visited[key] = true

	refs := []domainEntityRef{ref}
	if parentRef, ok := fb.lookupGeneralization(ref); ok {
		refs = append(refs, fb.collectEntityHierarchy(parentRef, visited)...)
	}
	return refs
}

func (fb *flowBuilder) lookupGeneralization(ref domainEntityRef) (domainEntityRef, bool) {
	if ref.entity == nil {
		return domainEntityRef{}, false
	}

	parentQN := ref.entity.GeneralizationRef
	if parentQN == "" && ref.entity.GeneralizationID != "" {
		if parent := findEntityByID(ref.dm, ref.entity.GeneralizationID); parent != nil {
			parentQN = ref.moduleName + "." + parent.Name
		}
	}
	if parentQN == "" {
		switch gen := ref.entity.Generalization.(type) {
		case domainmodel.GeneralizationBase:
			if parent := findEntityByID(ref.dm, gen.GeneralizationID); parent != nil {
				parentQN = ref.moduleName + "." + parent.Name
			}
		case *domainmodel.GeneralizationBase:
			if gen != nil {
				if parent := findEntityByID(ref.dm, gen.GeneralizationID); parent != nil {
					parentQN = ref.moduleName + "." + parent.Name
				}
			}
		}
	}
	if parentQN == "" {
		return domainEntityRef{}, false
	}
	return fb.lookupDomainEntity(parentQN)
}

func findEntityByID(dm *domainmodel.DomainModel, entityID model.ID) *domainmodel.Entity {
	if dm == nil || entityID == "" {
		return nil
	}
	for _, entity := range dm.Entities {
		if entity.ID == entityID {
			return entity
		}
	}
	return nil
}

// resolveMemberChangeFallback preserves the authored member name shape when the
// entity metadata is unavailable.
//
//   - 0 dots  => bare attribute name. If entityQN is known, qualify it as
//     `Module.Entity.Attribute`; otherwise preserve the bare attribute.
//   - 1 dot   => association qualified by module (`Module.Association`).
//   - >=2 dots => fully qualified attribute (`Module.Entity.Attribute`).
func resolveMemberChangeFallback(mc *microflows.MemberChange, memberName string, entityQN string) {
	if memberName == "" {
		return
	}
	switch strings.Count(memberName, ".") {
	case 0:
		if entityQN == "" {
			mc.AttributeQualifiedName = memberName
		} else {
			mc.AttributeQualifiedName = entityQN + "." + memberName
		}
	case 1:
		mc.AssociationQualifiedName = memberName
	default:
		mc.AttributeQualifiedName = memberName
	}
}

// assocLookupResult holds resolved association metadata.
type assocLookupResult struct {
	Type           domainmodel.AssociationType
	parentEntityQN string // Qualified name of the parent (FROM/owner) entity
	childEntityQN  string // Qualified name of the child (TO/referenced) entity
}

// lookupAssociation finds an association by module and name, returning its type
// and the qualified names of its parent and child entities. Returns nil if the
// association cannot be found (e.g., backend is nil or module doesn't exist).
func (fb *flowBuilder) lookupAssociation(moduleName, assocName string) *assocLookupResult {
	if fb.backend == nil {
		return nil
	}
	cacheKey := moduleName + "." + assocName
	if fb.assocLookupCache != nil {
		if cached, ok := fb.assocLookupCache[cacheKey]; ok {
			return cached
		}
	} else {
		fb.assocLookupCache = make(map[string]*assocLookupResult)
	}

	mod, err := fb.backend.GetModuleByName(moduleName)
	if err != nil || mod == nil {
		fb.assocLookupCache[cacheKey] = nil
		return nil
	}
	dm, err := fb.backend.GetDomainModel(mod.ID)
	if err != nil || dm == nil {
		fb.assocLookupCache[cacheKey] = nil
		return nil
	}

	// Build entity ID → qualified name map
	entityNames := make(map[model.ID]string, len(dm.Entities))
	for _, e := range dm.Entities {
		entityNames[e.ID] = moduleName + "." + e.Name
	}

	for _, a := range dm.Associations {
		if a.Name == assocName {
			result := &assocLookupResult{
				Type:           a.Type,
				parentEntityQN: entityNames[a.ParentID],
				childEntityQN:  entityNames[a.ChildID],
			}
			fb.assocLookupCache[cacheKey] = result
			return result
		}
	}
	fb.assocLookupCache[cacheKey] = nil
	return nil
}
