// SPDX-License-Identifier: Apache-2.0

// Package executor - Entity display and describe commands (SHOW/DESCRIBE ENTITY)
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

// showEntities handles SHOW ENTITIES command.
func (e *Executor) showEntities(moduleName string) error {
	// Build module ID -> name map (single query)
	modules, err := e.reader.ListModules()
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}
	moduleNames := make(map[model.ID]string)
	for _, m := range modules {
		moduleNames[m.ID] = m.Name
	}

	// Get all domain models in a single query (avoids O(n²) behavior)
	domainModels, err := e.reader.ListDomainModels()
	if err != nil {
		return fmt.Errorf("failed to list domain models: %w", err)
	}

	// Build entity ID -> association count map
	assocCounts := make(map[model.ID]int)
	for _, dm := range domainModels {
		for _, assoc := range dm.Associations {
			assocCounts[assoc.ParentID]++
			assocCounts[assoc.ChildID]++
		}
	}

	// Collect System entities referenced via generalizations
	systemEntities := make(map[string]bool)
	for _, dm := range domainModels {
		for _, entity := range dm.Entities {
			if entity.GeneralizationRef != "" && strings.HasPrefix(entity.GeneralizationRef, "System.") {
				systemEntities[entity.GeneralizationRef] = true
			}
		}
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName  string
		entityType     string
		generalization string
		attrs          int
		assocs         int
		validations    int
		indexes        int
		events         int
		accessRules    int
	}
	var rows []row
	qnWidth := len("Entity")
	typeWidth := len("Type")
	genWidth := len("Extends")

	// Add System entities first (if showing all or System module)
	if moduleName == "" || moduleName == "System" {
		for sysEntity := range systemEntities {
			r := row{
				qualifiedName: sysEntity,
				entityType:    "System",
				attrs:         -1, // Unknown - from runtime
				assocs:        -1,
				validations:   -1,
				indexes:       -1,
				events:        -1,
				accessRules:   -1,
			}
			rows = append(rows, r)
			if len(sysEntity) > qnWidth {
				qnWidth = len(sysEntity)
			}
		}
	}

	for _, dm := range domainModels {
		modName := moduleNames[dm.ContainerID]
		// Filter by module name if specified
		if moduleName != "" && modName != moduleName {
			continue
		}
		for _, entity := range dm.Entities {
			// Determine entity type based on Source field and Persistable flag
			entityType := "Persistent"
			if strings.Contains(entity.Source, "OqlView") {
				entityType = "View"
			} else if strings.Contains(entity.Source, "OData") || entity.RemoteSource != "" || entity.RemoteSourceDocument != "" {
				entityType = "External"
			} else if !entity.Persistable {
				entityType = "Non-Persistent"
			}

			qualifiedName := modName + "." + entity.Name
			r := row{
				qualifiedName:  qualifiedName,
				entityType:     entityType,
				generalization: entity.GeneralizationRef,
				attrs:          len(entity.Attributes),
				assocs:         assocCounts[entity.ID],
				validations:    len(entity.ValidationRules),
				indexes:        len(entity.Indexes),
				events:         len(entity.EventHandlers),
				accessRules:    len(entity.AccessRules),
			}
			rows = append(rows, r)

			if len(qualifiedName) > qnWidth {
				qnWidth = len(qualifiedName)
			}
			if len(entityType) > typeWidth {
				typeWidth = len(entityType)
			}
			if len(entity.GeneralizationRef) > genWidth {
				genWidth = len(entity.GeneralizationRef)
			}
		}
	}

	// Check if any entity has a generalization — only show column if needed
	hasGeneralizations := false
	for _, r := range rows {
		if r.generalization != "" {
			hasGeneralizations = true
			break
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Markdown table with aligned columns
	if hasGeneralizations {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %5s | %6s | %11s | %7s | %6s | %11s |\n",
			qnWidth, "Entity", typeWidth, "Type", genWidth, "Extends", "Attrs", "Assocs", "Validations", "Indexes", "Events", "AccessRules")
		fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-------|--------|-------------|---------|--------|-------------|\n",
			strings.Repeat("-", qnWidth), strings.Repeat("-", typeWidth), strings.Repeat("-", genWidth))
	} else {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %5s | %6s | %11s | %7s | %6s | %11s |\n",
			qnWidth, "Entity", typeWidth, "Type", "Attrs", "Assocs", "Validations", "Indexes", "Events", "AccessRules")
		fmt.Fprintf(e.output, "|-%s-|-%s-|-------|--------|-------------|---------|--------|-------------|\n",
			strings.Repeat("-", qnWidth), strings.Repeat("-", typeWidth))
	}
	for _, r := range rows {
		genStr := r.generalization
		// For System entities, show "-" instead of -1
		if r.entityType == "System" {
			if hasGeneralizations {
				fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %5s | %6s | %11s | %7s | %6s | %11s |\n",
					qnWidth, r.qualifiedName, typeWidth, r.entityType, genWidth, genStr, "-", "-", "-", "-", "-", "-")
			} else {
				fmt.Fprintf(e.output, "| %-*s | %-*s | %5s | %6s | %11s | %7s | %6s | %11s |\n",
					qnWidth, r.qualifiedName, typeWidth, r.entityType, "-", "-", "-", "-", "-", "-")
			}
		} else if hasGeneralizations {
			fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %5d | %6d | %11d | %7d | %6d | %11d |\n",
				qnWidth, r.qualifiedName, typeWidth, r.entityType, genWidth, genStr, r.attrs, r.assocs, r.validations, r.indexes, r.events, r.accessRules)
		} else {
			fmt.Fprintf(e.output, "| %-*s | %-*s | %5d | %6d | %11d | %7d | %6d | %11d |\n",
				qnWidth, r.qualifiedName, typeWidth, r.entityType, r.attrs, r.assocs, r.validations, r.indexes, r.events, r.accessRules)
		}
	}
	fmt.Fprintf(e.output, "\n(%d entities)\n", len(rows))
	return nil
}

