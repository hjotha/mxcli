// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow flow graph: IF/ELSE and LOOP control flow builders
package executor

import (
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// addIfStatement creates an IF/THEN/ELSE statement using ExclusiveSplit and ExclusiveMerge.
// Layout strategy:
// - IF with ELSE: TRUE path goes horizontal (happy path), FALSE path goes below
// - IF without ELSE: TRUE path goes below, FALSE path goes horizontal (happy path)
// When a branch ends with RETURN, it terminates at its own EndEvent and does not
// connect to the merge. When both branches end with RETURN, no merge is created.
func (fb *flowBuilder) addIfStatement(s *ast.IfStmt) model.ID {
	if len(s.ThenBody) == 0 && len(s.ElseBody) == 0 {
		fb.pendingAnnotations = nil
		return ""
	}

	// First, measure the branches to know how much space they need
	thenBounds := fb.measurer.measureStatements(s.ThenBody)
	elseBounds := fb.measurer.measureStatements(s.ElseBody)

	// Calculate branch width (max of both branches)
	branchWidth := max(thenBounds.Width, elseBounds.Width)
	if branchWidth == 0 {
		branchWidth = HorizontalSpacing / 2
	}

	// Check if branches end with RETURN (creating their own EndEvents)
	thenReturns := lastStmtIsReturn(s.ThenBody)
	hasElseBody := s.HasElse || len(s.ElseBody) > 0
	elseReturns := hasElseBody && lastStmtIsReturn(s.ElseBody)
	bothReturn := hasElseBody && thenReturns && elseReturns
	thenNeedsErrorMerge := thenReturns && bodyHasEmptyCustomErrorHandler(s.ThenBody)
	elseNeedsErrorMerge := elseReturns && bodyHasEmptyCustomErrorHandler(s.ElseBody)

	// Save/restore endsWithReturn around branch processing to avoid
	// a branch's RETURN affecting the parent flow state prematurely
	savedEndsWithReturn := fb.endsWithReturn

	// Save current center line position
	splitX := fb.posX
	centerY := fb.posY // This is the center line for the happy path

	// Decide whether the IF condition is a rule call or a plain expression.
	// A rule-based split must be serialized as Microflows$RuleSplitCondition;
	// emitting ExpressionSplitCondition for a rule call causes Studio Pro to
	// raise CE0117 "Error(s) in expression".
	caption := fb.exprToString(s.Condition)
	splitCondition := fb.buildSplitCondition(s.Condition, caption)

	split := &microflows.ExclusiveSplit{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: splitX, Y: centerY},
			Size:        model.Size{Width: SplitWidth, Height: SplitHeight},
		},
		Caption:           caption,
		SplitCondition:    splitCondition,
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
	}
	fb.objects = append(fb.objects, split)
	splitID := split.ID

	// Apply this IF's own annotations (e.g. @caption, @annotation) to the split
	// BEFORE recursing into branch bodies. Otherwise a nested statement's annotations
	// would overwrite fb.pendingAnnotations (shared state) and the outer loop in
	// buildFlowGraph would then attach the wrong caption/annotation to this split.
	//
	// Capture per-branch anchor overrides before pendingAnnotations is cleared so
	// the TRUE/FALSE flows emitted below can pick them up.
	var trueBranchAnchor, falseBranchAnchor *ast.FlowAnchors
	if fb.pendingAnnotations != nil {
		trueBranchAnchor = fb.pendingAnnotations.TrueBranchAnchor
		falseBranchAnchor = fb.pendingAnnotations.FalseBranchAnchor
		fb.applyAnnotations(splitID, fb.pendingAnnotations)
		fb.pendingAnnotations = nil
	}

	// Calculate merge position (after the longest branch)
	mergeX := splitX + SplitWidth + HorizontalSpacing/2 + branchWidth + HorizontalSpacing/2

	// Determine if the merge would have 2+ incoming edges (non-redundant).
	// Skip merge when only one branch flows into it (the other returns).
	needMerge := false
	if !bothReturn {
		if hasElseBody {
			needMerge = (!thenReturns && !elseReturns) || thenNeedsErrorMerge || elseNeedsErrorMerge
		} else {
			needMerge = !thenReturns // THEN continues + FALSE path → 2 inputs
		}
	}

	var mergeID model.ID
	if needMerge {
		merge := &microflows.ExclusiveMerge{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: mergeX, Y: centerY},
				Size:        model.Size{Width: MergeSize, Height: MergeSize},
			},
		}
		fb.objects = append(fb.objects, merge)
		mergeID = merge.ID
	}

	thenStartX := splitX + SplitWidth + HorizontalSpacing/2
	var noMergeExitID model.ID
	var noMergeExitCase string
	var noMergeExitAnchor *ast.FlowAnchors
	routePendingErrorToElse := hasElseBody && fb.errorHandlerSkipVar != "" && exprReferencesVar(s.Condition, fb.errorHandlerSkipVar)
	pendingErrorForElse := pendingErrorHandlerState{}

	if hasElseBody {
		// IF WITH ELSE: TRUE path horizontal (happy path), FALSE path below
		if routePendingErrorToElse {
			pendingErrorForElse = fb.capturePendingErrorHandler()
			fb.clearPendingErrorHandler()
		}
		fb.posX = thenStartX
		fb.posY = centerY
		fb.endsWithReturn = false

		var lastThenID model.ID
		var prevThenAnchor *ast.FlowAnchors
		var pendingThenCase string
		var pendingThenAnchor *ast.FlowAnchors
		for i, stmt := range s.ThenBody {
			thisAnchor := stmtOwnAnchor(stmt)
			actID := fb.addStatement(stmt)
			if actID != "" {
				fb.applyPendingAnnotations(actID)
				if lastThenID == "" {
					// First statement in THEN - connect from split with "true" case.
					// Origin: trueBranchAnchor.From (if set) — anchor on the split side.
					// Destination: prefer the first statement's own @anchor(to: ...) if it
					// has one; otherwise fall back to trueBranchAnchor.To.
					flow := newHorizontalFlowWithCase(splitID, actID, "true")
					applyUserAnchors(flow, trueBranchAnchor, branchDestinationAnchor(trueBranchAnchor, thisAnchor))
					fb.flows = append(fb.flows, flow)
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ThenBody[i+1:], fb.errorHandlerSkipVar))
				} else {
					var flow *microflows.SequenceFlow
					originAnchor := prevThenAnchor
					destAnchor := thisAnchor
					if pendingThenCase != "" {
						flow = newHorizontalFlowWithCase(lastThenID, actID, pendingThenCase)
						originAnchor, destAnchor = pendingFlowAnchors(prevThenAnchor, pendingThenAnchor, thisAnchor)
						pendingThenCase = ""
						pendingThenAnchor = nil
					} else {
						flow = newHorizontalFlow(lastThenID, actID)
					}
					applyUserAnchors(flow, originAnchor, destAnchor)
					fb.flows = append(fb.flows, flow)
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ThenBody[i+1:], fb.errorHandlerSkipVar))
				}
				prevThenAnchor = thisAnchor
				// For nested compound statements, use their exit point
				if fb.nextConnectionPoint != "" {
					lastThenID = fb.nextConnectionPoint
					fb.nextConnectionPoint = ""
					pendingThenCase = fb.nextFlowCase
					fb.nextFlowCase = ""
					pendingThenAnchor = fb.nextFlowAnchor
					fb.nextFlowAnchor = nil
				} else {
					lastThenID = actID
				}
			}
		}

		// Connect THEN body to merge only if it doesn't end with RETURN and a merge exists.
		// When needMerge=false, the continuing branch is wired up by the parent via
		// nextConnectionPoint/nextFlowCase, so we must not emit a dangling flow here.
		if !thenReturns && needMerge {
			if lastThenID != "" {
				var flow *microflows.SequenceFlow
				originAnchor := prevThenAnchor
				destAnchor := prevThenAnchor
				if pendingThenCase != "" {
					flow = newHorizontalFlowWithCase(lastThenID, mergeID, pendingThenCase)
					originAnchor, destAnchor = pendingFlowAnchors(prevThenAnchor, pendingThenAnchor, prevThenAnchor)
				} else {
					flow = newHorizontalFlow(lastThenID, mergeID)
				}
				applyUserAnchors(flow, originAnchor, destAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			} else {
				// Empty THEN body - connect split directly to merge with true case.
				// Pass trueBranchAnchor as destination too so the @anchor(true: (..., to: Y))
				// from the describer round-trips into the merge side of the flow.
				flow := newHorizontalFlowWithCase(splitID, mergeID, "true")
				applyUserAnchors(flow, trueBranchAnchor, trueBranchAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			}
		} else if thenReturns && needMerge {
			fb.addPendingErrorHandlerFlowTo(mergeID)
		}

		// Process ELSE body (below the THEN path)
		if routePendingErrorToElse && fb.capturePendingErrorHandler().isEmpty() {
			fb.restorePendingErrorHandler(pendingErrorForElse)
		}
		elseCenterY := centerY + VerticalSpacing
		fb.posX = thenStartX
		fb.posY = elseCenterY
		fb.endsWithReturn = false

		var lastElseID model.ID
		var prevElseAnchor *ast.FlowAnchors
		var pendingElseCase string
		var pendingElseAnchor *ast.FlowAnchors
		for i, stmt := range s.ElseBody {
			thisAnchor := stmtOwnAnchor(stmt)
			actID := fb.addStatement(stmt)
			if actID != "" {
				fb.applyPendingAnnotations(actID)
				if lastElseID == "" {
					// First statement in ELSE - connect from split going down (false path).
					// Same compositional rule as the THEN branch.
					flow := newDownwardFlowWithCase(splitID, actID, "false")
					applyUserAnchors(flow, falseBranchAnchor, branchDestinationAnchor(falseBranchAnchor, thisAnchor))
					fb.flows = append(fb.flows, flow)
					if routePendingErrorToElse {
						fb.routePendingErrorHandlerToAlternative(splitID, actID)
					}
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ElseBody[i+1:], fb.errorHandlerSkipVar))
				} else {
					var flow *microflows.SequenceFlow
					originAnchor := prevElseAnchor
					destAnchor := thisAnchor
					if pendingElseCase != "" {
						flow = newHorizontalFlowWithCase(lastElseID, actID, pendingElseCase)
						originAnchor, destAnchor = pendingFlowAnchors(prevElseAnchor, pendingElseAnchor, thisAnchor)
						pendingElseCase = ""
						pendingElseAnchor = nil
					} else {
						flow = newHorizontalFlow(lastElseID, actID)
					}
					applyUserAnchors(flow, originAnchor, destAnchor)
					fb.flows = append(fb.flows, flow)
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ElseBody[i+1:], fb.errorHandlerSkipVar))
				}
				prevElseAnchor = thisAnchor
				// For nested compound statements, use their exit point
				if fb.nextConnectionPoint != "" {
					lastElseID = fb.nextConnectionPoint
					fb.nextConnectionPoint = ""
					pendingElseCase = fb.nextFlowCase
					fb.nextFlowCase = ""
					pendingElseAnchor = fb.nextFlowAnchor
					fb.nextFlowAnchor = nil
				} else {
					lastElseID = actID
				}
			}
		}

		// Connect ELSE body to merge only if it doesn't end with RETURN and a merge exists.
		// When needMerge=false, the continuing branch is handled by the parent; emitting
		// a flow with an empty mergeID would create an orphan SequenceFlow.
		if !elseReturns && needMerge {
			if lastElseID != "" {
				flow := newUpwardFlow(lastElseID, mergeID)
				originAnchor := prevElseAnchor
				destAnchor := prevElseAnchor
				if pendingElseCase != "" {
					flow.CaseValue = microflows.EnumerationCase{
						BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
						Value:       pendingElseCase,
					}
					originAnchor, destAnchor = pendingFlowAnchors(prevElseAnchor, pendingElseAnchor, prevElseAnchor)
				}
				applyUserAnchors(flow, originAnchor, destAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			} else {
				flow := newDownwardFlowWithCase(splitID, mergeID, "false")
				applyUserAnchors(flow, falseBranchAnchor, falseBranchAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			}
		} else if elseReturns && needMerge {
			fb.addPendingErrorHandlerFlowTo(mergeID)
		}
		if !needMerge {
			if thenReturns && !elseReturns {
				if lastElseID != "" {
					noMergeExitID = lastElseID
					noMergeExitCase = pendingElseCase
					if pendingElseAnchor != nil {
						noMergeExitAnchor = pendingElseAnchor
					} else {
						noMergeExitAnchor = prevElseAnchor
					}
				} else {
					noMergeExitID = splitID
					noMergeExitCase = "false"
					noMergeExitAnchor = falseBranchAnchor
				}
			} else if elseReturns && !thenReturns {
				if lastThenID != "" {
					noMergeExitID = lastThenID
					noMergeExitCase = pendingThenCase
					if pendingThenAnchor != nil {
						noMergeExitAnchor = pendingThenAnchor
					} else {
						noMergeExitAnchor = prevThenAnchor
					}
				} else {
					noMergeExitID = splitID
					noMergeExitCase = "true"
					noMergeExitAnchor = trueBranchAnchor
				}
			}
		}
	} else {
		// IF WITHOUT ELSE: FALSE path horizontal (happy path), TRUE path below
		// This keeps the "do nothing" path straight and the "do something" path below

		if needMerge {
			// FALSE path: connect split directly to merge horizontally.
			// Pass falseBranchAnchor as destination too so @anchor(false: (..., to: Y))
			// round-trips through the merge side of the split-to-merge flow.
			flow := newHorizontalFlowWithCase(splitID, mergeID, "false")
			applyUserAnchors(flow, falseBranchAnchor, falseBranchAnchor)
			fb.flows = append(fb.flows, flow)
			fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
		}
		// When !needMerge (thenReturns): FALSE flow is deferred — the parent will
		// connect splitID to the next activity with nextFlowCase="false".

		// TRUE path: goes below the main line
		thenCenterY := centerY + VerticalSpacing
		fb.posX = thenStartX
		fb.posY = thenCenterY
		fb.endsWithReturn = false

		var lastThenID model.ID
		var prevThenAnchor *ast.FlowAnchors
		var pendingThenCase string
		var pendingThenAnchor *ast.FlowAnchors
		for i, stmt := range s.ThenBody {
			thisAnchor := stmtOwnAnchor(stmt)
			actID := fb.addStatement(stmt)
			if actID != "" {
				fb.applyPendingAnnotations(actID)
				if lastThenID == "" {
					// First statement in THEN - connect from split going down with "true" case
					flow := newDownwardFlowWithCase(splitID, actID, "true")
					applyUserAnchors(flow, trueBranchAnchor, branchDestinationAnchor(trueBranchAnchor, thisAnchor))
					fb.flows = append(fb.flows, flow)
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ThenBody[i+1:], fb.errorHandlerSkipVar))
				} else {
					var flow *microflows.SequenceFlow
					originAnchor := prevThenAnchor
					destAnchor := thisAnchor
					if pendingThenCase != "" {
						flow = newHorizontalFlowWithCase(lastThenID, actID, pendingThenCase)
						originAnchor, destAnchor = pendingFlowAnchors(prevThenAnchor, pendingThenAnchor, thisAnchor)
						pendingThenCase = ""
						pendingThenAnchor = nil
					} else {
						flow = newHorizontalFlow(lastThenID, actID)
					}
					applyUserAnchors(flow, originAnchor, destAnchor)
					fb.flows = append(fb.flows, flow)
					fb.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.ThenBody[i+1:], fb.errorHandlerSkipVar))
				}
				prevThenAnchor = thisAnchor
				// For nested compound statements, use their exit point
				if fb.nextConnectionPoint != "" {
					lastThenID = fb.nextConnectionPoint
					fb.nextConnectionPoint = ""
					pendingThenCase = fb.nextFlowCase
					fb.nextFlowCase = ""
					pendingThenAnchor = fb.nextFlowAnchor
					fb.nextFlowAnchor = nil
				} else {
					lastThenID = actID
				}
			}
		}

		// Connect THEN body to merge only if it doesn't end with RETURN and a merge exists.
		// With no ELSE + thenReturns, needMerge=false and the FALSE path is threaded through
		// the parent — any flow emitted here would dangle with mergeID="".
		if !thenReturns && needMerge {
			if lastThenID != "" {
				flow := newUpwardFlow(lastThenID, mergeID)
				originAnchor := prevThenAnchor
				destAnchor := prevThenAnchor
				if pendingThenCase != "" {
					flow.CaseValue = microflows.EnumerationCase{
						BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
						Value:       pendingThenCase,
					}
					originAnchor, destAnchor = pendingFlowAnchors(prevThenAnchor, pendingThenAnchor, prevThenAnchor)
				}
				applyUserAnchors(flow, originAnchor, destAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			} else {
				// Empty THEN body - connect split directly to merge going down and back up.
				// Pass trueBranchAnchor as destination too so @anchor(true: (..., to: Y))
				// round-trips into the merge side of the flow.
				flow := newDownwardFlowWithCase(splitID, mergeID, "true")
				applyUserAnchors(flow, trueBranchAnchor, trueBranchAnchor)
				fb.flows = append(fb.flows, flow)
				fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
			}
		}
		if !needMerge {
			noMergeExitID = splitID
			noMergeExitCase = "false"
			noMergeExitAnchor = falseBranchAnchor
		}
	}

	// If both branches end with RETURN, the flow terminates here
	if bothReturn {
		fb.endsWithReturn = true
		return splitID
	}

	// Restore endsWithReturn - a single branch returning doesn't end the overall flow
	fb.endsWithReturn = savedEndsWithReturn

	if needMerge {
		// Update position to after the merge, on the happy path center line
		fb.posX = mergeX + MergeSize + HorizontalSpacing/2
		fb.posY = centerY
		fb.nextConnectionPoint = mergeID
	} else {
		// No merge: the split's continuing branch connects directly to the next activity.
		// Position after the split, past the downward branch's horizontal extent.
		afterSplit := splitX + SplitWidth + HorizontalSpacing
		afterBranch := thenStartX + thenBounds.Width + HorizontalSpacing/2
		if !hasElseBody {
			fb.posX = max(afterSplit, afterBranch)
		} else {
			fb.posX = max(afterSplit, afterBranch)
		}
		fb.posY = centerY
		if noMergeExitID != "" {
			fb.nextConnectionPoint = noMergeExitID
			fb.nextFlowCase = noMergeExitCase
			fb.nextFlowAnchor = noMergeExitAnchor
		} else {
			fb.nextConnectionPoint = splitID
		}
	}

	return splitID
}

