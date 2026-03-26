// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow helper functions
package executor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// convertASTToMicroflowDataType converts an AST DataType to a microflows.DataType.
// entityResolver is optional - if provided, it resolves entity qualified names to IDs.
func convertASTToMicroflowDataType(dt ast.DataType, entityResolver func(ast.QualifiedName) model.ID) microflows.DataType {
	switch dt.Kind {
	case ast.TypeBoolean:
		return &microflows.BooleanType{}
	case ast.TypeInteger:
		return &microflows.IntegerType{}
	case ast.TypeLong:
		return &microflows.LongType{}
	case ast.TypeDecimal:
		return &microflows.DecimalType{}
	case ast.TypeString:
		return &microflows.StringType{}
	case ast.TypeDateTime:
		return &microflows.DateTimeType{}
	case ast.TypeDate:
		return &microflows.DateType{}
	case ast.TypeBinary:
		return &microflows.BinaryType{}
	case ast.TypeVoid:
		return &microflows.VoidType{}
	case ast.TypeEntity:
		lt := &microflows.ObjectType{}
		if dt.EntityRef != nil {
			// Set qualified name for BY_NAME_REFERENCE serialization
			lt.EntityQualifiedName = dt.EntityRef.Module + "." + dt.EntityRef.Name
			if entityResolver != nil {
				lt.EntityID = entityResolver(*dt.EntityRef)
			}
		}
		return lt
	case ast.TypeListOf:
		lt := &microflows.ListType{}
		if dt.EntityRef != nil {
			// Set qualified name for BY_NAME_REFERENCE serialization
			lt.EntityQualifiedName = dt.EntityRef.Module + "." + dt.EntityRef.Name
			if entityResolver != nil {
				lt.EntityID = entityResolver(*dt.EntityRef)
			}
		}
		return lt
	case ast.TypeEnumeration:
		et := &microflows.EnumerationType{}
		if dt.EnumRef != nil {
			// Set qualified name for BY_NAME_REFERENCE serialization
			et.EnumerationQualifiedName = dt.EnumRef.Module + "." + dt.EnumRef.Name
		}
		return et
	default:
		return &microflows.VoidType{}
	}
}

// expressionToString converts an AST Expression to a Mendix expression string.
func expressionToString(expr ast.Expression) string {
	// Check for nil interface
	if expr == nil {
		return ""
	}

	// Use reflection to check for nil pointer inside interface
	// This handles the Go interface gotcha where the type is set but pointer is nil
	if reflect.ValueOf(expr).IsNil() {
		return ""
	}

	switch e := expr.(type) {
	case *ast.LiteralExpr:
		switch e.Kind {
		case ast.LiteralString:
			// Escape single quotes for Mendix expression syntax (use '' inside strings)
			strVal := fmt.Sprintf("%v", e.Value)
			strVal = strings.ReplaceAll(strVal, `'`, `''`)
			return "'" + strVal + "'"
		case ast.LiteralBoolean:
			if e.Value.(bool) {
				return "true"
			}
			return "false"
		case ast.LiteralNull:
			return "empty"
		default:
			return fmt.Sprintf("%v", e.Value)
		}
	case *ast.VariableExpr:
		return "$" + e.Name
	case *ast.AttributePathExpr:
		return "$" + e.Variable + "/" + strings.Join(e.Path, "/")
	case *ast.BinaryExpr:
		left := expressionToString(e.Left)
		right := expressionToString(e.Right)
		// Mendix expressions use lowercase operators (and, or, div, mod)
		op := strings.ToLower(e.Operator)
		return left + " " + op + " " + right
	case *ast.UnaryExpr:
		operand := expressionToString(e.Operand)
		// Mendix expressions use lowercase operators (not)
		op := strings.ToLower(e.Operator)
		return op + " " + operand
	case *ast.FunctionCallExpr:
		var args []string
		for _, arg := range e.Arguments {
			args = append(args, expressionToString(arg))
		}
		return e.Name + "(" + strings.Join(args, ", ") + ")"
	case *ast.TokenExpr:
		return "[%" + e.Token + "%]"
	case *ast.ParenExpr:
		return "(" + expressionToString(e.Inner) + ")"
	case *ast.IdentifierExpr:
		// Unquoted identifier (attribute name in XPath)
		return e.Name
	case *ast.QualifiedNameExpr:
		// Qualified name (association name, entity reference) - unquoted
		return e.QualifiedName.String()
	case *ast.IfThenElseExpr:
		cond := expressionToString(e.Condition)
		thenStr := expressionToString(e.ThenExpr)
		elseStr := expressionToString(e.ElseExpr)
		return "if " + cond + " then " + thenStr + " else " + elseStr
	default:
		return ""
	}
}

