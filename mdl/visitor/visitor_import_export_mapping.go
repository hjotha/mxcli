// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/grammar/parser"
)

// ExitCreateImportMappingStatement is called when exiting the createImportMappingStatement production.
func (b *Builder) ExitCreateImportMappingStatement(ctx *parser.CreateImportMappingStatementContext) {
	stmt := &ast.CreateImportMappingStmt{
		Name: buildQualifiedName(ctx.QualifiedName()),
	}

	// Parse schema clause
	if schemaCtx := ctx.ImportMappingSchemaClause(); schemaCtx != nil {
		sc := schemaCtx.(*parser.ImportMappingSchemaClauseContext)
		if sc.JSON() != nil {
			stmt.SchemaKind = "JSON_STRUCTURE"
		} else {
			stmt.SchemaKind = "XML_SCHEMA"
		}
		if sc.QualifiedName() != nil {
			stmt.SchemaRef = buildQualifiedName(sc.QualifiedName())
		}
	}

	// Parse the root mapping element
	if elemCtx := ctx.ImportMappingElement(); elemCtx != nil {
		stmt.RootElement = buildImportMappingElement(elemCtx.(*parser.ImportMappingElementContext))
	}

	b.statements = append(b.statements, stmt)
}

// buildImportMappingElement converts an importMappingElement context to an AST node.
func buildImportMappingElement(ctx *parser.ImportMappingElementContext) *ast.ImportMappingElementDef {
	elem := &ast.ImportMappingElementDef{}

	allQN := ctx.AllQualifiedName()
	allIdent := ctx.AllIdentifierOrKeyword()

	// JSON field name (left side of AS): strip QUOTED_IDENTIFIER delimiters.
	if len(allIdent) > 0 {
		elem.JsonName = identifierOrKeywordText(allIdent[0])
	}

	// Check if this is an object mapping (has qualifiedName RHS with entity)
	// or value mapping (has identifierOrKeyword RHS with attribute name + type in parens)
	if ctx.ImportMappingHandling() != nil {
		// Object mapping: IDENTIFIER AS qualifiedName LPAREN handling RPAREN
		if len(allQN) >= 1 {
			elem.Entity = allQN[0].GetText()
		}
		handlingCtx := ctx.ImportMappingHandling().(*parser.ImportMappingHandlingContext)
		elem.ObjectHandling = extractImportMappingHandling(handlingCtx)

		// VIA clause: second qualifiedName
		if ctx.VIA() != nil && len(allQN) >= 2 {
			elem.Association = allQN[1].GetText()
		}

		// Nested children
		for _, childCtx := range ctx.AllImportMappingElement() {
			child := buildImportMappingElement(childCtx.(*parser.ImportMappingElementContext))
			elem.Children = append(elem.Children, child)
		}
	} else {
		// Value mapping: IDENTIFIER AS identifierOrKeyword LPAREN type (COMMA KEY)? RPAREN
		if len(allIdent) >= 2 {
			elem.Attribute = identifierOrKeywordText(allIdent[1])
		}
		if vtCtx := ctx.ImportMappingValueType(); vtCtx != nil {
			elem.DataType = extractImportValueType(vtCtx.(*parser.ImportMappingValueTypeContext))
		}
		if ctx.KEY() != nil {
			elem.IsKey = true
		}
	}

	return elem
}