func bodyHasEmptyCustomErrorHandler(stmts []ast.MicroflowStatement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.CallMicroflowStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.CallJavaActionStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.RestCallStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.ImportFromMappingStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.CreateObjectStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.MfCommitStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.DeleteObjectStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.CallWebServiceStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.CallExternalActionStmt:
			if isEmptyCustomErrorHandler(s.ErrorHandling) || bodyHasEmptyCustomErrorHandler(errorBody(s.ErrorHandling)) {
				return true
			}
		case *ast.IfStmt:
			if bodyHasEmptyCustomErrorHandler(s.ThenBody) || bodyHasEmptyCustomErrorHandler(s.ElseBody) {
				return true
			}
		case *ast.LoopStmt:
			if bodyHasEmptyCustomErrorHandler(s.Body) {
				return true
			}
		case *ast.WhileStmt:
			if bodyHasEmptyCustomErrorHandler(s.Body) {
				return true
			}
		}
	}
	return false
}

func errorBody(eh *ast.ErrorHandlingClause) []ast.MicroflowStatement {
	if eh == nil {
		return nil
	}
	return eh.Body
}

// addLoopStatement creates a LOOP statement using LoopedActivity.
// Layout: Auto-sizes the loop box to fit content with padding
func (fb *flowBuilder) addLoopStatement(s *ast.LoopStmt) model.ID {
	// Snapshot & clear this loop's own annotations so the body's recursive
	// addStatement calls can't consume them. We re-apply to the loop activity
	// after it's created below.
	savedLoopAnnotations := fb.pendingAnnotations
	fb.pendingAnnotations = nil

	// First, measure the loop body to determine size
	bodyBounds := fb.measurer.measureStatements(s.Body)

	// Calculate loop box size with padding
	// Extra width for iterator icon and its label (100 pixels)
	iteratorSpace := 100
	loopWidth := max(bodyBounds.Width+2*LoopPadding+iteratorSpace, MinLoopWidth)
	loopHeight := max(bodyBounds.Height+2*LoopPadding, MinLoopHeight)

	// Inner positioning: activities need to be offset from the iterator icon
	// The iterator takes up space in the top-left, so we need extra X offset
	// Looking at working Mendix loops, inner content starts further right
	innerStartX := LoopPadding + iteratorSpace    // Extra offset for iterator icon and label
	innerStartY := LoopPadding + ActivityHeight/2 // Center activities vertically with some padding

	loopLeftX := fb.posX
	loopCenterX := loopLeftX + loopWidth/2
	if s.Annotations != nil && s.Annotations.Position != nil {
		loopCenterX = s.Annotations.Position.X
		loopLeftX = loopCenterX - loopWidth/2
	}

	// Add loop variable to varTypes with element type derived from list type
	// If $ProductList is "List of MfTest.Product", then $Product is "MfTest.Product"
	if fb.varTypes != nil {
		listType := fb.varTypes[s.ListVariable]
		if after, ok := strings.CutPrefix(listType, "List of "); ok {
			elementType := after
			fb.varTypes[s.LoopVariable] = elementType
		}
	}

	// Build nested ObjectCollection for loop body
	loopBuilder := &flowBuilder{
		posX:                   innerStartX,
		posY:                   innerStartY,
		baseY:                  innerStartY,
		spacing:                HorizontalSpacing,
		returnType:             fb.returnType,
		hasReturnValue:         fb.hasReturnValue,
		varTypes:               fb.varTypes,     // Share variable scope
		declaredVars:           fb.declaredVars, // Share declared vars (fixes nil map panic)
		measurer:               fb.measurer,     // Share measurer
		backend:                fb.backend,      // Share backend
		hierarchy:              fb.hierarchy,    // Share hierarchy
		restServices:           fb.restServices, // Share REST services for parameter classification
		callOutputDeclarations: fb.callOutputDeclarations,
	}

	// Process loop body statements and connect them with flows.
	var lastBodyID model.ID
	var pendingBodyCase string
	var pendingBodyAnchor *ast.FlowAnchors
	var prevBodyAnchor *ast.FlowAnchors
	for i, stmt := range s.Body {
		thisAnchor := stmtOwnAnchor(stmt)
		actID := loopBuilder.addStatement(stmt)
		if actID != "" {
			loopBuilder.applyPendingAnnotations(actID)
			if lastBodyID != "" {
				var flow *microflows.SequenceFlow
				originAnchor := prevBodyAnchor
				destAnchor := thisAnchor
				if pendingBodyCase != "" {
					flow = newHorizontalFlowWithCase(lastBodyID, actID, pendingBodyCase)
					originAnchor, destAnchor = pendingFlowAnchors(prevBodyAnchor, pendingBodyAnchor, thisAnchor)
					pendingBodyCase = ""
					pendingBodyAnchor = nil
				} else {
					flow = newHorizontalFlow(lastBodyID, actID)
				}
				applyUserAnchors(flow, originAnchor, destAnchor)
				loopBuilder.flows = append(loopBuilder.flows, flow)
				loopBuilder.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.Body[i+1:], loopBuilder.errorHandlerSkipVar))
			}
			prevBodyAnchor = thisAnchor
			// Handle nextConnectionPoint for compound statements (nested IF, etc.)
			if loopBuilder.nextConnectionPoint != "" {
				lastBodyID = loopBuilder.nextConnectionPoint
				loopBuilder.nextConnectionPoint = ""
				pendingBodyCase = loopBuilder.nextFlowCase
				loopBuilder.nextFlowCase = ""
				pendingBodyAnchor = loopBuilder.nextFlowAnchor
				loopBuilder.nextFlowAnchor = nil
			} else {
				lastBodyID = actID
			}
		}
	}
	if pendingBodyCase != "" && lastBodyID != "" {
		merge := &microflows.ExclusiveMerge{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: loopBuilder.posX, Y: loopBuilder.posY},
				Size:        model.Size{Width: MergeSize, Height: MergeSize},
			},
		}
		loopBuilder.objects = append(loopBuilder.objects, merge)
		flow := newHorizontalFlowWithCase(lastBodyID, merge.ID, pendingBodyCase)
		applyUserAnchors(flow, pendingBodyAnchor, pendingBodyAnchor)
		loopBuilder.flows = append(loopBuilder.flows, flow)
		loopBuilder.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
	}

	// Create LoopedActivity with calculated size
	// Position is the CENTER point (RelativeMiddlePoint in Mendix)
	loop := &microflows.LoopedActivity{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: loopCenterX, Y: fb.posY},
			Size:        model.Size{Width: loopWidth, Height: loopHeight},
		},
		LoopSource: &microflows.IterableList{
			BaseElement:      model.BaseElement{ID: model.ID(types.GenerateID())},
			ListVariableName: s.ListVariable,
			VariableName:     s.LoopVariable,
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Objects:     loopBuilder.objects,
			Flows:       nil, // Internal flows go at top-level, not inside the loop's ObjectCollection
		},
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
	}

	// @anchor(iterator: ..., tail: ...) parses and survives on
	// savedLoopAnnotations for forward compatibility, but we deliberately do
	// not serialise either edge as a SequenceFlow: Studio Pro rejects loop→body
	// and body→loop with CE0709 "Sequence flow is not accepted by origin or
	// destination", since the iterator icon is drawn implicitly by the loop
	// geometry.

	fb.objects = append(fb.objects, loop)

	// Add the internal flows to the parent's flows (top-level), not inside loop
	// This is how Mendix stores them - all flows at the microflow level
	fb.flows = append(fb.flows, loopBuilder.flows...)

	// Re-apply this loop's own annotations now that its activity exists.
	if savedLoopAnnotations != nil {
		fb.applyAnnotations(loop.ID, savedLoopAnnotations)
	}

	fb.posX = loopLeftX + loopWidth + HorizontalSpacing

	return loop.ID
}

