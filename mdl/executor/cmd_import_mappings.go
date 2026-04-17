// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// showImportMappings prints a table of all import mapping documents.
func (e *Executor) showImportMappings(inModule string) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	all, err := e.reader.ListImportMappings()
	if err != nil {
		return mdlerrors.NewBackend("list import mappings", err)
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
		rows = append(rows, row{qualifiedName: qn, name: im.Name, schemaSource: src, elementCount: len(im.Elements)})
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

	result := &TableResult{
		Columns: []string{"Import Mapping", "Name", "Schema Source", "Elements"},
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.name, r.schemaSource, r.elementCount})
	}
	return e.writeResult(result)
}

// describeImportMapping prints the MDL representation of an import mapping.
func (e *Executor) describeImportMapping(name ast.QualifiedName) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	im, err := e.reader.GetImportMappingByQualifiedName(name.Module, name.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return mdlerrors.NewNotFound("import mapping", name.String())
		}
		return mdlerrors.NewBackend("get import mapping", err)
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
		return mdlerrors.NewNotConnectedWrite()
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return mdlerrors.NewNotFound("module", s.Name.Module)
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

	// Build path→JsonElement map from JSON structure — mapping elements clone from this
	jsElementsByPath := map[string]*mpr.JsonElement{}
	if s.SchemaKind == "JSON_STRUCTURE" && s.SchemaRef.Module != "" {
		if js, err2 := e.reader.GetJsonStructureByQualifiedName(s.SchemaRef.Module, s.SchemaRef.Name); err2 == nil {
			buildJsonElementPathMap(js.Elements, jsElementsByPath)
		}
	}

	// Build element tree from the AST definition, cloning JSON structure properties
	if s.RootElement != nil {
		root := buildImportMappingElementModel(s.Name.Module, s.RootElement, "", "(Object)", e.reader, jsElementsByPath, true)
		im.Elements = append(im.Elements, root)
	}

	if err := e.writer.CreateImportMapping(im); err != nil {
		return mdlerrors.NewBackend("create import mapping", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created import mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}

// buildImportMappingElementModel converts an AST element definition to a model element.
// It clones properties from the matching JSON structure element (ExposedName, JsonPath,
// MaxOccurs, ElementType, etc.) and adds mapping-specific bindings (Entity, Attribute,
// Association, ObjectHandling).
func buildImportMappingElementModel(moduleName string, def *ast.ImportMappingElementDef, parentEntity, parentPath string, reader *mpr.Reader, jsElems map[string]*mpr.JsonElement, isRoot bool) *model.ImportMappingElement {
	elem := &model.ImportMappingElement{
		BaseElement: model.BaseElement{
			ID: model.ID(mpr.GenerateID()),
		},
	}

	// Determine lookup path in JSON structure
	var lookupPath string
	if isRoot {
		lookupPath = "(Object)"
	} else {
		lookupPath = parentPath + "|" + def.JsonName
	}

	// Clone properties from the matching JSON structure element
	if jsElem, ok := jsElems[lookupPath]; ok {
		elem.ExposedName = jsElem.ExposedName
		elem.JsonPath = jsElem.Path
		elem.MinOccurs = jsElem.MinOccurs
		elem.MaxOccurs = jsElem.MaxOccurs
		elem.Nillable = jsElem.Nillable
		elem.OriginalValue = jsElem.OriginalValue
		elem.FractionDigits = jsElem.FractionDigits
		elem.TotalDigits = jsElem.TotalDigits
		elem.MaxLength = jsElem.MaxLength
	} else {
		elem.ExposedName = def.JsonName
		elem.JsonPath = lookupPath
		elem.Nillable = true
		elem.FractionDigits = -1
		elem.TotalDigits = -1
	}

	if def.Entity != "" {
		// Object/Array mapping — bind to entity
		elem.Kind = "Object"
		elem.TypeName = "ImportMappings$ObjectMappingElement"

		entity := def.Entity
		if !strings.Contains(entity, ".") {
			entity = moduleName + "." + entity
		}

		assoc := def.Association
		if assoc != "" && !strings.Contains(assoc, ".") {
			assoc = moduleName + "." + assoc
		}

		handling := def.ObjectHandling
		if handling == "" {
			handling = "Create"
		}

		elem.Entity = entity
		elem.Association = assoc
		elem.ObjectHandling = handling

		// For arrays: skip the container, use the item path directly.
		// Studio Pro represents arrays as a single ObjectMappingElement at the |(Object) item path.
		childPath := lookupPath
		if jsElem, ok := jsElems[lookupPath]; ok && jsElem.ElementType == "Array" {
			itemPath := lookupPath + "|(Object)"
			if jsItem, ok2 := jsElems[itemPath]; ok2 {
				elem.ExposedName = jsItem.ExposedName
				elem.JsonPath = jsItem.Path
				elem.MinOccurs = jsItem.MinOccurs
				elem.MaxOccurs = jsItem.MaxOccurs
				elem.Nillable = jsItem.Nillable
			}
			childPath = itemPath
		}

		for _, child := range def.Children {
			elem.Children = append(elem.Children, buildImportMappingElementModel(moduleName, child, entity, childPath, reader, jsElems, false))
		}
	} else {
		// Value mapping — bind to attribute
		elem.Kind = "Value"
		elem.TypeName = "ImportMappings$ValueMappingElement"
		elem.DataType = resolveAttributeType(parentEntity, def.Attribute, reader)
		elem.IsKey = def.IsKey
		attr := def.Attribute
		if parentEntity != "" && !strings.Contains(attr, ".") {
			attr = parentEntity + "." + attr
		}
		elem.Attribute = attr
	}

	return elem
}

// buildJsonElementPathMap recursively builds a map from JSON path → JsonElement.
func buildJsonElementPathMap(elems []*mpr.JsonElement, m map[string]*mpr.JsonElement) {
	for _, e := range elems {
		if e == nil {
			continue
		}
		m[e.Path] = e
		buildJsonElementPathMap(e.Children, m)
	}
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
		return mdlerrors.NewNotConnectedWrite()
	}

	im, err := e.reader.GetImportMappingByQualifiedName(s.Name.Module, s.Name.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return mdlerrors.NewNotFound("import mapping", s.Name.String())
		}
		return mdlerrors.NewBackend("get import mapping", err)
	}

	if err := e.writer.DeleteImportMapping(im.ID); err != nil {
		return mdlerrors.NewBackend("drop import mapping", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Dropped import mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}
