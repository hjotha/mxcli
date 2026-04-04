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

// showExportMappings prints a table of all export mapping documents.
func (e *Executor) showExportMappings(inModule string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	all, err := e.reader.ListExportMappings()
	if err != nil {
		return fmt.Errorf("failed to list export mappings: %w", err)
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
	qnWidth, nameWidth, srcWidth := len("Export Mapping"), len("Name"), len("Schema Source")

	for _, em := range all {
		modID := h.FindModuleID(em.ContainerID)
		moduleName := h.GetModuleName(modID)
		if inModule != "" && !strings.EqualFold(moduleName, inModule) {
			continue
		}
		qn := moduleName + "." + em.Name
		src := em.JsonStructure
		if src == "" {
			src = em.XmlSchema
		}
		if src == "" {
			src = em.MessageDefinition
		}
		if src == "" {
			src = "(none)"
		}
		r := row{qualifiedName: qn, name: em.Name, schemaSource: src, elementCount: len(em.Elements)}
		if len(qn) > qnWidth {
			qnWidth = len(qn)
		}
		if len(em.Name) > nameWidth {
			nameWidth = len(em.Name)
		}
		if len(src) > srcWidth {
			srcWidth = len(src)
		}
		rows = append(rows, r)
	}

	if len(rows) == 0 {
		if inModule != "" {
			fmt.Fprintf(e.output, "No export mappings found in module %s\n", inModule)
		} else {
			fmt.Fprintln(e.output, "No export mappings found")
		}
		return nil
	}

	// Sort alphabetically by qualified name
	sort.Slice(rows, func(i, j int) bool { return rows[i].qualifiedName < rows[j].qualifiedName })

	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %s |\n",
		qnWidth, "Export Mapping", nameWidth, "Name", srcWidth, "Schema Source", "Elements")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|----------|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", nameWidth), strings.Repeat("-", srcWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %8d |\n",
			qnWidth, r.qualifiedName, nameWidth, r.name, srcWidth, r.schemaSource, r.elementCount)
	}
	return nil
}

// describeExportMapping prints the MDL representation of an export mapping.
func (e *Executor) describeExportMapping(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	em, err := e.reader.GetExportMappingByQualifiedName(name.Module, name.Name)
	if err != nil {
		return fmt.Errorf("export mapping %s not found", name)
	}

	if em.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", strings.ReplaceAll(em.Documentation, "\n", "\n * "))
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(em.ContainerID)
	moduleName := h.GetModuleName(modID)

	fmt.Fprintf(e.output, "CREATE EXPORT MAPPING %s.%s\n", moduleName, em.Name)

	if em.JsonStructure != "" {
		fmt.Fprintf(e.output, "  WITH JSON STRUCTURE %s\n", em.JsonStructure)
	} else if em.XmlSchema != "" {
		fmt.Fprintf(e.output, "  WITH XML SCHEMA %s\n", em.XmlSchema)
	}

	if em.NullValueOption != "" && em.NullValueOption != "LeaveOutElement" {
		fmt.Fprintf(e.output, "  NULL VALUES %s\n", em.NullValueOption)
	}

	if len(em.Elements) > 0 {
		fmt.Fprintln(e.output, "{")
		for _, elem := range em.Elements {
			printExportMappingElement(e, elem, 1, true)
			fmt.Fprintln(e.output)
		}
		fmt.Fprintln(e.output, "};")
	}
	return nil
}

func printExportMappingElement(e *Executor, elem *model.ExportMappingElement, depth int, isRoot bool) {
	indent := strings.Repeat("  ", depth)
	if elem.Kind == "Object" {
		if isRoot {
			// Root: Module.Entity { — use "." if entity is empty (parameter mapping)
			entity := elem.Entity
			if entity == "" {
				entity = "."
			}
			fmt.Fprintf(e.output, "%s%s {\n", indent, entity)
		} else {
			// Nested object element. Several cases:
			//   Assoc/Entity AS jsonKey  — normal association path
			//   ./Entity AS jsonKey      — self-reference (no association, entity set)
			//   . AS jsonKey             — structural grouping (no association, no entity)
			assoc := elem.Association
			entity := elem.Entity
			if assoc == "" && entity == "" {
				fmt.Fprintf(e.output, "%s. AS %s", indent, elem.ExposedName)
			} else if assoc == "" {
				fmt.Fprintf(e.output, "%s./%s AS %s", indent, entity, elem.ExposedName)
			} else {
				fmt.Fprintf(e.output, "%s%s/%s AS %s", indent, assoc, entity, elem.ExposedName)
			}
			if len(elem.Children) > 0 {
				fmt.Fprintln(e.output, " {")
			}
		}
		if len(elem.Children) > 0 {
			for i, child := range elem.Children {
				printExportMappingElement(e, child, depth+1, false)
				if i < len(elem.Children)-1 {
					fmt.Fprintln(e.output, ",")
				} else {
					fmt.Fprintln(e.output)
				}
			}
			fmt.Fprintf(e.output, "%s}", indent)
		}
	} else {
		// Value mapping: jsonField = Attr
		attrName := elem.Attribute
		// Strip module prefix if present (Module.Entity.Attr → Attr)
		if parts := strings.Split(attrName, "."); len(parts) == 3 {
			attrName = parts[2]
		}
		fmt.Fprintf(e.output, "%s%s = %s", indent, elem.ExposedName, attrName)
	}
}

