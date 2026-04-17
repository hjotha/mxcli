// SPDX-License-Identifier: Apache-2.0

// Package executor - Enumeration commands (SHOW/DESCRIBE/CREATE/ALTER/DROP ENUMERATION)
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/mdl/linter"
	"github.com/mendixlabs/mxcli/model"
)

// execCreateEnumeration handles CREATE ENUMERATION statements.
func (e *Executor) execCreateEnumeration(s *ast.CreateEnumerationStmt) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Validate enumeration values for reserved words
	if violations := ValidateEnumeration(s); len(violations) > 0 {
		var msgs []string
		for _, v := range violations {
			msgs = append(msgs, v.Message)
		}
		return mdlerrors.NewValidationf("invalid enumeration '%s':\n  - %s",
			s.Name.String(), strings.Join(msgs, "\n  - "))
	}

	// Find or auto-create module
	module, err := e.findOrCreateModule(s.Name.Module)
	if err != nil {
		return err
	}

	// Check if enumeration already exists
	existingEnum := e.findEnumeration(s.Name.Module, s.Name.Name)
	if existingEnum != nil && !s.CreateOrModify {
		return mdlerrors.NewAlreadyExistsMsg("enumeration", s.Name.Module+"."+s.Name.Name, fmt.Sprintf("enumeration already exists: %s.%s (use CREATE OR MODIFY to update)", s.Name.Module, s.Name.Name))
	}

	// Create enumeration values
	var values []model.EnumerationValue
	for _, v := range s.Values {
		values = append(values, model.EnumerationValue{
			Name: v.Name,
			Caption: &model.Text{
				Translations: map[string]string{"en_US": v.Caption},
			},
		})
	}

	// If enumeration exists and CREATE OR MODIFY, delete it first
	if existingEnum != nil && s.CreateOrModify {
		if err := e.writer.DeleteEnumeration(existingEnum.ID); err != nil {
			return mdlerrors.NewBackend("delete existing enumeration", err)
		}
	}

	// Create enumeration
	enum := &model.Enumeration{
		ContainerID:   module.ID,
		Name:          s.Name.Name,
		Documentation: s.Documentation,
		Values:        values,
	}

	if err := e.writer.CreateEnumeration(enum); err != nil {
		return mdlerrors.NewBackend("create enumeration", err)
	}

	// Invalidate hierarchy cache so the new enumeration's container is visible
	e.invalidateHierarchy()

	fmt.Fprintf(e.output, "Created enumeration: %s\n", s.Name)
	return nil
}

// findEnumeration finds an enumeration by module and name.
func (e *Executor) findEnumeration(moduleName, enumName string) *model.Enumeration {
	enums, err := e.reader.ListEnumerations()
	if err != nil {
		return nil
	}

	h, err := e.getHierarchy()
	if err != nil {
		return nil
	}

	for _, enum := range enums {
		modID := h.FindModuleID(enum.ContainerID)
		modName := h.GetModuleName(modID)
		if enum.Name == enumName && modName == moduleName {
			return enum
		}
	}
	return nil
}

// execAlterEnumeration handles ALTER ENUMERATION statements.
func (e *Executor) execAlterEnumeration(s *ast.AlterEnumerationStmt) error {
	// TODO: Implement ALTER ENUMERATION
	return mdlerrors.NewUnsupported("ALTER ENUMERATION not yet implemented")
}

// execDropEnumeration handles DROP ENUMERATION statements.
func (e *Executor) execDropEnumeration(s *ast.DropEnumerationStmt) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Find enumeration
	enums, err := e.reader.ListEnumerations()
	if err != nil {
		return mdlerrors.NewBackend("list enumerations", err)
	}

	for _, enum := range enums {
		if enum.Name == s.Name.Name {
			// Check module matches
			module, err := e.findModuleByID(enum.ContainerID)
			if err == nil && (s.Name.Module == "" || module.Name == s.Name.Module) {
				if err := e.writer.DeleteEnumeration(enum.ID); err != nil {
					return mdlerrors.NewBackend("delete enumeration", err)
				}
				fmt.Fprintf(e.output, "Dropped enumeration: %s\n", s.Name)
				return nil
			}
		}
	}

	return mdlerrors.NewNotFound("enumeration", s.Name.String())
}

