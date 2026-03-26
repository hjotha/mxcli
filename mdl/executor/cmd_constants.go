// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
)

// showConstants handles SHOW CONSTANTS command.
func (e *Executor) showConstants(moduleName string) error {
	constants, err := e.reader.ListConstants()
	if err != nil {
		return fmt.Errorf("failed to list constants: %w", err)
	}

	// Use hierarchy for proper module resolution (handles constants inside folders)
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		typeStr       string
		defaultStr    string
		exposed       string
	}
	var rows []row
	qnWidth := len("Qualified Name")
	modWidth := len("Module")
	nameWidth := len("Name")
	pathWidth := len("Folder")
	typeWidth := len("Type")
	defaultWidth := len("Default")
	exposedWidth := len("Exposed")

	for _, c := range constants {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && !strings.EqualFold(modName, moduleName) {
			continue
		}

		qualifiedName := modName + "." + c.Name
		folderPath := h.BuildFolderPath(c.ContainerID)
		typeStr := formatConstantType(c.Type)
		defaultStr := c.DefaultValue
		if len(defaultStr) > 40 {
			defaultStr = defaultStr[:37] + "..."
		}
		exposed := "No"
		if c.ExposedToClient {
			exposed = "Yes"
		}

		rows = append(rows, row{qualifiedName, modName, c.Name, folderPath, typeStr, defaultStr, exposed})
		if len(qualifiedName) > qnWidth {
			qnWidth = len(qualifiedName)
		}
		if len(modName) > modWidth {
			modWidth = len(modName)
		}
		if len(c.Name) > nameWidth {
			nameWidth = len(c.Name)
		}
		if len(folderPath) > pathWidth {
			pathWidth = len(folderPath)
		}
		if len(typeStr) > typeWidth {
			typeWidth = len(typeStr)
		}
		if len(defaultStr) > defaultWidth {
			defaultWidth = len(defaultStr)
		}
		if len(exposed) > exposedWidth {
			exposedWidth = len(exposed)
		}
	}

	if len(rows) == 0 {
		fmt.Fprintln(e.output, "No constants found.")
		return nil
	}

	// Sort by qualified name (module.name)
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Markdown table with aligned columns
	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		qnWidth, "Qualified Name", modWidth, "Module", nameWidth, "Name",
		pathWidth, "Folder", typeWidth, "Type", defaultWidth, "Default", exposedWidth, "Exposed")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", modWidth), strings.Repeat("-", nameWidth),
		strings.Repeat("-", pathWidth), strings.Repeat("-", typeWidth),
		strings.Repeat("-", defaultWidth), strings.Repeat("-", exposedWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
			qnWidth, r.qualifiedName, modWidth, r.module, nameWidth, r.name,
			pathWidth, r.folderPath, typeWidth, r.typeStr, defaultWidth, r.defaultStr, exposedWidth, r.exposed)
	}
	fmt.Fprintf(e.output, "\n(%d constants)\n", len(rows))

	return nil
}

// describeConstant handles DESCRIBE CONSTANT command.
func (e *Executor) describeConstant(name ast.QualifiedName) error {
	constants, err := e.reader.ListConstants()
	if err != nil {
		return fmt.Errorf("failed to list constants: %w", err)
	}

	// Use hierarchy for proper module resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Find the constant
	for _, c := range constants {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		if strings.EqualFold(modName, name.Module) && strings.EqualFold(c.Name, name.Name) {
			return e.outputConstantMDL(c, modName)
		}
	}

	return fmt.Errorf("constant not found: %s", name)
}