func isManualWhileTrueCandidate(s *ast.WhileStmt) bool {
	if s == nil || containsBreakStmt(s.Body) || (!containsContinueStmt(s.Body) && !containsTerminalStmt(s.Body)) {
		return false
	}
	lit, ok := s.Condition.(*ast.LiteralExpr)
	if !ok || lit.Kind != ast.LiteralBoolean {
		return false
	}
	value, ok := lit.Value.(bool)
	return ok && value
}

func containsContinueStmt(stmts []ast.MicroflowStatement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ContinueStmt:
			return true
		case *ast.IfStmt:
			if containsContinueStmt(s.ThenBody) || containsContinueStmt(s.ElseBody) {
				return true
			}
		case *ast.LoopStmt:
			if containsContinueStmt(s.Body) {
				return true
			}
		case *ast.WhileStmt:
			if containsContinueStmt(s.Body) {
				return true
			}
		case *ast.InheritanceSplitStmt:
			if containsContinueStmt(s.ElseBody) {
				return true
			}
			for _, c := range s.Cases {
				if containsContinueStmt(c.Body) {
					return true
				}
			}
		}
	}
	return false
}

func (fb *flowBuilder) addManualWhileTrueStatement(s *ast.WhileStmt) model.ID {
	mergeX := fb.posX
	mergeY := fb.posY
	if s.Annotations != nil && s.Annotations.Position != nil {
		mergeX = s.Annotations.Position.X
		mergeY = s.Annotations.Position.Y
	}

	merge := &microflows.ExclusiveMerge{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: mergeX, Y: mergeY},
			Size:        model.Size{Width: MergeSize, Height: MergeSize},
		},
	}
	fb.objects = append(fb.objects, merge)
	if fb.pendingAnnotations != nil {
		fb.applyAnnotations(merge.ID, fb.pendingAnnotations)
		fb.pendingAnnotations = nil
	}

	savedBackTarget := fb.manualLoopBackTarget
	fb.manualLoopBackTarget = merge.ID
	defer func() { fb.manualLoopBackTarget = savedBackTarget }()

	fb.posX = mergeX + MergeSize + HorizontalSpacing/2
	fb.posY = mergeY

	lastBodyID := merge.ID
	var pendingBodyCase string
	var pendingBodyAnchor *ast.FlowAnchors
	var prevBodyAnchor *ast.FlowAnchors
	for _, stmt := range s.Body {
		thisAnchor := stmtOwnAnchor(stmt)
		actID := fb.addStatement(stmt)
		if actID == "" {
			continue
		}
		fb.applyPendingAnnotations(actID)
		if lastBodyID != "" && lastBodyID != actID {
			var flow *microflows.SequenceFlow
			originAnchor := prevBodyAnchor
			destAnchor := thisAnchor
			if pendingBodyCase != "" {
				flow = newHorizontalFlowWithCase(lastBodyID, actID, pendingBodyCase)
				originAnchor, destAnchor = pendingFlowAnchors(prevBodyAnchor, pendingBodyAnchor, thisAnchor)
				pendingBodyCase = ""
				pendingBodyAnchor = nil
			} else {
				flow = newHorizontalFlow(lastBodyID, actID)
			}
			applyUserAnchors(flow, originAnchor, destAnchor)
			fb.flows = append(fb.flows, flow)
			fb.addPendingEmptyErrorHandlerFlow(flow.OriginID, flow.DestinationID)
		}
		prevBodyAnchor = thisAnchor
		if fb.nextConnectionPoint != "" {
			lastBodyID = fb.nextConnectionPoint
			fb.nextConnectionPoint = ""
			pendingBodyCase = fb.nextFlowCase
			fb.nextFlowCase = ""
			pendingBodyAnchor = fb.nextFlowAnchor
			fb.nextFlowAnchor = nil
		} else {
			lastBodyID = actID
		}
	}

	if lastBodyID != "" && lastBodyID != merge.ID && !lastStmtIsReturn(s.Body) {
		var flow *microflows.SequenceFlow
		if pendingBodyCase != "" {
			flow = newHorizontalFlowWithCase(lastBodyID, merge.ID, pendingBodyCase)
		} else {
			flow = newHorizontalFlow(lastBodyID, merge.ID)
		}
		applyUserAnchors(flow, pendingBodyAnchor, pendingBodyAnchor)
		fb.flows = append(fb.flows, flow)
	}
	fb.endsWithReturn = true

	return merge.ID
}