// showEntity handles SHOW ENTITY command.
func (e *Executor) showEntity(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("entity name required")
	}

	module, err := e.findModule(name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	for _, entity := range dm.Entities {
		if entity.Name == name.Name {
			fmt.Fprintf(e.output, "**Entity: %s.%s**\n\n", module.Name, entity.Name)
			fmt.Fprintf(e.output, "- Persistable: %v\n", entity.Persistable)
			if entity.GeneralizationRef != "" {
				fmt.Fprintf(e.output, "- Extends: %s\n", entity.GeneralizationRef)
			}
			fmt.Fprintf(e.output, "- Location: (%d, %d)\n\n", entity.Location.X, entity.Location.Y)

			if len(entity.Attributes) > 0 {
				// Calculate column widths
				nameWidth, typeWidth := len("Attribute"), len("Type")
				type attrRow struct {
					name, typeName string
				}
				var rows []attrRow
				for _, attr := range entity.Attributes {
					typeName := getAttributeTypeName(attr.Type)
					rows = append(rows, attrRow{attr.Name, typeName})
					if len(attr.Name) > nameWidth {
						nameWidth = len(attr.Name)
					}
					if len(typeName) > typeWidth {
						typeWidth = len(typeName)
					}
				}

				fmt.Fprintf(e.output, "| %-*s | %-*s |\n", nameWidth, "Attribute", typeWidth, "Type")
				fmt.Fprintf(e.output, "|-%s-|-%s-|\n", strings.Repeat("-", nameWidth), strings.Repeat("-", typeWidth))
				for _, r := range rows {
					fmt.Fprintf(e.output, "| %-*s | %-*s |\n", nameWidth, r.name, typeWidth, r.typeName)
				}
				fmt.Fprintf(e.output, "\n(%d attributes)\n", len(entity.Attributes))
			}
			return nil
		}
	}

	return fmt.Errorf("entity not found: %s", name)
}