// ExitCreateExportMappingStatement is called when exiting the createExportMappingStatement production.
func (b *Builder) ExitCreateExportMappingStatement(ctx *parser.CreateExportMappingStatementContext) {
	stmt := &ast.CreateExportMappingStmt{
		Name: buildQualifiedName(ctx.QualifiedName()),
	}

	// Parse schema clause
	if schemaCtx := ctx.ExportMappingSchemaClause(); schemaCtx != nil {
		sc := schemaCtx.(*parser.ExportMappingSchemaClauseContext)
		if sc.JSON() != nil {
			stmt.SchemaKind = "JSON_STRUCTURE"
		} else {
			stmt.SchemaKind = "XML_SCHEMA"
		}
		if sc.QualifiedName() != nil {
			stmt.SchemaRef = buildQualifiedName(sc.QualifiedName())
		}
	}

	// Parse null values clause
	if nullCtx := ctx.ExportMappingNullValuesClause(); nullCtx != nil {
		nc := nullCtx.(*parser.ExportMappingNullValuesClauseContext)
		if nc.IdentifierOrKeyword() != nil {
			stmt.NullValueOption = identifierOrKeywordText(nc.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext))
		}
	}

	// Parse the root mapping element
	if elemCtx := ctx.ExportMappingElement(); elemCtx != nil {
		stmt.RootElement = buildExportMappingElement(elemCtx.(*parser.ExportMappingElementContext))
	}

	b.statements = append(b.statements, stmt)
}

// buildExportMappingElement converts an exportMappingElement context to an AST node.
func buildExportMappingElement(ctx *parser.ExportMappingElementContext) *ast.ExportMappingElementDef {
	elem := &ast.ExportMappingElementDef{}

	// Distinguish object vs value element by presence of importMappingValueType
	if ctx.ImportMappingValueType() != nil {
		// Value mapping: identifierOrKeyword AS identifierOrKeyword (Type)
		// AllIdentifierOrKeyword()[0] = attribute name, [1] = JSON name
		allIdent := ctx.AllIdentifierOrKeyword()
		if len(allIdent) >= 1 {
			elem.Attribute = identifierOrKeywordText(allIdent[0].(*parser.IdentifierOrKeywordContext))
		}
		if len(allIdent) >= 2 {
			elem.JsonName = identifierOrKeywordText(allIdent[1].(*parser.IdentifierOrKeywordContext))
		}
		elem.DataType = extractImportValueType(ctx.ImportMappingValueType().(*parser.ImportMappingValueTypeContext))
	} else {
		// Object mapping: qualifiedName [VIA qualifiedName] AS identifierOrKeyword { children }
		allQN := ctx.AllQualifiedName()
		if len(allQN) >= 1 {
			elem.Entity = allQN[0].GetText()
		}
		if ctx.VIA() != nil && len(allQN) >= 2 {
			elem.Association = allQN[1].GetText()
		}
		// The identifierOrKeyword after AS is the JSON exposed name
		allIdent := ctx.AllIdentifierOrKeyword()
		if len(allIdent) >= 1 {
			elem.JsonName = identifierOrKeywordText(allIdent[0].(*parser.IdentifierOrKeywordContext))
		}
		// Nested children
		for _, childCtx := range ctx.AllExportMappingElement() {
			child := buildExportMappingElement(childCtx.(*parser.ExportMappingElementContext))
			elem.Children = append(elem.Children, child)
		}
	}

	return elem
}

// extractImportMappingHandling extracts the handling string from the grammar context.
func extractImportMappingHandling(ctx *parser.ImportMappingHandlingContext) string {
	if ctx.CREATE() != nil {
		return "Create"
	}
	if ctx.FIND() != nil {
		return "Find"
	}
	if ctx.UPDATE() != nil {
		return "FindOrCreate"
	}
	if ctx.IDENTIFIER() != nil {
		return ctx.IDENTIFIER().GetText()
	}
	return "Create"
}

// extractImportValueType maps a grammar type keyword to a string.
func extractImportValueType(ctx *parser.ImportMappingValueTypeContext) string {
	if ctx.STRING_TYPE() != nil {
		return "String"
	}
	if ctx.INTEGER_TYPE() != nil {
		return "Integer"
	}
	if ctx.LONG_TYPE() != nil {
		return "Long"
	}
	if ctx.DECIMAL_TYPE() != nil {
		return "Decimal"
	}
	if ctx.BOOLEAN_TYPE() != nil {
		return "Boolean"
	}
	if ctx.DATETIME_TYPE() != nil {
		return "DateTime"
	}
	if ctx.DATE_TYPE() != nil {
		return "Date"
	}
	if ctx.BINARY_TYPE() != nil {
		return "Binary"
	}
	return "String"
}