// showEnumerations handles SHOW ENUMERATIONS command.
func (e *Executor) showEnumerations(moduleName string) error {
	enums, err := e.reader.ListEnumerations()
	if err != nil {
		return mdlerrors.NewBackend("list enumerations", err)
	}

	// Get hierarchy for module/folder resolution
	h, err := e.getHierarchy()
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Collect rows
	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		values        int
	}
	var rows []row

	for _, enum := range enums {
		modID := h.FindModuleID(enum.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName == "" || modName == moduleName {
			qualifiedName := modName + "." + enum.Name
			folderPath := h.BuildFolderPath(enum.ContainerID)

			rows = append(rows, row{qualifiedName, modName, enum.Name, folderPath, len(enum.Values)})
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Build TableResult
	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Folder", "Values"},
		Summary: fmt.Sprintf("(%d enumerations)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.folderPath, r.values})
	}
	return e.writeResult(result)
}

// describeEnumeration handles DESCRIBE ENUMERATION command.
func (e *Executor) describeEnumeration(name ast.QualifiedName) error {
	enums, err := e.reader.ListEnumerations()
	if err != nil {
		return mdlerrors.NewBackend("list enumerations", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	for _, enum := range enums {
		modID := h.FindModuleID(enum.ContainerID)
		modName := h.GetModuleName(modID)
		if enum.Name == name.Name && (name.Module == "" || modName == name.Module) {
			// Output JavaDoc documentation if present
			if enum.Documentation != "" {
				fmt.Fprintf(e.output, "/**\n * %s\n */\n", enum.Documentation)
			}

			fmt.Fprintf(e.output, "CREATE OR MODIFY ENUMERATION %s.%s (\n", modName, enum.Name)
			for i, v := range enum.Values {
				comma := ","
				if i == len(enum.Values)-1 {
					comma = ""
				}
				caption := ""
				if v.Caption != nil {
					caption = v.Caption.GetTranslation("en_US")
				}
				fmt.Fprintf(e.output, "  %s '%s'%s\n", v.Name, caption, comma)
			}
			fmt.Fprintln(e.output, ");")
			fmt.Fprintln(e.output, "/")
			return nil
		}
	}

	return mdlerrors.NewNotFound("enumeration", name.String())
}

// mendixReservedWords contains words that cannot be used as enumeration value names.
// These are Java reserved words plus Mendix-specific reserved identifiers.
// Using any of these triggers CE7247: "The name 'X' is a reserved word."
var mendixReservedWords = map[string]bool{
	// Java reserved words
	"abstract": true, "assert": true, "boolean": true, "break": true,
	"byte": true, "case": true, "catch": true, "char": true,
	"class": true, "const": true, "continue": true, "default": true,
	"do": true, "double": true, "else": true, "enum": true,
	"extends": true, "false": true, "final": true, "finally": true,
	"float": true, "for": true, "goto": true, "if": true,
	"implements": true, "import": true, "instanceof": true, "int": true,
	"interface": true, "long": true, "native": true, "new": true,
	"null": true, "package": true, "private": true, "protected": true,
	"public": true, "return": true, "short": true, "static": true,
	"strictfp": true, "super": true, "switch": true, "synchronized": true,
	"this": true, "throw": true, "throws": true, "transient": true,
	"true": true, "try": true, "void": true, "volatile": true,
	"while": true,
	// Mendix-specific reserved identifiers
	"changedby": true, "changeddate": true, "con": true, "context": true,
	"createddate": true, "currentuser": true, "guid": true,
	"id": true, "mendixobject": true, "submetaobjectname": true,
}

// ValidateEnumeration checks enumeration value names for reserved words.
// Returns a list of structured violations with rule IDs (CE7247 equivalent).
// This function does not require a project connection.
func ValidateEnumeration(stmt *ast.CreateEnumerationStmt) []linter.Violation {
	var violations []linter.Violation
	for _, v := range stmt.Values {
		if mendixReservedWords[strings.ToLower(v.Name)] {
			violations = append(violations, linter.Violation{
				RuleID:   "MDL010",
				Severity: linter.SeverityError,
				Message: fmt.Sprintf(
					"enumeration value '%s' is a reserved word (CE7247)",
					v.Name),
				Location: linter.Location{
					DocumentType: "enumeration",
					DocumentName: stmt.Name.String(),
				},
				Suggestion: fmt.Sprintf("Rename to a non-reserved name (e.g., '%s_' or 'Is%s')", v.Name, v.Name),
			})
		}
	}
	return violations
}

// mendixSystemAttributeNames are attribute names reserved by the Mendix runtime.
// These are auto-managed system attributes on persistent entities and cannot be
// used as user-defined attribute names.
var mendixSystemAttributeNames = map[string]bool{
	"createddate": true,
	"changeddate": true,
	"owner":       true,
	"changedby":   true,
}

// ValidateEntity checks entity attribute names for reserved system names.
// Returns a list of structured violations with rule IDs. This function does not require a project connection.
func ValidateEntity(stmt *ast.CreateEntityStmt) []linter.Violation {
	var violations []linter.Violation
	// Only persistent entities have system attributes
	if stmt.Kind != ast.EntityPersistent {
		return violations
	}
	for _, attr := range stmt.Attributes {
		// Skip pseudo-types — these ARE the system attributes
		if attr.Type.Kind == ast.TypeAutoOwner || attr.Type.Kind == ast.TypeAutoChangedBy ||
			attr.Type.Kind == ast.TypeAutoCreatedDate || attr.Type.Kind == ast.TypeAutoChangedDate {
			continue
		}
		if mendixSystemAttributeNames[strings.ToLower(attr.Name)] {
			violations = append(violations, linter.Violation{
				RuleID:   "MDL020",
				Severity: linter.SeverityError,
				Message: fmt.Sprintf(
					"attribute '%s' conflicts with a Mendix system attribute name. "+
						"Mendix automatically manages '%s' on persistent entities",
					attr.Name, attr.Name),
				Location: linter.Location{
					DocumentType: "entity",
					DocumentName: stmt.Name.String(),
				},
				Suggestion: fmt.Sprintf("Rename to avoid conflicts (e.g., 'Custom%s')", attr.Name),
			})
		}
	}
	return violations
}