// addWhileStatement creates a WHILE loop using LoopedActivity with WhileLoopCondition.
// Layout matches addLoopStatement but without iterator icon space.
func (fb *flowBuilder) addWhileStatement(s *ast.WhileStmt) model.ID {
	if isManualWhileTrueCandidate(s) {
		return fb.addManualWhileTrueStatement(s)
	}

	// Snapshot & clear this WHILE's own annotations so the body's recursive
	// addStatement calls can't consume them (see addLoopStatement).
	savedWhileAnnotations := fb.pendingAnnotations
	fb.pendingAnnotations = nil

	bodyBounds := fb.measurer.measureStatements(s.Body)

	loopWidth := max(bodyBounds.Width+2*LoopPadding, MinLoopWidth)
	loopHeight := max(bodyBounds.Height+2*LoopPadding, MinLoopHeight)

	innerStartX := LoopPadding
	innerStartY := LoopPadding + ActivityHeight/2

	loopLeftX := fb.posX
	loopCenterX := loopLeftX + loopWidth/2
	if s.Annotations != nil && s.Annotations.Position != nil {
		loopCenterX = s.Annotations.Position.X
		loopLeftX = loopCenterX - loopWidth/2
	}

	loopBuilder := &flowBuilder{
		posX:                   innerStartX,
		posY:                   innerStartY,
		baseY:                  innerStartY,
		spacing:                HorizontalSpacing,
		returnType:             fb.returnType,
		hasReturnValue:         fb.hasReturnValue,
		varTypes:               fb.varTypes,
		declaredVars:           fb.declaredVars,
		measurer:               fb.measurer,
		backend:                fb.backend,
		hierarchy:              fb.hierarchy,
		restServices:           fb.restServices,
		callOutputDeclarations: fb.callOutputDeclarations,
	}

	var lastBodyID model.ID
	var pendingBodyCase string
	var pendingBodyAnchor *ast.FlowAnchors
	var prevBodyAnchor *ast.FlowAnchors
	for i, stmt := range s.Body {
		thisAnchor := stmtOwnAnchor(stmt)
		actID := loopBuilder.addStatement(stmt)
		if actID != "" {
			loopBuilder.applyPendingAnnotations(actID)
			if lastBodyID != "" {
				var flow *microflows.SequenceFlow
				originAnchor := prevBodyAnchor
				destAnchor := thisAnchor
				if pendingBodyCase != "" {
					flow = newHorizontalFlowWithCase(lastBodyID, actID, pendingBodyCase)
					originAnchor, destAnchor = pendingFlowAnchors(prevBodyAnchor, pendingBodyAnchor, thisAnchor)
					pendingBodyCase = ""
					pendingBodyAnchor = nil
				} else {
					flow = newHorizontalFlow(lastBodyID, actID)
				}
				applyUserAnchors(flow, originAnchor, destAnchor)
				loopBuilder.flows = append(loopBuilder.flows, flow)
				loopBuilder.addPendingErrorHandlerFlowForStatement(flow.OriginID, flow.DestinationID, stmt, statementsReferenceVar(s.Body[i+1:], loopBuilder.errorHandlerSkipVar))
			}
			prevBodyAnchor = thisAnchor
			if loopBuilder.nextConnectionPoint != "" {
				lastBodyID = loopBuilder.nextConnectionPoint
				loopBuilder.nextConnectionPoint = ""
				pendingBodyCase = loopBuilder.nextFlowCase
				loopBuilder.nextFlowCase = ""
				pendingBodyAnchor = loopBuilder.nextFlowAnchor
				loopBuilder.nextFlowAnchor = nil
			} else {
				lastBodyID = actID
			}
		}
	}

	whileExpr := fb.exprToString(s.Condition)

	loop := &microflows.LoopedActivity{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: loopCenterX, Y: fb.posY},
			Size:        model.Size{Width: loopWidth, Height: loopHeight},
		},
		LoopSource: &microflows.WhileLoopCondition{
			BaseElement:     model.BaseElement{ID: model.ID(types.GenerateID())},
			WhileExpression: whileExpr,
		},
		ObjectCollection: &microflows.MicroflowObjectCollection{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Objects:     loopBuilder.objects,
			Flows:       nil,
		},
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
	}

	// See addLoopStatement — @anchor(iterator/tail) is parsed but not
	// serialised, since Studio Pro does not permit explicit edges between a
	// LoopedActivity and its body statements.

	fb.objects = append(fb.objects, loop)
	fb.flows = append(fb.flows, loopBuilder.flows...)

	if savedWhileAnnotations != nil {
		fb.applyAnnotations(loop.ID, savedWhileAnnotations)
	}

	fb.posX = loopLeftX + loopWidth + HorizontalSpacing

	return loop.ID
}

func (fb *flowBuilder) addBreakEvent() model.ID {
	event := &microflows.BreakEvent{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX, Y: fb.posY},
			Size:        model.Size{Width: EventSize, Height: EventSize},
		},
	}
	fb.objects = append(fb.objects, event)
	fb.posX += fb.spacing
	return event.ID
}

func (fb *flowBuilder) addContinueEvent() model.ID {
	if fb.manualLoopBackTarget != "" {
		return fb.manualLoopBackTarget
	}

	event := &microflows.ContinueEvent{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX, Y: fb.posY},
			Size:        model.Size{Width: EventSize, Height: EventSize},
		},
	}
	fb.objects = append(fb.objects, event)
	fb.posX += fb.spacing
	return event.ID
}
