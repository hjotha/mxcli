// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"strconv"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/grammar/parser"
)

// ============================================================================
// REST Client CREATE Statements
// ============================================================================

// ExitCreateRestClientStatement handles the property-based CREATE REST CLIENT syntax.
func (b *Builder) ExitCreateRestClientStatement(ctx *parser.CreateRestClientStatementContext) {
	stmt := &ast.CreateRestClientStmt{
		Name: buildQualifiedName(ctx.QualifiedName()),
	}

	// Parse service-level properties (BaseUrl, Authentication, Folder)
	for _, propCtx := range ctx.AllRestClientProperty() {
		pc := propCtx.(*parser.RestClientPropertyContext)
		key := strings.ToLower(identifierOrKeywordText(pc.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext)))
		switch key {
		case "baseurl":
			if sl := pc.STRING_LITERAL(); sl != nil {
				stmt.BaseUrl = unquoteString(sl.GetText())
			}
		case "folder":
			if sl := pc.STRING_LITERAL(); sl != nil {
				stmt.Folder = unquoteString(sl.GetText())
			}
		case "authentication":
			if pc.BASIC() != nil {
				authDef := &ast.RestAuthDef{Scheme: "BASIC"}
				for _, subProp := range pc.AllRestClientProperty() {
					sp := subProp.(*parser.RestClientPropertyContext)
					subKey := strings.ToLower(identifierOrKeywordText(sp.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext)))
					if sl := sp.STRING_LITERAL(); sl != nil {
						switch subKey {
						case "username":
							authDef.Username = unquoteString(sl.GetText())
						case "password":
							authDef.Password = unquoteString(sl.GetText())
						}
					}
				}
				stmt.Authentication = authDef
			}
			// NONE: leave Authentication nil
		}
	}

	// Parse operations
	for _, opCtx := range ctx.AllRestClientOperation() {
		oc := opCtx.(*parser.RestClientOperationContext)
		opDef := parseRestClientOperation(oc)
		stmt.Operations = append(stmt.Operations, opDef)
	}

	// Check for CREATE OR MODIFY
	createStmt := findParentCreateStatement(ctx)
	if createStmt != nil {
		if createStmt.OR() != nil && (createStmt.MODIFY() != nil || createStmt.REPLACE() != nil) {
			stmt.CreateOrModify = true
		}
	}
	stmt.Documentation = findDocCommentText(ctx)

	b.statements = append(b.statements, stmt)
}

// parseRestClientOperation parses a single OPERATION { Key: Value, ... } block.
func parseRestClientOperation(ctx *parser.RestClientOperationContext) *ast.RestOperationDef {
	op := &ast.RestOperationDef{}

	// Operation name
	if iok := ctx.IdentifierOrKeyword(); iok != nil {
		op.Name = identifierOrKeywordText(iok)
	} else if sl := ctx.STRING_LITERAL(); sl != nil {
		op.Name = unquoteString(sl.GetText())
	}

	// Documentation
	if docCtx := ctx.DocComment(); docCtx != nil {
		op.Documentation = extractDocCommentText(docCtx)
	}

	// Parse properties
	for _, propCtx := range ctx.AllRestClientOpProp() {
		pc := propCtx.(*parser.RestClientOpPropContext)
		parseRestClientOpProp(pc, op)
	}

	return op
}