// expressionToXPath converts an AST Expression to an XPath constraint string.
// Mendix tokens like [%CurrentDateTime%] are written unquoted — they are
// special XPath placeholders, not string literals.
func expressionToXPath(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	if reflect.ValueOf(expr).IsNil() {
		return ""
	}

	switch e := expr.(type) {
	case *ast.TokenExpr:
		return "[%" + e.Token + "%]"
	case *ast.BinaryExpr:
		left := expressionToXPath(e.Left)
		right := expressionToXPath(e.Right)
		op := strings.ToLower(e.Operator)
		return left + " " + op + " " + right
	case *ast.UnaryExpr:
		operand := expressionToXPath(e.Operand)
		op := strings.ToLower(e.Operator)
		// For 'not' with parenthesized operand, output as not(expr)
		if op == "not" {
			if p, ok := e.Operand.(*ast.ParenExpr); ok {
				return "not(" + expressionToXPath(p.Inner) + ")"
			}
			return "not(" + operand + ")"
		}
		return op + " " + operand
	case *ast.ParenExpr:
		return "(" + expressionToXPath(e.Inner) + ")"
	case *ast.XPathPathExpr:
		return xpathPathExprToString(e)
	case *ast.FunctionCallExpr:
		var args []string
		for _, arg := range e.Arguments {
			args = append(args, expressionToXPath(arg))
		}
		return e.Name + "(" + strings.Join(args, ", ") + ")"
	case *ast.LiteralExpr:
		if e.Kind == ast.LiteralEmpty {
			return "empty"
		}
		return expressionToString(expr)
	default:
		// For all other expression types, the standard serialization is correct
		return expressionToString(expr)
	}
}

// xpathPathExprToString serializes an XPathPathExpr to an XPath path string.
func xpathPathExprToString(path *ast.XPathPathExpr) string {
	var parts []string
	for _, step := range path.Steps {
		s := expressionToXPath(step.Expr)
		if step.Predicate != nil {
			s += "[" + expressionToXPath(step.Predicate) + "]"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "/")
}

// countMicroflowActivities counts the number of meaningful activities in a microflow.
// Excludes structural elements like StartEvent, EndEvent, and merge nodes.
func countMicroflowActivities(mf *microflows.Microflow) int {
	if mf.ObjectCollection == nil {
		return 0
	}

	count := 0
	for _, obj := range mf.ObjectCollection.Objects {
		switch obj.(type) {
		case *microflows.StartEvent, *microflows.EndEvent:
			// Don't count start/end events
		case *microflows.ExclusiveMerge:
			// Don't count merge nodes (they're structural)
		default:
			// Count all other activities (ActionActivity, ExclusiveSplit, LoopedActivity, etc.)
			count++
		}
	}
	return count
}

// calculateMicroflowComplexity calculates the McCabe cyclomatic complexity of a microflow.
// McCabe complexity = 1 + number of decision points (IF, LOOP, error handlers)
// A higher complexity indicates more paths through the code and higher testing burden.
// Typical thresholds: 1-10 (simple), 11-20 (moderate), 21-50 (complex), 50+ (untestable)
func calculateMicroflowComplexity(mf *microflows.Microflow) int {
	// Base complexity is 1 (the main path through the microflow)
	complexity := 1

	if mf.ObjectCollection == nil {
		return complexity
	}

	// Count decision points in the main flow
	complexity += countMicroflowDecisionPoints(mf.ObjectCollection.Objects)

	return complexity
}

// countMicroflowDecisionPoints counts decision points in a list of microflow objects.
// This recursively processes nested structures like LoopedActivity.
func countMicroflowDecisionPoints(objects []microflows.MicroflowObject) int {
	count := 0

	for _, obj := range objects {
		switch activity := obj.(type) {
		case *microflows.ExclusiveSplit:
			// Each IF/decision adds 1 to complexity
			count++

		case *microflows.InheritanceSplit:
			// Type check split adds 1 to complexity
			count++

		case *microflows.LoopedActivity:
			// Each loop adds 1 to complexity
			count++
			// Also count decision points inside the loop body
			if activity.ObjectCollection != nil {
				count += countMicroflowDecisionPoints(activity.ObjectCollection.Objects)
			}

		case *microflows.ErrorEvent:
			// Error handling path adds complexity
			count++
		}
	}

	return count
}
