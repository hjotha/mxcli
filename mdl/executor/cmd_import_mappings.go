// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// showImportMappings prints a table of all import mapping documents.
func (e *Executor) showImportMappings(inModule string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	all, err := e.reader.ListImportMappings()
	if err != nil {
		return fmt.Errorf("failed to list import mappings: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	type row struct {
		qualifiedName, name, schemaSource string
		elementCount                      int
	}
	var rows []row
	qnWidth, nameWidth, srcWidth := len("Import Mapping"), len("Name"), len("Schema Source")

	for _, im := range all {
		modID := h.FindModuleID(im.ContainerID)
		moduleName := h.GetModuleName(modID)
		if inModule != "" && !strings.EqualFold(moduleName, inModule) {
			continue
		}
		qn := moduleName + "." + im.Name
		src := im.JsonStructure
		if src == "" {
			src = im.XmlSchema
		}
		if src == "" {
			src = im.MessageDefinition
		}
		if src == "" {
			src = "(none)"
		}
		r := row{qualifiedName: qn, name: im.Name, schemaSource: src, elementCount: len(im.Elements)}
		if len(qn) > qnWidth {
			qnWidth = len(qn)
		}
		if len(im.Name) > nameWidth {
			nameWidth = len(im.Name)
		}
		if len(src) > srcWidth {
			srcWidth = len(src)
		}
		rows = append(rows, r)
	}

	if len(rows) == 0 {
		if inModule != "" {
			fmt.Fprintf(e.output, "No import mappings found in module %s\n", inModule)
		} else {
			fmt.Fprintln(e.output, "No import mappings found")
		}
		return nil
	}

	// Sort alphabetically by qualified name
	sort.Slice(rows, func(i, j int) bool { return rows[i].qualifiedName < rows[j].qualifiedName })

	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %s |\n",
		qnWidth, "Import Mapping", nameWidth, "Name", srcWidth, "Schema Source", "Elements")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|----------|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", nameWidth), strings.Repeat("-", srcWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %8d |\n",
			qnWidth, r.qualifiedName, nameWidth, r.name, srcWidth, r.schemaSource, r.elementCount)
	}
	return nil
}

// describeImportMapping prints the MDL representation of an import mapping.
func (e *Executor) describeImportMapping(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	im, err := e.reader.GetImportMappingByQualifiedName(name.Module, name.Name)
	if err != nil {
		return fmt.Errorf("import mapping %s not found", name)
	}

	if im.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", strings.ReplaceAll(im.Documentation, "\n", "\n * "))
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(im.ContainerID)
	moduleName := h.GetModuleName(modID)

	fmt.Fprintf(e.output, "CREATE IMPORT MAPPING %s.%s\n", moduleName, im.Name)

	if im.JsonStructure != "" {
		fmt.Fprintf(e.output, "  WITH JSON STRUCTURE %s\n", im.JsonStructure)
	} else if im.XmlSchema != "" {
		fmt.Fprintf(e.output, "  WITH XML SCHEMA %s\n", im.XmlSchema)
	}

	if len(im.Elements) > 0 {
		fmt.Fprintln(e.output, "{")
		for _, elem := range im.Elements {
			printImportMappingElement(e, elem, 1, true)
			fmt.Fprintln(e.output)
		}
		fmt.Fprintln(e.output, "};")
	}
	return nil
}

// handlingKeyword returns the MDL keyword for a Mendix ObjectHandling value.
func handlingKeyword(handling string) string {
	switch handling {
	case "Find":
		return "FIND"
	case "FindOrCreate":
		return "FIND OR CREATE"
	default:
		return "CREATE"
	}
}

func printImportMappingElement(e *Executor, elem *model.ImportMappingElement, depth int, isRoot bool) {
	indent := strings.Repeat("  ", depth)
	if elem.Kind == "Object" {
		handling := handlingKeyword(elem.ObjectHandling)
		if isRoot {
			// Root: CREATE Module.Entity { — use "." if entity is empty
			entity := elem.Entity
			if entity == "" {
				entity = "."
			}
			fmt.Fprintf(e.output, "%s%s %s {\n", indent, handling, entity)
		} else {
			// Nested object element:
			//   CREATE Assoc/Entity = jsonKey   — normal association path
			//   CREATE ./Entity = jsonKey       — self-reference (no association)
			//   CREATE . = jsonKey              — structural grouping (no association, no entity)
			assoc := elem.Association
			entity := elem.Entity
			if assoc == "" && entity == "" {
				fmt.Fprintf(e.output, "%s%s . = %s", indent, handling, elem.ExposedName)
			} else if assoc == "" {
				fmt.Fprintf(e.output, "%s%s ./%s = %s", indent, handling, entity, elem.ExposedName)
			} else {
				fmt.Fprintf(e.output, "%s%s %s/%s = %s", indent, handling, assoc, entity, elem.ExposedName)
			}
			if len(elem.Children) > 0 {
				fmt.Fprintln(e.output, " {")
			}
		}
		if len(elem.Children) > 0 {
			for i, child := range elem.Children {
				printImportMappingElement(e, child, depth+1, false)
				if i < len(elem.Children)-1 {
					fmt.Fprintln(e.output, ",")
				} else {
					fmt.Fprintln(e.output)
				}
			}
			fmt.Fprintf(e.output, "%s}", indent)
		}
	} else {
		// Value mapping: Attr = jsonField KEY
		attrName := elem.Attribute
		// Strip module prefix if present (Module.Entity.Attr → Attr)
		if parts := strings.Split(attrName, "."); len(parts) == 3 {
			attrName = parts[2]
		}
		keyStr := ""
		if elem.IsKey {
			keyStr = " KEY"
		}
		fmt.Fprintf(e.output, "%s%s = %s%s", indent, attrName, elem.ExposedName, keyStr)
	}
}