// parseRestClientOpProp dispatches a single operation property to the right handler.
func parseRestClientOpProp(ctx *parser.RestClientOpPropContext, op *ast.RestOperationDef) {
	// Determine key name from the identifierOrKeyword
	iokCtx := ctx.IdentifierOrKeyword()
	if iokCtx == nil {
		return
	}
	key := strings.ToLower(identifierOrKeywordText(iokCtx.(*parser.IdentifierOrKeywordContext)))

	// Method: GET
	if mCtx := ctx.RestHttpMethod(); mCtx != nil {
		op.Method = strings.ToUpper(mCtx.GetText())
		return
	}

	// MAPPING qualifiedName { ... }
	if ctx.MAPPING() != nil && ctx.QualifiedName() != nil {
		md := &ast.RestMappingDef{
			Entity: buildQualifiedName(ctx.QualifiedName()),
		}
		for _, entryCtx := range ctx.AllRestClientMappingEntry() {
			md.Entries = append(md.Entries, parseRestClientMappingEntry(entryCtx.(*parser.RestClientMappingEntryContext)))
		}
		switch key {
		case "body":
			op.BodyType = "MAPPING"
			op.BodyMapping = md
		case "response":
			op.ResponseType = "MAPPING"
			op.ResponseMapping = md
		}
		return
	}

	// TEMPLATE 'string'
	if ctx.TEMPLATE() != nil {
		if sl := ctx.STRING_LITERAL(); sl != nil {
			op.BodyType = "TEMPLATE"
			op.BodyVariable = unquoteString(sl.GetText())
		}
		return
	}

	// (JSON|FILE|STRING|STATUS) (FROM|AS) $var
	if ctx.VARIABLE() != nil {
		varName := ctx.VARIABLE().GetText()
		if ctx.FROM() != nil {
			// Body: JSON FROM $var, FILE FROM $var
			if ctx.JSON() != nil {
				op.BodyType = "JSON"
			} else if ctx.FILE_KW() != nil {
				op.BodyType = "FILE"
			}
			op.BodyVariable = varName
		} else if ctx.AS() != nil {
			// Response: JSON AS $var, STRING AS $var, etc.
			if ctx.JSON() != nil {
				op.ResponseType = "JSON"
			} else if ctx.STRING_TYPE() != nil {
				op.ResponseType = "STRING"
			} else if ctx.FILE_KW() != nil {
				op.ResponseType = "FILE"
			} else if ctx.STATUS() != nil {
				op.ResponseType = "STATUS"
			}
			op.ResponseVariable = varName
		}
		return
	}

	// NONE
	if ctx.NONE() != nil {
		switch key {
		case "response":
			op.ResponseType = "NONE"
		case "authentication":
			// handled at service level
		}
		return
	}

	// Param list: ($var: Type, ...)
	paramItems := ctx.AllRestClientParamItem()
	if len(paramItems) > 0 {
		for _, pi := range paramItems {
			pic := pi.(*parser.RestClientParamItemContext)
			param := ast.RestParamDef{}
			if v := pic.VARIABLE(); v != nil {
				param.Name = v.GetText()
			}
			if dt := pic.DataType(); dt != nil {
				param.DataType = buildDataType(dt).Kind.String()
			}
			switch key {
			case "parameters":
				op.Parameters = append(op.Parameters, param)
			case "query":
				op.QueryParameters = append(op.QueryParameters, param)
			}
		}
		return
	}

	// Header list: ('Name' = 'Value', ...)
	headerItems := ctx.AllRestClientHeaderItem()
	if len(headerItems) > 0 {
		for _, hi := range headerItems {
			hic := hi.(*parser.RestClientHeaderItemContext)
			header := ast.RestHeaderDef{}
			allSL := hic.AllSTRING_LITERAL()
			if len(allSL) >= 1 {
				header.Name = unquoteString(allSL[0].GetText())
			}
			if hic.PLUS() != nil {
				// 'prefix' + $Variable
				if len(allSL) >= 2 {
					header.Prefix = unquoteString(allSL[1].GetText())
				}
				if v := hic.VARIABLE(); v != nil {
					header.Variable = v.GetText()
				}
			} else if hic.VARIABLE() != nil {
				header.Variable = hic.VARIABLE().GetText()
			} else if len(allSL) >= 2 {
				header.Value = unquoteString(allSL[1].GetText())
			}
			op.Headers = append(op.Headers, header)
		}
		return
	}

	// String property: Path: '/items'
	if sl := ctx.STRING_LITERAL(); sl != nil {
		switch key {
		case "path":
			op.Path = unquoteString(sl.GetText())
		}
		return
	}

	// Number property: Timeout: 30
	if nl := ctx.NUMBER_LITERAL(); nl != nil {
		switch key {
		case "timeout":
			if val, err := strconv.Atoi(nl.GetText()); err == nil {
				op.Timeout = val
			}
		}
		return
	}
}