// execCreateExportMapping creates a new export mapping.
func (e *Executor) execCreateExportMapping(s *ast.CreateExportMappingStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return fmt.Errorf("module %s not found", s.Name.Module)
	}
	containerID := module.ID

	em := &model.ExportMapping{
		ContainerID:     containerID,
		Name:            s.Name.Name,
		ExportLevel:     "Hidden",
		NullValueOption: s.NullValueOption,
	}
	if em.NullValueOption == "" {
		em.NullValueOption = "LeaveOutElement"
	}

	// Set schema source reference
	switch s.SchemaKind {
	case "JSON_STRUCTURE":
		em.JsonStructure = s.SchemaRef.String()
	case "XML_SCHEMA":
		em.XmlSchema = s.SchemaRef.String()
	}

	// Build a path→element info map from the JSON structure for schema alignment.
	jsElements := map[string]*jsonElementInfo{}
	if s.SchemaKind == "JSON_STRUCTURE" && s.SchemaRef.Module != "" {
		if js, err2 := e.reader.GetJsonStructureByQualifiedName(s.SchemaRef.Module, s.SchemaRef.Name); err2 == nil {
			buildJsonPathElementMap(js.Elements, jsElements)
		}
	}

	// Build element tree from the AST definition
	if s.RootElement != nil {
		root := buildExportMappingElementModel(s.Name.Module, s.RootElement, "", "(Object)", jsElements, e.reader, true)
		em.Elements = append(em.Elements, root)
	}

	if err := e.writer.CreateExportMapping(em); err != nil {
		return fmt.Errorf("failed to create export mapping: %w", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created export mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}

// jsonElementInfo holds metadata about a JSON structure element for mapping alignment.
type jsonElementInfo struct {
	ElementType string // "Object", "Array", "Value"
	ExposedName string
	MaxOccurs   int // 1 for Object, -1 for Array (unbounded)
}

// buildJsonPathElementMap recursively walks a JSON structure element tree and populates
// a map of JSON path → element info.
func buildJsonPathElementMap(elems []*mpr.JsonElement, m map[string]*jsonElementInfo) {
	for _, e := range elems {
		if e == nil {
			continue
		}
		m[e.Path] = &jsonElementInfo{
			ElementType: e.ElementType,
			ExposedName: e.ExposedName,
			MaxOccurs:   e.MaxOccurs,
		}
		buildJsonPathElementMap(e.Children, m)
	}
}

// buildExportMappingElementModel converts an AST element definition to a model element.
// parentEntity is the fully-qualified entity name of the enclosing object element (for
// qualifying attribute names). parentPath is the JSON path of the parent element.
// jsElements maps JSON structure paths to their element info.
// isRoot indicates whether this is the root element (gets ObjectHandling "Parameter").
func buildExportMappingElementModel(moduleName string, def *ast.ExportMappingElementDef, parentEntity, parentPath string, jsElements map[string]*jsonElementInfo, reader *mpr.Reader, isRoot bool) *model.ExportMappingElement {
	elem := &model.ExportMappingElement{
		BaseElement: model.BaseElement{
			ID:       model.ID(mpr.GenerateID()),
			TypeName: "ExportMappings$ObjectMappingElement",
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
		if def.Association != "" {
			assoc := def.Association
			if !strings.Contains(assoc, ".") {
				assoc = moduleName + "." + assoc
			}
			elem.Association = assoc
		}

		// Set ObjectHandling: root = "Parameter", children = "Find"
		if isRoot {
			elem.ObjectHandling = "Parameter"
		} else {
			elem.ObjectHandling = "Find"
		}

		// Compute JsonPath and align with JSON structure metadata
		var jsonPath string
		if isRoot {
			jsonPath = parentPath // "(Object)"
			elem.ExposedName = ""
			if info, ok := jsElements[jsonPath]; ok {
				elem.MaxOccurs = info.MaxOccurs
			}
			elem.JsonPath = jsonPath
		} else {
			// Look up by original JSON key, then use ExposedName for the mapping's JsonPath
			lookupPath := parentPath + "|" + def.JsonName
			if info, ok := jsElements[lookupPath]; ok {
				elem.ExposedName = info.ExposedName
				elem.MaxOccurs = info.MaxOccurs
				jsonPath = parentPath + "|" + info.ExposedName
				if info.ElementType == "Array" {
					jsonPath = lookupPath // array keeps original path
				}
			} else {
				jsonPath = lookupPath
			}
			elem.JsonPath = jsonPath
			// Array children use the item path
			if jsElements[lookupPath] != nil && jsElements[lookupPath].ElementType == "Array" {
				jsonPath = lookupPath + "|(Object)"
			}
		}

		for _, child := range def.Children {
			elem.Children = append(elem.Children, buildExportMappingElementModel(moduleName, child, entity, jsonPath, jsElements, reader, false))
		}
	} else {
		// Value mapping — qualify attribute name as Module.Entity.Attribute
		elem.Kind = "Value"
		elem.TypeName = "ExportMappings$ValueMappingElement"
		elem.DataType = resolveAttributeType(parentEntity, def.Attribute, reader)
		attr := def.Attribute
		if parentEntity != "" && !strings.Contains(attr, ".") {
			attr = parentEntity + "." + attr
		}
		elem.Attribute = attr
		elem.JsonPath = parentPath + "|" + def.JsonName
	}

	return elem
}

// execDropExportMapping deletes an export mapping.
func (e *Executor) execDropExportMapping(s *ast.DropExportMappingStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	em, err := e.reader.GetExportMappingByQualifiedName(s.Name.Module, s.Name.Name)
	if err != nil {
		return fmt.Errorf("export mapping %s not found", s.Name)
	}

	if err := e.writer.DeleteExportMapping(em.ID); err != nil {
		return fmt.Errorf("failed to drop export mapping: %w", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Dropped export mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}