// outputConstantMDL outputs a constant definition in MDL format.
func (e *Executor) outputConstantMDL(c *model.Constant, moduleName string) error {
	// Format default value based on type
	defaultValueStr := formatDefaultValue(c.Type, c.DefaultValue)

	fmt.Fprintf(e.output, "CREATE CONSTANT %s.%s\n", moduleName, c.Name)
	fmt.Fprintf(e.output, "  TYPE %s\n", formatConstantTypeForMDL(c.Type))
	fmt.Fprintf(e.output, "  DEFAULT %s", defaultValueStr)

	// Add folder if present
	h, _ := e.getHierarchy()
	if h != nil {
		if folderPath := h.BuildFolderPath(c.ContainerID); folderPath != "" {
			fmt.Fprintf(e.output, "\n  FOLDER '%s'", folderPath)
		}
	}

	// Add options if present
	if c.Documentation != "" {
		escaped := strings.ReplaceAll(c.Documentation, "'", "''")
		fmt.Fprintf(e.output, "\n  COMMENT '%s'", escaped)
	}
	if c.ExposedToClient {
		fmt.Fprintf(e.output, "\n  EXPOSED TO CLIENT")
	}

	fmt.Fprintln(e.output, ";")
	fmt.Fprintln(e.output, "/")

	return nil
}

// formatConstantType returns a human-readable type string.
func formatConstantType(dt model.ConstantDataType) string {
	switch dt.Kind {
	case "String":
		return "String"
	case "Integer":
		return "Integer"
	case "Long":
		return "Long"
	case "Decimal":
		return "Decimal"
	case "Boolean":
		return "Boolean"
	case "DateTime":
		return "DateTime"
	case "Date":
		return "Date"
	case "Binary":
		return "Binary"
	case "Enumeration":
		if dt.EnumRef != "" {
			return fmt.Sprintf("Enumeration(%s)", dt.EnumRef)
		}
		return "Enumeration"
	case "Object":
		if dt.EntityRef != "" {
			return dt.EntityRef
		}
		return "Object"
	case "List":
		if dt.EntityRef != "" {
			return fmt.Sprintf("List of %s", dt.EntityRef)
		}
		return "List"
	default:
		if dt.Kind == "" {
			return "Unknown"
		}
		return dt.Kind
	}
}

// formatConstantTypeForMDL returns the MDL syntax for the type.
func formatConstantTypeForMDL(dt model.ConstantDataType) string {
	switch dt.Kind {
	case "String":
		return "String"
	case "Integer":
		return "Integer"
	case "Long":
		return "Long"
	case "Decimal":
		return "Decimal"
	case "Boolean":
		return "Boolean"
	case "DateTime":
		return "DateTime"
	case "Date":
		return "Date"
	case "Binary":
		return "Binary"
	case "Enumeration":
		if dt.EnumRef != "" {
			return fmt.Sprintf("Enumeration(%s)", dt.EnumRef)
		}
		return "Enumeration"
	default:
		if dt.Kind == "" {
			return "String"
		}
		return dt.Kind
	}
}