// parseRestClientMappingEntry parses a single mapping entry (value or object).
func parseRestClientMappingEntry(ctx *parser.RestClientMappingEntryContext) ast.RestMappingEntry {
	allQN := ctx.AllQualifiedName()
	allIOK := ctx.AllIdentifierOrKeyword()

	// Object mapping: [CREATE] Association/Entity = exposedName { ... }
	if len(allQN) >= 2 {
		entry := ast.RestMappingEntry{
			Create:      ctx.CREATE() != nil,
			Association: buildQualifiedName(allQN[0]),
			Entity:      buildQualifiedName(allQN[1]),
		}
		// ExposedName is the identifierOrKeyword after EQUALS
		if len(allIOK) > 0 {
			entry.ExposedName = identifierOrKeywordText(allIOK[0].(*parser.IdentifierOrKeywordContext))
		}
		for _, childCtx := range ctx.AllRestClientMappingEntry() {
			entry.Children = append(entry.Children, parseRestClientMappingEntry(childCtx.(*parser.RestClientMappingEntryContext)))
		}
		return entry
	}

	// Value mapping: Left = Right
	entry := ast.RestMappingEntry{}
	if len(allIOK) >= 2 {
		entry.Left = identifierOrKeywordText(allIOK[0].(*parser.IdentifierOrKeywordContext))
		entry.Right = identifierOrKeywordText(allIOK[1].(*parser.IdentifierOrKeywordContext))
	}
	return entry
}

// extractDocCommentText extracts documentation text from a DocComment context.
func extractDocCommentText(ctx parser.IDocCommentContext) string {
	if ctx == nil {
		return ""
	}
	text := ctx.GetText()
	// Remove /** and */ markers
	text = strings.TrimPrefix(text, "/**")
	text = strings.TrimSuffix(text, "*/")

	// Process lines: trim whitespace and leading *
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// ExitCreatePublishedRestServiceStatement handles CREATE PUBLISHED REST SERVICE.
func (b *Builder) ExitCreatePublishedRestServiceStatement(ctx *parser.CreatePublishedRestServiceStatementContext) {
	stmt := &ast.CreatePublishedRestServiceStmt{
		Name: buildQualifiedName(ctx.QualifiedName()),
	}

	// Check for CREATE OR REPLACE
	createStmt := findParentCreateStatement(ctx)
	if createStmt != nil {
		if createStmt.OR() != nil && (createStmt.REPLACE() != nil || createStmt.MODIFY() != nil) {
			stmt.CreateOrReplace = true
		}
	}

	// Parse properties (Path, Version, ServiceName)
	for _, propCtx := range ctx.AllPublishedRestProperty() {
		pc := propCtx.(*parser.PublishedRestPropertyContext)
		key := identifierOrKeywordText(pc.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext))
		val := unquoteString(pc.STRING_LITERAL().GetText())
		switch strings.ToLower(key) {
		case "path":
			stmt.Path = val
		case "version":
			stmt.Version = val
		case "servicename":
			stmt.ServiceName = val
		case "folder":
			stmt.Folder = val
		}
	}

	// Parse resources
	for _, resCtx := range ctx.AllPublishedRestResource() {
		rc := resCtx.(*parser.PublishedRestResourceContext)
		stmt.Resources = append(stmt.Resources, buildPublishedRestResourceDef(rc))
	}

	b.statements = append(b.statements, stmt)
}