// describeEntity handles DESCRIBE ENTITY command.
func (e *Executor) describeEntity(name ast.QualifiedName) error {
	module, err := e.findModule(name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	for _, entity := range dm.Entities {
		if entity.Name == name.Name {
			// Output JavaDoc documentation if present
			if entity.Documentation != "" {
				fmt.Fprintf(e.output, "/**\n * %s\n */\n", entity.Documentation)
			}

			// Output position annotation
			fmt.Fprintf(e.output, "@Position(%d, %d)\n", entity.Location.X, entity.Location.Y)

			// Determine entity type based on Source field and Persistable flag
			entityType := "PERSISTENT"
			if strings.Contains(entity.Source, "OqlView") {
				entityType = "VIEW"
			} else if strings.Contains(entity.Source, "OData") || entity.RemoteSource != "" || entity.RemoteSourceDocument != "" {
				entityType = "EXTERNAL"
			} else if !entity.Persistable {
				entityType = "NON-PERSISTENT"
			}

			if entity.GeneralizationRef != "" {
				fmt.Fprintf(e.output, "CREATE OR MODIFY %s ENTITY %s.%s EXTENDS %s (\n", entityType, module.Name, entity.Name, entity.GeneralizationRef)
			} else {
				fmt.Fprintf(e.output, "CREATE OR MODIFY %s ENTITY %s.%s (\n", entityType, module.Name, entity.Name)
			}

			// Build validation rules map by attribute ID and name
			// The AttributeID can be a UUID or a qualified name like "DmTest.Cars.CarId"
			validationsByAttr := make(map[model.ID][]*domainmodel.ValidationRule)
			validationsByName := make(map[string][]*domainmodel.ValidationRule)
			for _, vr := range entity.ValidationRules {
				validationsByAttr[vr.AttributeID] = append(validationsByAttr[vr.AttributeID], vr)
				// Also index by attribute name extracted from qualified name
				attrName := extractAttrNameFromQualified(string(vr.AttributeID))
				if attrName != "" {
					validationsByName[attrName] = append(validationsByName[attrName], vr)
				}
			}

			// Output attributes
			for i, attr := range entity.Attributes {
				// Attribute documentation
				if attr.Documentation != "" {
					fmt.Fprintf(e.output, "  /** %s */\n", attr.Documentation)
				}

				typeStr := formatAttributeType(attr.Type)
				var constraints strings.Builder

				// Check for validation rules - try by ID first, then by name
				attrValidations := validationsByAttr[attr.ID]
				if len(attrValidations) == 0 {
					attrValidations = validationsByName[attr.Name]
				}
				for _, vr := range attrValidations {
					if vr.Type == "Required" {
						constraints.WriteString(" NOT NULL")
						if vr.ErrorMessage != nil {
							errMsg := vr.ErrorMessage.GetTranslation("en_US")
							if errMsg != "" {
								constraints.WriteString(fmt.Sprintf(" ERROR '%s'", errMsg))
							}
						}
					}
					if vr.Type == "Unique" {
						constraints.WriteString(" UNIQUE")
						if vr.ErrorMessage != nil {
							errMsg := vr.ErrorMessage.GetTranslation("en_US")
							if errMsg != "" {
								constraints.WriteString(fmt.Sprintf(" ERROR '%s'", errMsg))
							}
						}
					}
				}

				// Value type: CALCULATED or DEFAULT
				if attr.Value != nil && attr.Value.Type == "CalculatedValue" {
					constraints.WriteString(" CALCULATED")
					if attr.Value.MicroflowName != "" {
						constraints.WriteString(" BY " + attr.Value.MicroflowName)
					} else if attr.Value.MicroflowID != "" {
						if mfName := e.lookupMicroflowName(attr.Value.MicroflowID); mfName != "" {
							constraints.WriteString(" BY " + mfName)
						}
					}
				} else if attr.Value != nil && attr.Value.DefaultValue != "" {
					defaultVal := attr.Value.DefaultValue
					// Quote string defaults
					if _, ok := attr.Type.(*domainmodel.StringAttributeType); ok {
						defaultVal = fmt.Sprintf("'%s'", defaultVal)
					}
					// Re-qualify enum defaults for MDL syntax (BSON stores just the value name)
					if enumType, ok := attr.Type.(*domainmodel.EnumerationAttributeType); ok {
						if enumType.EnumerationRef != "" && !strings.Contains(defaultVal, ".") {
							defaultVal = enumType.EnumerationRef + "." + defaultVal
						}
					}
					constraints.WriteString(fmt.Sprintf(" DEFAULT %s", defaultVal))
				}

				comma := ","
				if i == len(entity.Attributes)-1 {
					comma = ""
				}
				fmt.Fprintf(e.output, "  %s: %s%s%s\n", attr.Name, typeStr, constraints.String(), comma)
			}
			fmt.Fprint(e.output, ")")

			// For VIEW entities, output the OQL query
			if entityType == "VIEW" && entity.OqlQuery != "" {
				fmt.Fprint(e.output, " AS (\n")
				// Indent OQL query lines
				oqlLines := strings.SplitSeq(entity.OqlQuery, "\n")
				for line := range oqlLines {
					fmt.Fprintf(e.output, "  %s\n", line)
				}
				fmt.Fprint(e.output, ")")
			}

			// Build attribute name map
			attrNames := make(map[model.ID]string)
			for _, attr := range entity.Attributes {
				attrNames[attr.ID] = attr.Name
			}

			// Output indexes
			for _, idx := range entity.Indexes {
				var cols []string
				for _, ia := range idx.Attributes {
					colName := attrNames[ia.AttributeID]
					if !ia.Ascending {
						colName += " DESC"
					}
					cols = append(cols, colName)
				}
				if len(cols) > 0 {
					fmt.Fprintf(e.output, "\nINDEX (%s)", strings.Join(cols, ", "))
				}
			}

			fmt.Fprintln(e.output, ";")

			// Output access rule GRANT statements
			e.outputEntityAccessGrants(entity, name.Module, name.Name)

			fmt.Fprintln(e.output, "/")
			return nil
		}
	}

	return fmt.Errorf("entity not found: %s", name)
}

// describeEntityToString generates MDL source for an entity and returns it as a string.
func (e *Executor) describeEntityToString(name ast.QualifiedName) (string, error) {
	var buf strings.Builder
	origOutput := e.output
	e.output = &buf
	err := e.describeEntity(name)
	e.output = origOutput
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// extractAttrNameFromQualified extracts the attribute name from a qualified name.
// e.g., "DmTest.Cars.CarId" -> "CarId"
func extractAttrNameFromQualified(qualifiedName string) string {
	// Split by "." and return the last part
	parts := strings.Split(qualifiedName, ".")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return ""
}

// resolveMicroflowByName resolves a qualified microflow name to its ID.
// It checks both microflows created during this session and existing microflows in the project.
func (e *Executor) resolveMicroflowByName(qualifiedName string) (model.ID, error) {
	parts := strings.Split(qualifiedName, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid microflow name: %s (expected Module.Name)", qualifiedName)
	}
	moduleName := parts[0]
	mfName := strings.Join(parts[1:], ".")

	// Check microflows created during this session
	if e.cache != nil && e.cache.createdMicroflows != nil {
		if info, ok := e.cache.createdMicroflows[qualifiedName]; ok {
			return info.ID, nil
		}
	}

	// Search existing microflows
	allMicroflows, err := e.reader.ListMicroflows()
	if err != nil {
		return "", fmt.Errorf("failed to list microflows: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return "", fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, mf := range allMicroflows {
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == moduleName && mf.Name == mfName {
			return mf.ID, nil
		}
	}

	return "", fmt.Errorf("microflow not found: %s", qualifiedName)
}

// lookupMicroflowName reverse-looks up a microflow ID to its qualified name.
func (e *Executor) lookupMicroflowName(mfID model.ID) string {
	allMicroflows, err := e.reader.ListMicroflows()
	if err != nil {
		return ""
	}

	h, err := e.getHierarchy()
	if err != nil {
		return ""
	}

	for _, mf := range allMicroflows {
		if mf.ID == mfID {
			modID := h.FindModuleID(mf.ContainerID)
			modName := h.GetModuleName(modID)
			if modName != "" {
				return modName + "." + mf.Name
			}
			return mf.Name
		}
	}
	return ""
}