// execCreateImportMapping creates a new import mapping.
func (e *Executor) execCreateImportMapping(s *ast.CreateImportMappingStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return fmt.Errorf("module %s not found", s.Name.Module)
	}
	containerID := module.ID

	im := &model.ImportMapping{
		ContainerID: containerID,
		Name:        s.Name.Name,
		ExportLevel: "Hidden",
	}

	// Set schema source reference
	switch s.SchemaKind {
	case "JSON_STRUCTURE":
		im.JsonStructure = s.SchemaRef.String()
	case "XML_SCHEMA":
		im.XmlSchema = s.SchemaRef.String()
	}

	// Build path→element info map from JSON structure for schema alignment
	jsElements := map[string]*jsonElementInfo{}
	if s.SchemaKind == "JSON_STRUCTURE" && s.SchemaRef.Module != "" {
		if js, err2 := e.reader.GetJsonStructureByQualifiedName(s.SchemaRef.Module, s.SchemaRef.Name); err2 == nil {
			buildJsonPathElementMap(js.Elements, jsElements)
		}
	}

	// Build element tree from the AST definition
	if s.RootElement != nil {
		root := buildImportMappingElementModel(s.Name.Module, s.RootElement, "", "(Object)", e.reader, jsElements, true)
		im.Elements = append(im.Elements, root)
	}

	if err := e.writer.CreateImportMapping(im); err != nil {
		return fmt.Errorf("failed to create import mapping: %w", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created import mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}

// buildImportMappingElementModel converts an AST element definition to a model element,
// resolving attribute qualified names, data types, and JSON structure alignment.
func buildImportMappingElementModel(moduleName string, def *ast.ImportMappingElementDef, parentEntity, parentPath string, reader *mpr.Reader, jsElements map[string]*jsonElementInfo, isRoot bool) *model.ImportMappingElement {
	elem := &model.ImportMappingElement{
		BaseElement: model.BaseElement{
			ID:       model.ID(mpr.GenerateID()),
			TypeName: "ImportMappings$ObjectMappingElement",
		},
		ExposedName: def.JsonName,
	}

	if def.Entity != "" {
		// Object mapping
		elem.Kind = "Object"
		entity := def.Entity
		if !strings.Contains(entity, ".") {
			entity = moduleName + "." + entity
		}
		elem.Entity = entity
		elem.ObjectHandling = def.ObjectHandling
		if elem.ObjectHandling == "" {
			elem.ObjectHandling = "Create"
		}
		if def.Association != "" {
			assoc := def.Association
			if !strings.Contains(assoc, ".") {
				assoc = moduleName + "." + assoc
			}
			elem.Association = assoc
		}

		// Compute JsonPath and align ExposedName/ElementType with JSON structure
		var jsonPath string
		if isRoot {
			jsonPath = "(Object)"
		} else {
			jsonPath = parentPath + "|" + def.JsonName
		}
		elem.JsonPath = jsonPath

		// Look up JSON structure element to align ExposedName and ElementType
		if info, ok := jsElements[jsonPath]; ok {
			elem.ExposedName = info.ExposedName
			if info.ElementType == "Array" {
				elem.Kind = "Array"
				// Children of array containers use the array item path
				jsonPath = jsonPath + "|(Object)"
			}
		}

		for _, child := range def.Children {
			elem.Children = append(elem.Children, buildImportMappingElementModel(moduleName, child, entity, jsonPath, reader, jsElements, false))
		}
	} else {
		// Value mapping — qualify attribute name as Module.Entity.Attribute
		elem.Kind = "Value"
		elem.TypeName = "ImportMappings$ValueMappingElement"
		elem.DataType = resolveAttributeType(parentEntity, def.Attribute, reader)
		elem.IsKey = def.IsKey
		attr := def.Attribute
		if parentEntity != "" && !strings.Contains(attr, ".") {
			attr = parentEntity + "." + attr
		}
		elem.Attribute = attr

		// Compute JsonPath and align ExposedName with JSON structure
		jsonPath := parentPath + "|" + def.JsonName
		elem.JsonPath = jsonPath
		if info, ok := jsElements[jsonPath]; ok {
			elem.ExposedName = info.ExposedName
		}
	}

	return elem
}

// resolveAttributeType looks up the data type of an entity attribute from the project.
// Returns "String" as default if the attribute cannot be found.
func resolveAttributeType(entityQN, attrName string, reader *mpr.Reader) string {
	if reader == nil || entityQN == "" {
		return "String"
	}
	parts := strings.SplitN(entityQN, ".", 2)
	if len(parts) != 2 {
		return "String"
	}
	dms, err := reader.ListDomainModels()
	if err != nil {
		return "String"
	}
	for _, dm := range dms {
		for _, e := range dm.Entities {
			if e.Name == parts[1] {
				for _, a := range e.Attributes {
					if a.Name == attrName && a.Type != nil {
						return a.Type.GetTypeName()
					}
				}
			}
		}
	}
	return "String"
}

// execDropImportMapping deletes an import mapping.
func (e *Executor) execDropImportMapping(s *ast.DropImportMappingStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	im, err := e.reader.GetImportMappingByQualifiedName(s.Name.Module, s.Name.Name)
	if err != nil {
		return fmt.Errorf("import mapping %s not found", s.Name)
	}

	if err := e.writer.DeleteImportMapping(im.ID); err != nil {
		return fmt.Errorf("failed to drop import mapping: %w", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Dropped import mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}