// buildPublishedRestResourceDef converts a PublishedRestResourceContext to a
// PublishedRestResourceDef AST node. Shared by CREATE and ALTER.
func buildPublishedRestResourceDef(rc *parser.PublishedRestResourceContext) *ast.PublishedRestResourceDef {
	resDef := &ast.PublishedRestResourceDef{
		Name: unquoteString(rc.STRING_LITERAL().GetText()),
	}

	for _, opCtx := range rc.AllPublishedRestOperation() {
		oc := opCtx.(*parser.PublishedRestOperationContext)
		opDef := &ast.PublishedRestOperationDef{}

		// HTTP method
		if mCtx := oc.RestHttpMethod(); mCtx != nil {
			opDef.HTTPMethod = strings.ToUpper(mCtx.GetText())
		}

		// Operation path — strip leading/trailing slashes (CE6550/CE6551)
		if pCtx := oc.PublishedRestOpPath(); pCtx != nil {
			pc := pCtx.(*parser.PublishedRestOpPathContext)
			if pc.STRING_LITERAL() != nil {
				opDef.Path = strings.Trim(unquoteString(pc.STRING_LITERAL().GetText()), "/")
			}
		}

		// Microflow reference
		allQN := oc.AllQualifiedName()
		if len(allQN) >= 1 {
			opDef.Microflow = buildQualifiedName(allQN[0])
		}

		if oc.DEPRECATED() != nil {
			opDef.Deprecated = true
		}

		// Import/Export mapping (qualifiedName after IMPORT/EXPORT MAPPING)
		if oc.IMPORT() != nil && len(allQN) >= 2 {
			opDef.ImportMapping = allQN[1].GetText()
		}
		if oc.EXPORT() != nil {
			idx := 1
			if oc.IMPORT() != nil {
				idx = 2
			}
			if len(allQN) > idx {
				opDef.ExportMapping = allQN[idx].GetText()
			}
		}

		if oc.COMMIT() != nil {
			if idCtx := oc.IdentifierOrKeyword(); idCtx != nil {
				opDef.Commit = identifierOrKeywordText(idCtx.(*parser.IdentifierOrKeywordContext))
			}
		}

		resDef.Operations = append(resDef.Operations, opDef)
	}

	return resDef
}

// exitAlterPublishedRestServiceStatement handles ALTER PUBLISHED REST SERVICE.
func (b *Builder) exitAlterPublishedRestServiceStatement(ctx *parser.AlterStatementContext) {
	qn := ctx.QualifiedName()
	if qn == nil {
		return
	}

	stmt := &ast.AlterPublishedRestServiceStmt{
		Name: buildQualifiedName(qn),
	}

	for _, actCtx := range ctx.AllAlterPublishedRestServiceAction() {
		ac, ok := actCtx.(*parser.AlterPublishedRestServiceActionContext)
		if !ok {
			continue
		}

		// SET key = 'value' [, ...]
		if ac.SET() != nil {
			changes := make(map[string]string)
			for _, asnCtx := range ac.AllPublishedRestAlterAssignment() {
				asn := asnCtx.(*parser.PublishedRestAlterAssignmentContext)
				key := identifierOrKeywordText(asn.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext))
				val := unquoteString(asn.STRING_LITERAL().GetText())
				changes[key] = val
			}
			stmt.Actions = append(stmt.Actions, &ast.PublishedRestSetAction{Changes: changes})
			continue
		}

		// ADD RESOURCE 'name' { ... }
		if ac.ADD() != nil {
			if rc := ac.PublishedRestResource(); rc != nil {
				resDef := buildPublishedRestResourceDef(rc.(*parser.PublishedRestResourceContext))
				stmt.Actions = append(stmt.Actions, &ast.PublishedRestAddResourceAction{Resource: resDef})
			}
			continue
		}

		// DROP RESOURCE 'name'
		if ac.DROP() != nil && ac.RESOURCE() != nil {
			name := unquoteString(ac.STRING_LITERAL().GetText())
			stmt.Actions = append(stmt.Actions, &ast.PublishedRestDropResourceAction{Name: name})
			continue
		}
	}

	b.statements = append(b.statements, stmt)
}