// formatDefaultValue formats the default value for MDL output.
func formatDefaultValue(dt model.ConstantDataType, value string) string {
	if value == "" {
		switch dt.Kind {
		case "String":
			return "''"
		case "Boolean":
			return "false"
		case "Integer", "Long", "Decimal":
			return "0"
		default:
			return "''"
		}
	}

	switch dt.Kind {
	case "String":
		// Quote the string value, escaping single quotes
		escaped := strings.ReplaceAll(value, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case "Boolean":
		if strings.EqualFold(value, "true") {
			return "true"
		}
		return "false"
	case "Integer", "Long", "Decimal":
		return value
	case "Enumeration":
		// Enumeration values are stored as qualified names
		return value
	default:
		escaped := strings.ReplaceAll(value, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// createConstant handles CREATE CONSTANT command.
func (e *Executor) createConstant(stmt *ast.CreateConstantStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected in write mode")
	}

	// Validate module name is specified
	if stmt.Name.Module == "" {
		return fmt.Errorf("module name required for constant: use CREATE CONSTANT Module.ConstantName")
	}

	// Find or auto-create module
	module, err := e.findOrCreateModule(stmt.Name.Module)
	if err != nil {
		return err
	}

	// Convert AST data type to model constant data type
	constType := astDataTypeToConstantDataType(stmt.DataType)

	// Format default value as string
	defaultValue := ""
	if stmt.DefaultValue != nil {
		defaultValue = fmt.Sprintf("%v", stmt.DefaultValue)
	}

	// Check if constant already exists in this module
	existingConstants, err := e.reader.ListConstants()
	if err == nil {
		h, _ := e.getHierarchy()
		for _, c := range existingConstants {
			modID := h.FindModuleID(c.ContainerID)
			modName := h.GetModuleName(modID)
			if strings.EqualFold(modName, stmt.Name.Module) && strings.EqualFold(c.Name, stmt.Name.Name) {
				if stmt.CreateOrModify {
					// Update existing constant — COMMENT takes precedence over doc-comment
					if stmt.Comment != "" {
						c.Documentation = stmt.Comment
					} else {
						c.Documentation = stmt.Documentation
					}
					c.Type = constType
					c.DefaultValue = defaultValue
					c.ExposedToClient = stmt.ExposedToClient
					if err := e.writer.UpdateConstant(c); err != nil {
						return fmt.Errorf("failed to update constant: %w", err)
					}
					e.invalidateHierarchy()
					fmt.Fprintf(e.output, "Modified constant: %s.%s\n", modName, c.Name)
					return nil
				}
				return fmt.Errorf("constant already exists: %s.%s (use CREATE OR MODIFY to update)", modName, c.Name)
			}
		}
	}

	// COMMENT 'text' takes precedence; fall back to /** */ doc-comment
	doc := stmt.Comment
	if doc == "" {
		doc = stmt.Documentation
	}

	containerID := module.ID
	if stmt.Folder != "" {
		folderID, err := e.resolveFolder(module.ID, stmt.Folder)
		if err != nil {
			return fmt.Errorf("failed to resolve folder %s: %w", stmt.Folder, err)
		}
		containerID = folderID
	}

	constant := &model.Constant{
		ContainerID:     containerID,
		Name:            stmt.Name.Name,
		Documentation:   doc,
		Type:            constType,
		DefaultValue:    defaultValue,
		ExposedToClient: stmt.ExposedToClient,
	}

	if err := e.writer.CreateConstant(constant); err != nil {
		return fmt.Errorf("failed to create constant: %w", err)
	}
	e.invalidateHierarchy()
	fmt.Fprintf(e.output, "Created constant: %s.%s\n", stmt.Name.Module, stmt.Name.Name)
	return nil
}

// dropConstant handles DROP CONSTANT command.
func (e *Executor) dropConstant(stmt *ast.DropConstantStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected in write mode")
	}

	constants, err := e.reader.ListConstants()
	if err != nil {
		return fmt.Errorf("failed to list constants: %w", err)
	}

	// Use hierarchy for proper module resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Find the constant
	for _, c := range constants {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		if strings.EqualFold(modName, stmt.Name.Module) && strings.EqualFold(c.Name, stmt.Name.Name) {
			if err := e.writer.DeleteConstant(c.ID); err != nil {
				return fmt.Errorf("failed to drop constant: %w", err)
			}
			e.invalidateHierarchy()
			fmt.Fprintf(e.output, "Dropped constant: %s.%s\n", modName, c.Name)
			return nil
		}
	}

	return fmt.Errorf("constant not found: %s", stmt.Name)
}

// astDataTypeToConstantDataType converts AST DataType to model.ConstantDataType.
func astDataTypeToConstantDataType(dt ast.DataType) model.ConstantDataType {
	switch dt.Kind {
	case ast.TypeString:
		return model.ConstantDataType{Kind: "String"}
	case ast.TypeInteger:
		return model.ConstantDataType{Kind: "Integer"}
	case ast.TypeLong:
		return model.ConstantDataType{Kind: "Long"}
	case ast.TypeDecimal:
		return model.ConstantDataType{Kind: "Decimal"}
	case ast.TypeBoolean:
		return model.ConstantDataType{Kind: "Boolean"}
	case ast.TypeDateTime:
		return model.ConstantDataType{Kind: "DateTime"}
	case ast.TypeDate:
		return model.ConstantDataType{Kind: "Date"}
	case ast.TypeBinary:
		return model.ConstantDataType{Kind: "Binary"}
	case ast.TypeEnumeration:
		enumRef := ""
		if dt.EnumRef != nil {
			enumRef = dt.EnumRef.String()
		}
		return model.ConstantDataType{Kind: "Enumeration", EnumRef: enumRef}
	default:
		return model.ConstantDataType{Kind: "String"}
	}
}
