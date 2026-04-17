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

// showExportMappings prints a table of all export mapping documents.
func (e *Executor) showExportMappings(inModule string) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	all, err := e.reader.ListExportMappings()
	if err != nil {
		return mdlerrors.NewBackend("list export mappings", err)
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
		rows = append(rows, row{qualifiedName: qn, name: em.Name, schemaSource: src, elementCount: len(em.Elements)})
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

	result := &TableResult{
		Columns: []string{"Export Mapping", "Name", "Schema Source", "Elements"},
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.name, r.schemaSource, r.elementCount})
	}
	return e.writeResult(result)
}

// describeExportMapping prints the MDL representation of an export mapping.
func (e *Executor) describeExportMapping(name ast.QualifiedName) error {
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	em, err := e.reader.GetExportMappingByQualifiedName(name.Module, name.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return mdlerrors.NewNotFound("export mapping", name.String())
		}
		return mdlerrors.NewBackend("get export mapping", err)
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
		return mdlerrors.NewNotConnectedWrite()
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return mdlerrors.NewNotFound("module", s.Name.Module)
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
	jsElems := map[string]*mpr.JsonElement{}
	if s.SchemaKind == "JSON_STRUCTURE" && s.SchemaRef.Module != "" {
		if js, err2 := e.reader.GetJsonStructureByQualifiedName(s.SchemaRef.Module, s.SchemaRef.Name); err2 == nil {
			buildJsonElementPathMap(js.Elements, jsElems)
		}
	}

	// Build element tree from the AST definition, cloning JSON structure properties
	if s.RootElement != nil {
		root := buildExportMappingElementModel(s.Name.Module, s.RootElement, "", "(Object)", jsElems, e.reader, true)
		em.Elements = append(em.Elements, root)
	}

	if err := e.writer.CreateExportMapping(em); err != nil {
		return mdlerrors.NewBackend("create export mapping", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created export mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}

// buildExportMappingElementModel converts an AST element definition to a model element.
// It clones properties from the matching JSON structure element and adds mapping bindings.
func buildExportMappingElementModel(moduleName string, def *ast.ExportMappingElementDef, parentEntity, parentPath string, jsElems map[string]*mpr.JsonElement, reader *mpr.Reader, isRoot bool) *model.ExportMappingElement {
	elem := &model.ExportMappingElement{
		BaseElement: model.BaseElement{
			ID: model.ID(mpr.GenerateID()),
		},
	}

	// Determine lookup path
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
		elem.MaxOccurs = jsElem.MaxOccurs
	} else {
		elem.ExposedName = def.JsonName
		elem.JsonPath = lookupPath
	}

	if def.Entity != "" {
		// Object/Array mapping — bind to entity
		elem.Kind = "Object"
		elem.TypeName = "ExportMappings$ObjectMappingElement"

		entity := def.Entity
		if !strings.Contains(entity, ".") {
			entity = moduleName + "." + entity
		}

		assoc := def.Association
		if assoc != "" && !strings.Contains(assoc, ".") {
			assoc = moduleName + "." + assoc
		}

		handling := "Parameter"
		if !isRoot {
			handling = "Find"
		}

		// Check if this is an array element in the JSON structure
		if jsElem, ok := jsElems[lookupPath]; ok && jsElem.ElementType == "Array" {
			// Export arrays have two levels:
			// 1. Array container: Kind=Array, entity=container entity, assoc to parent
			// 2. Item object: Kind=Object, entity=item entity, assoc to container
			//
			// MDL syntax: Assoc/Entity AS items { ItemAssoc/ItemEntity AS ItemsItem { values } }
			// The outer Assoc/Entity is for the container, the nested child provides the item.
			elem.Kind = "Array"
			elem.Association = assoc
			elem.ObjectHandling = handling
			elem.Entity = entity

			itemPath := lookupPath + "|(Object)"

			// The first (and typically only) child of the array in the MDL is the item definition.
			// Its children become the item element's value children.
			if len(def.Children) == 1 && def.Children[0].Entity != "" {
				itemDef := def.Children[0]
				itemEntity := itemDef.Entity
				if !strings.Contains(itemEntity, ".") {
					itemEntity = moduleName + "." + itemEntity
				}
				itemAssoc := itemDef.Association
				if itemAssoc != "" && !strings.Contains(itemAssoc, ".") {
					itemAssoc = moduleName + "." + itemAssoc
				}

				itemElem := &model.ExportMappingElement{
					BaseElement: model.BaseElement{
						ID:       model.ID(mpr.GenerateID()),
						TypeName: "ExportMappings$ObjectMappingElement",
					},
					Kind:           "Object",
					Entity:         itemEntity,
					Association:    itemAssoc,
					ObjectHandling: "Find",
				}
				if jsItem, ok2 := jsElems[itemPath]; ok2 {
					itemElem.ExposedName = jsItem.ExposedName
					itemElem.JsonPath = jsItem.Path
					itemElem.MaxOccurs = jsItem.MaxOccurs
				} else {
					itemElem.ExposedName = elem.ExposedName + "Item"
					itemElem.JsonPath = itemPath
					itemElem.MaxOccurs = -1
				}
				// Item's children are the value elements
				for _, valChild := range itemDef.Children {
					itemElem.Children = append(itemElem.Children, buildExportMappingElementModel(moduleName, valChild, itemEntity, itemPath, jsElems, reader, false))
				}
				elem.Children = append(elem.Children, itemElem)
			} else {
				// Fallback: treat children as direct item children (no intermediate entity)
				for _, child := range def.Children {
					elem.Children = append(elem.Children, buildExportMappingElementModel(moduleName, child, entity, itemPath, jsElems, reader, false))
				}
			}
		} else {
			// Regular object element
			elem.Entity = entity
			elem.Association = assoc
			elem.ObjectHandling = handling
			for _, child := range def.Children {
				elem.Children = append(elem.Children, buildExportMappingElementModel(moduleName, child, entity, lookupPath, jsElems, reader, false))
			}
		}
	} else {
		// Value mapping — bind to attribute
		elem.Kind = "Value"
		elem.TypeName = "ExportMappings$ValueMappingElement"
		elem.DataType = resolveAttributeType(parentEntity, def.Attribute, reader)
		attr := def.Attribute
		if parentEntity != "" && !strings.Contains(attr, ".") {
			attr = parentEntity + "." + attr
		}
		elem.Attribute = attr
		// JsonPath already set from JSON structure clone above
	}

	return elem
}

// execDropExportMapping deletes an export mapping.
func (e *Executor) execDropExportMapping(s *ast.DropExportMappingStmt) error {
	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	em, err := e.reader.GetExportMappingByQualifiedName(s.Name.Module, s.Name.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return mdlerrors.NewNotFound("export mapping", s.Name.String())
		}
		return mdlerrors.NewBackend("get export mapping", err)
	}

	if err := e.writer.DeleteExportMapping(em.ID); err != nil {
		return mdlerrors.NewBackend("drop export mapping", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Dropped export mapping %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}
