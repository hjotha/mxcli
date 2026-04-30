// SPDX-License-Identifier: Apache-2.0

package executor

import "github.com/mendixlabs/mxcli/mdl/ast"

// referencedVariableSet returns the set of variable names referenced anywhere
// in the given statement bodies.
func referencedVariableSet(stmts []ast.MicroflowStatement) map[string]bool {
	refs := make(map[string]bool)
	for _, ref := range statementsVarRefs(stmts) {
		if ref != "" {
			refs[ref] = true
		}
	}
	return refs
}

// statementsVarRefs returns the variable names referenced by the given
// statement bodies, including nested branches and bodies.
func statementsVarRefs(stmts []ast.MicroflowStatement) []string {
	var refs []string
	for _, stmt := range stmts {
		refs = append(refs, statementVarRefs(stmt)...)
	}
	return refs
}

// statementVarRefs returns the variable names a single statement reads. It is
// the canonical reference walker used by the variable-alias planner and any
// other builder pass that needs to know whether a particular variable is
// consumed downstream.
//
// New microflow statement types should be added here so all callers stay in
// sync. Statements that only PRODUCE a variable (and never read one) are
// allowed to fall through to the default no-op case.
func statementVarRefs(stmt ast.MicroflowStatement) []string {
	var refs []string
	switch s := stmt.(type) {
	case *ast.DeclareStmt:
		refs = append(refs, exprVarRefs(s.InitialValue)...)
	case *ast.ReturnStmt:
		refs = append(refs, exprVarRefs(s.Value)...)
	case *ast.LogStmt:
		refs = append(refs, exprVarRefs(s.Node)...)
		refs = append(refs, exprVarRefs(s.Message)...)
		for _, param := range s.Template {
			refs = append(refs, exprVarRefs(param.Value)...)
		}
	case *ast.IfStmt:
		refs = append(refs, exprVarRefs(s.Condition)...)
		refs = append(refs, statementsVarRefs(s.ThenBody)...)
		refs = append(refs, statementsVarRefs(s.ElseBody)...)
	case *ast.WhileStmt:
		refs = append(refs, exprVarRefs(s.Condition)...)
		refs = append(refs, statementsVarRefs(s.Body)...)
	case *ast.LoopStmt:
		refs = append(refs, s.ListVariable)
		refs = append(refs, statementsVarRefs(s.Body)...)
	case *ast.MfSetStmt:
		refs = append(refs, extractVarName(s.Target))
		refs = append(refs, exprVarRefs(s.Value)...)
	case *ast.ChangeObjectStmt:
		refs = append(refs, s.Variable)
		for _, change := range s.Changes {
			refs = append(refs, exprVarRefs(change.Value)...)
		}
	case *ast.CreateObjectStmt:
		for _, change := range s.Changes {
			refs = append(refs, exprVarRefs(change.Value)...)
		}
	case *ast.RetrieveStmt:
		if s.StartVariable != "" {
			refs = append(refs, s.StartVariable)
		}
		refs = append(refs, exprVarRefs(s.Where)...)
	case *ast.ListOperationStmt:
		refs = append(refs, s.InputVariable, s.SecondVariable)
		refs = append(refs, exprVarRefs(s.Condition)...)
		refs = append(refs, exprVarRefs(s.OffsetExpr)...)
		refs = append(refs, exprVarRefs(s.LimitExpr)...)
	case *ast.AggregateListStmt:
		refs = append(refs, s.InputVariable)
		refs = append(refs, exprVarRefs(s.Expression)...)
	case *ast.CreateListStmt:
		// Output-only statement.
	case *ast.CallMicroflowStmt:
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.CallJavaActionStmt:
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.CallExternalActionStmt:
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.ShowPageStmt:
		refs = append(refs, s.ForObject)
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.ShowMessageStmt:
		refs = append(refs, exprVarRefs(s.Message)...)
		for _, arg := range s.TemplateArgs {
			refs = append(refs, exprVarRefs(arg)...)
		}
	case *ast.ValidationFeedbackStmt:
		if s.AttributePath != nil {
			refs = append(refs, s.AttributePath.Variable)
		}
		refs = append(refs, exprVarRefs(s.Message)...)
		for _, arg := range s.TemplateArgs {
			refs = append(refs, exprVarRefs(arg)...)
		}
	case *ast.RestCallStmt:
		refs = append(refs, exprVarRefs(s.URL)...)
		for _, param := range s.URLParams {
			refs = append(refs, exprVarRefs(param.Value)...)
		}
		for _, header := range s.Headers {
			refs = append(refs, exprVarRefs(header.Value)...)
		}
		if s.Auth != nil {
			refs = append(refs, exprVarRefs(s.Auth.Username)...)
			refs = append(refs, exprVarRefs(s.Auth.Password)...)
		}
		if s.Body != nil {
			refs = append(refs, exprVarRefs(s.Body.Template)...)
			for _, param := range s.Body.TemplateParams {
				refs = append(refs, exprVarRefs(param.Value)...)
			}
			refs = append(refs, s.Body.SourceVariable)
		}
		refs = append(refs, exprVarRefs(s.Timeout)...)
	case *ast.SendRestRequestStmt:
		for _, param := range s.Parameters {
			refs = append(refs, expressionStringVarRefs(param.Expression)...)
		}
		refs = append(refs, s.BodyVariable)
	case *ast.ImportFromMappingStmt:
		refs = append(refs, s.SourceVariable)
	case *ast.ExportToMappingStmt:
		refs = append(refs, s.SourceVariable)
	case *ast.TransformJsonStmt:
		refs = append(refs, s.InputVariable)
	case *ast.MfCommitStmt:
		refs = append(refs, s.Variable)
	case *ast.DeleteObjectStmt:
		refs = append(refs, s.Variable)
	case *ast.AddToListStmt:
		if s.Item != "" {
			refs = append(refs, s.Item)
		}
		refs = append(refs, s.List)
	case *ast.RemoveFromListStmt:
		refs = append(refs, s.Item, s.List)
	}
	return refs
}

// expressionStringVarRefs scans a raw Mendix expression string for `$Var`
// references. Used by statement walkers that store expressions as text rather
// than parsed AST nodes.
func expressionStringVarRefs(expr string) []string {
	matches := mendixExpressionVariableRefPattern.FindAllStringSubmatch(expr, -1)
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		refs = append(refs, match[1])
	}
	return refs
}
