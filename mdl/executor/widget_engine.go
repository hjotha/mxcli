// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"github.com/mendixlabs/mxcli/sdk/widgets"
	"go.mongodb.org/mongo-driver/bson"
)

// defaultSlotContainer is the MDLContainer name that receives default (non-containerized) child widgets.
const defaultSlotContainer = "TEMPLATE"

// =============================================================================
// Pluggable Widget Engine — Core Types
// =============================================================================

// WidgetDefinition describes how to construct a pluggable widget from MDL syntax.
// Loaded from embedded JSON definition files (*.def.json).
type WidgetDefinition struct {
	WidgetID         string             `json:"widgetId"`
	MDLName          string             `json:"mdlName"`
	WidgetKind       string             `json:"widgetKind,omitempty"` // "pluggable" (React) or "custom" (legacy Dojo)
	TemplateFile     string             `json:"templateFile"`
	DefaultEditable  string             `json:"defaultEditable"`
	PropertyMappings []PropertyMapping  `json:"propertyMappings,omitempty"`
	ChildSlots       []ChildSlotMapping `json:"childSlots,omitempty"`
	Modes            []WidgetMode       `json:"modes,omitempty"`
}

// WidgetMode defines a conditional configuration variant for a widget.
// For example, ComboBox has "enumeration" and "association" modes.
// Modes are evaluated in order; the first matching condition wins.
// A mode with no condition acts as the default fallback.
type WidgetMode struct {
	Name             string             `json:"name,omitempty"`
	Condition        string             `json:"condition,omitempty"`
	Description      string             `json:"description,omitempty"`
	PropertyMappings []PropertyMapping  `json:"propertyMappings"`
	ChildSlots       []ChildSlotMapping `json:"childSlots,omitempty"`
}

// ChildSlotMapping maps an MDL child container (e.g., TEMPLATE, FILTER) to a
// widget property that holds child widgets.
type ChildSlotMapping struct {
	PropertyKey  string `json:"propertyKey"`
	MDLContainer string `json:"mdlContainer"`
	Operation    string `json:"operation"`
}

// BuildContext carries resolved values from MDL parsing for use by operations.
type BuildContext struct {
	AttributePath  string
	AttributePaths []string // For operations that process multiple attributes
	AssocPath      string
	EntityName     string
	PrimitiveVal   string
	DataSource     pages.DataSource
	ChildWidgets   []bson.D
	ActionBSON     bson.D // Serialized client action BSON for opAction
	pageBuilder    *pageBuilder
}

// =============================================================================
// Pluggable Widget Engine
// =============================================================================

// PluggableWidgetEngine builds CustomWidget instances from WidgetDefinition + AST.
type PluggableWidgetEngine struct {
	operations  *OperationRegistry
	pageBuilder *pageBuilder
}

// NewPluggableWidgetEngine creates a new engine with the given registry and page builder.
func NewPluggableWidgetEngine(ops *OperationRegistry, pb *pageBuilder) *PluggableWidgetEngine {
	return &PluggableWidgetEngine{
		operations:  ops,
		pageBuilder: pb,
	}
}

// Build constructs a CustomWidget from a definition and AST widget node.
func (e *PluggableWidgetEngine) Build(def *WidgetDefinition, w *ast.WidgetV3) (*pages.CustomWidget, error) {
	// Save and restore entity context (DataSource mappings may change it)
	oldEntityContext := e.pageBuilder.entityContext
	defer func() { e.pageBuilder.entityContext = oldEntityContext }()

	// 1. Load template
	embeddedType, embeddedObject, embeddedIDs, embeddedObjectTypeID, err :=
		widgets.GetTemplateFullBSON(def.WidgetID, types.GenerateID, e.pageBuilder.reader.Path())
	if err != nil {
		return nil, mdlerrors.NewBackend("load "+def.MDLName+" template", err)
	}
	if embeddedType == nil || embeddedObject == nil {
		return nil, mdlerrors.NewNotFound("template", def.MDLName)
	}

	propertyTypeIDs := convertPropertyTypeIDs(embeddedIDs)
	updatedObject := embeddedObject

	// 2. Select mode and get mappings/slots
	mappings, slots, err := e.selectMappings(def, w)
	if err != nil {
		return nil, err
	}

	// 3. Apply property mappings
	for _, mapping := range mappings {
		ctx, err := e.resolveMapping(mapping, w)
		if err != nil {
			return nil, mdlerrors.NewBackend("resolve mapping for "+mapping.PropertyKey, err)
		}

		op := e.operations.Lookup(mapping.Operation)
		if op == nil {
			return nil, mdlerrors.NewValidationf("unknown operation %q for property %s", mapping.Operation, mapping.PropertyKey)
		}

		updatedObject = op(updatedObject, propertyTypeIDs, mapping.PropertyKey, ctx)
	}

	// 4. Apply child slots (.def.json)
	if err := e.applyChildSlots(slots, w, propertyTypeIDs, &updatedObject); err != nil {
		return nil, err
	}

	// 4.1 Auto datasource: map AST DataSource to first DataSource-type property.
	// Must run BEFORE child slots and explicit properties so entityContext is set.
	dsHandledByMapping := false
	for _, m := range mappings {
		if m.Source == "DataSource" {
			dsHandledByMapping = true
			break
		}
	}
	if !dsHandledByMapping {
		if ds := w.GetDataSource(); ds != nil {
			for propKey, entry := range propertyTypeIDs {
				if entry.ValueType == "DataSource" {
					dataSource, entityName, err := e.pageBuilder.buildDataSourceV3(ds)
					if err != nil {
						return nil, mdlerrors.NewBackend("auto datasource for "+propKey, err)
					}
					ctx := &BuildContext{DataSource: dataSource, EntityName: entityName}
					updatedObject = opDatasource(updatedObject, propertyTypeIDs, propKey, ctx)
					if entityName != "" {
						e.pageBuilder.entityContext = entityName
					}
					break
				}
			}
		}
	}

	// 4.3 Auto child slots: match AST children to Widgets-type template properties.
	// Two matching strategies:
	//   1. Named match: CONTAINER trigger { ... } → property "trigger" (by child name)
	//   2. Default slot: direct children not matching any named slot → first Widgets property
	// This allows pluggable widget child containers without requiring .def.json ChildSlot entries.
	handledSlotKeys := make(map[string]bool)
	for _, s := range slots {
		handledSlotKeys[s.PropertyKey] = true
	}
	// Collect Widgets-type property keys
	var widgetsPropKeys []string
	for propKey, entry := range propertyTypeIDs {
		if entry.ValueType == "Widgets" && !handledSlotKeys[propKey] {
			widgetsPropKeys = append(widgetsPropKeys, propKey)
		}
	}
	// Phase 1: Named matching — match children by name against property keys
	matchedChildren := make(map[int]bool) // indices of matched children
	for _, propKey := range widgetsPropKeys {
		upperKey := strings.ToUpper(propKey)
		for i, child := range w.Children {
			if matchedChildren[i] {
				continue
			}
			if strings.ToUpper(child.Name) == upperKey {
				var childBSONs []bson.D
				for _, slotChild := range child.Children {
					widgetBSON, err := e.pageBuilder.buildWidgetV3ToBSON(slotChild)
					if err != nil {
						return nil, err
					}
					if widgetBSON != nil {
						childBSONs = append(childBSONs, widgetBSON)
					}
				}
				if len(childBSONs) > 0 {
					updatedObject = opWidgets(updatedObject, propertyTypeIDs, propKey, &BuildContext{ChildWidgets: childBSONs})
					handledSlotKeys[propKey] = true
				}
				matchedChildren[i] = true
				break
			}
		}
	}
	// Phase 2: Default slot — unmatched direct children go to first unmatched Widgets property.
	// Skip if .def.json has childSlots defined — applyChildSlots already handles direct children.
	defSlotContainers := make(map[string]bool)
	for _, s := range slots {
		defSlotContainers[strings.ToUpper(s.MDLContainer)] = true
	}
	var defaultWidgetBSONs []bson.D
	for i, child := range w.Children {
		if matchedChildren[i] {
			continue
		}
		if len(slots) > 0 {
			continue // applyChildSlots handles both container and direct children
		}
		if defSlotContainers[strings.ToUpper(child.Type)] {
			continue
		}
		widgetBSON, err := e.pageBuilder.buildWidgetV3ToBSON(child)
		if err != nil {
			return nil, err
		}
		if widgetBSON != nil {
			defaultWidgetBSONs = append(defaultWidgetBSONs, widgetBSON)
		}
	}
	if len(defaultWidgetBSONs) > 0 {
		for _, propKey := range widgetsPropKeys {
			if !handledSlotKeys[propKey] {
				updatedObject = opWidgets(updatedObject, propertyTypeIDs, propKey, &BuildContext{ChildWidgets: defaultWidgetBSONs})
				break
			}
		}
	}

	// 4.6 Apply explicit properties (not covered by .def.json mappings)
	mappedKeys := make(map[string]bool)
	for _, m := range mappings {
		if m.Source != "" {
			mappedKeys[m.Source] = true
		}
	}
	for _, s := range slots {
		mappedKeys[s.MDLContainer] = true
	}
	for propName, propVal := range w.Properties {
		if mappedKeys[propName] || isBuiltinPropName(propName) {
			continue
		}
		entry, ok := propertyTypeIDs[propName]
		if !ok {
			continue // not a known widget property key
		}
		// Convert non-string values (bool, int, float) to string for property setting
		var strVal string
		switch v := propVal.(type) {
		case string:
			strVal = v
		case bool:
			strVal = fmt.Sprintf("%t", v)
		case int:
			strVal = fmt.Sprintf("%d", v)
		case float64:
			strVal = fmt.Sprintf("%g", v)
		default:
			continue
		}
		ctx := &BuildContext{}

		// Route by ValueType when available
		switch entry.ValueType {
		case "Expression":
			// Expression properties: set Expression field (not PrimitiveValue)
			ctx.PrimitiveVal = strVal
			updatedObject = opExpression(updatedObject, propertyTypeIDs, propName, ctx)
		case "TextTemplate":
			// TextTemplate properties: create ClientTemplate with attribute parameter binding.
			// Syntax: '{AttributeName} - {OtherAttr}' → text '{1} - {2}' with TemplateParameters.
			entityCtx := e.pageBuilder.entityContext
			tmplBSON := createClientTemplateBSONWithParams(strVal, entityCtx)
			updatedObject = updateWidgetPropertyValue(updatedObject, propertyTypeIDs, propName, func(val bson.D) bson.D {
				result := make(bson.D, 0, len(val))
				for _, elem := range val {
					if elem.Key == "TextTemplate" {
						result = append(result, bson.E{Key: "TextTemplate", Value: tmplBSON})
					} else {
						result = append(result, elem)
					}
				}
				return result
			})
		case "Attribute":
			// Attribute properties: resolve path
			if strings.Count(strVal, ".") >= 2 {
				ctx.AttributePath = strVal
			} else if e.pageBuilder.entityContext != "" {
				ctx.AttributePath = e.pageBuilder.resolveAttributePath(strVal)
			}
			if ctx.AttributePath != "" {
				updatedObject = opAttribute(updatedObject, propertyTypeIDs, propName, ctx)
			}
		default:
			// Known non-attribute types: always use primitive
			if entry.ValueType != "" && entry.ValueType != "Attribute" {
				ctx.PrimitiveVal = strVal
				updatedObject = opPrimitive(updatedObject, propertyTypeIDs, propName, ctx)
				continue
			}
			// Legacy routing for properties without ValueType info
			if strings.Count(strVal, ".") >= 2 {
				ctx.AttributePath = strVal
				updatedObject = opAttribute(updatedObject, propertyTypeIDs, propName, ctx)
			} else if e.pageBuilder.entityContext != "" && !strings.ContainsAny(strVal, " '\"") {
				ctx.AttributePath = e.pageBuilder.resolveAttributePath(strVal)
				updatedObject = opAttribute(updatedObject, propertyTypeIDs, propName, ctx)
			} else {
				ctx.PrimitiveVal = strVal
				updatedObject = opPrimitive(updatedObject, propertyTypeIDs, propName, ctx)
			}
		}
	}

	// 4.9 Auto-populate required empty object lists (e.g., Accordion groups, AreaChart series)
	updatedObject = ensureRequiredObjectLists(updatedObject, propertyTypeIDs)

	// 5. Build CustomWidget
	widgetID := model.ID(types.GenerateID())
	cw := &pages.CustomWidget{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       widgetID,
				TypeName: "CustomWidgets$CustomWidget",
			},
			Name: w.Name,
		},
		Label:             w.GetLabel(),
		Editable:          def.DefaultEditable,
		RawType:           embeddedType,
		RawObject:         updatedObject,
		PropertyTypeIDMap: propertyTypeIDs,
		ObjectTypeID:      embeddedObjectTypeID,
	}

	if err := e.pageBuilder.registerWidgetName(w.Name, cw.ID); err != nil {
		return nil, err
	}

	return cw, nil
}

// selectMappings selects the active PropertyMappings and ChildSlotMappings based on mode.
// Modes are evaluated in definition order; the first matching condition wins.
// A mode with no condition acts as the default fallback.
func (e *PluggableWidgetEngine) selectMappings(def *WidgetDefinition, w *ast.WidgetV3) ([]PropertyMapping, []ChildSlotMapping, error) {
	// No modes defined — use top-level mappings directly
	if len(def.Modes) == 0 {
		return def.PropertyMappings, def.ChildSlots, nil
	}

	// Evaluate modes in order; first match wins
	var fallback *WidgetMode
	var fallbackCount int
	for i := range def.Modes {
		mode := &def.Modes[i]
		if mode.Condition == "" {
			fallbackCount++
			if fallback == nil {
				fallback = mode
			}
			continue
		}
		if e.evaluateCondition(mode.Condition, w) {
			return mode.PropertyMappings, mode.ChildSlots, nil
		}
	}

	// Use fallback mode
	if fallback != nil {
		if fallbackCount > 1 {
			return nil, nil, mdlerrors.NewValidationf("widget %s has %d modes without conditions; only one default mode is allowed", def.MDLName, fallbackCount)
		}
		return fallback.PropertyMappings, fallback.ChildSlots, nil
	}

	return nil, nil, mdlerrors.NewValidationf("no matching mode for widget %s", def.MDLName)
}

// evaluateCondition checks a built-in condition string against the AST widget.
func (e *PluggableWidgetEngine) evaluateCondition(condition string, w *ast.WidgetV3) bool {
	switch {
	case condition == "hasDataSource":
		return w.GetDataSource() != nil
	case condition == "hasAttribute":
		return w.GetAttribute() != ""
	case strings.HasPrefix(condition, "hasProp:"):
		propName := strings.TrimPrefix(condition, "hasProp:")
		return w.GetStringProp(propName) != ""
	default:
		return false
	}
}

// resolveMapping resolves a PropertyMapping's source into a BuildContext.
func (e *PluggableWidgetEngine) resolveMapping(mapping PropertyMapping, w *ast.WidgetV3) (*BuildContext, error) {
	ctx := &BuildContext{pageBuilder: e.pageBuilder}

	// Static value takes priority
	if mapping.Value != "" {
		ctx.PrimitiveVal = mapping.Value
		return ctx, nil
	}

	source := mapping.Source
	if source == "" {
		return ctx, nil
	}

	switch source {
	case "Attribute":
		if attr := w.GetAttribute(); attr != "" {
			ctx.AttributePath = e.pageBuilder.resolveAttributePath(attr)
		}

	case "Attributes":
		if attrs := w.GetAttributes(); len(attrs) > 0 {
			ctx.AttributePaths = make([]string, 0, len(attrs))
			for _, attr := range attrs {
				ctx.AttributePaths = append(ctx.AttributePaths, e.pageBuilder.resolveAttributePath(attr))
			}
		}

	case "DataSource":
		if ds := w.GetDataSource(); ds != nil {
			dataSource, entityName, err := e.pageBuilder.buildDataSourceV3(ds)
			if err != nil {
				return nil, mdlerrors.NewBackend("build datasource", err)
			}
			ctx.DataSource = dataSource
			ctx.EntityName = entityName
			if entityName != "" {
				e.pageBuilder.entityContext = entityName
				if w.Name != "" {
					e.pageBuilder.paramEntityNames[w.Name] = entityName
				}
			}
		}

	case "Selection":
		val := w.GetSelection()
		if val == "" && mapping.Default != "" {
			val = mapping.Default
		}
		ctx.PrimitiveVal = val

	case "CaptionAttribute":
		if captionAttr := w.GetStringProp("CaptionAttribute"); captionAttr != "" {
			// Resolve relative to entity context
			if !strings.Contains(captionAttr, ".") && e.pageBuilder.entityContext != "" {
				captionAttr = e.pageBuilder.entityContext + "." + captionAttr
			}
			ctx.AttributePath = captionAttr
		}

	case "Association":
		// For association operation: resolve both assoc path AND entity name from DataSource
		if attr := w.GetAttribute(); attr != "" {
			ctx.AssocPath = e.pageBuilder.resolveAssociationPath(attr)
		}
		// Entity name comes from DataSource context (must be resolved first by a DataSource mapping)
		ctx.EntityName = e.pageBuilder.entityContext
		if ctx.AssocPath != "" && ctx.EntityName == "" {
			return nil, mdlerrors.NewValidationf("association %q requires an entity context (add a DataSource mapping before Association)", ctx.AssocPath)
		}

	case "OnClick":
		// Resolve AST action (stored as Properties["Action"]) into serialized BSON
		if action := w.GetAction(); action != nil {
			act, err := e.pageBuilder.buildClientActionV3(action)
			if err != nil {
				return nil, mdlerrors.NewBackend("build action", err)
			}
			ctx.ActionBSON = mpr.SerializeClientAction(act)
		}

	default:
		// Generic fallback: treat source as a property name on the AST widget
		val := w.GetStringProp(source)
		if val == "" && mapping.Default != "" {
			val = mapping.Default
		}
		ctx.PrimitiveVal = val
	}

	return ctx, nil
}

// applyChildSlots processes child slot mappings, building child widgets and embedding them.
func (e *PluggableWidgetEngine) applyChildSlots(slots []ChildSlotMapping, w *ast.WidgetV3, propertyTypeIDs map[string]pages.PropertyTypeIDEntry, updatedObject *bson.D) error {
	if len(slots) == 0 {
		return nil
	}

	// Build a set of slot container names for matching
	slotContainers := make(map[string]*ChildSlotMapping, len(slots))
	for i := range slots {
		slotContainers[slots[i].MDLContainer] = &slots[i]
	}

	// Group children by slot
	slotWidgets := make(map[string][]bson.D)
	var defaultWidgets []bson.D

	for _, child := range w.Children {
		upperType := strings.ToUpper(child.Type)
		if slot, ok := slotContainers[upperType]; ok {
			// Container matches a slot — build its children
			for _, slotChild := range child.Children {
				widgetBSON, err := e.pageBuilder.buildWidgetV3ToBSON(slotChild)
				if err != nil {
					return err
				}
				if widgetBSON != nil {
					slotWidgets[slot.PropertyKey] = append(slotWidgets[slot.PropertyKey], widgetBSON)
				}
			}
		} else {
			// Direct child — default content
			widgetBSON, err := e.pageBuilder.buildWidgetV3ToBSON(child)
			if err != nil {
				return err
			}
			if widgetBSON != nil {
				defaultWidgets = append(defaultWidgets, widgetBSON)
			}
		}
	}

	// Apply each slot's widgets via its operation
	for _, slot := range slots {
		childBSONs := slotWidgets[slot.PropertyKey]
		// If no explicit container children, use default widgets for the first slot
		if len(childBSONs) == 0 && len(defaultWidgets) > 0 && slot.MDLContainer == defaultSlotContainer {
			childBSONs = defaultWidgets
			defaultWidgets = nil // consume once
		}
		if len(childBSONs) == 0 {
			continue
		}

		op := e.operations.Lookup(slot.Operation)
		if op == nil {
			return mdlerrors.NewValidationf("unknown operation %q for child slot %s", slot.Operation, slot.PropertyKey)
		}

		ctx := &BuildContext{ChildWidgets: childBSONs}
		*updatedObject = op(*updatedObject, propertyTypeIDs, slot.PropertyKey, ctx)
	}

	return nil
}

// isBuiltinPropName returns true for property names that are handled by
// dedicated MDL keywords (DataSource, Attribute, etc.) rather than by
// the explicit property pass.
func isBuiltinPropName(name string) bool {
	switch name {
	case "DataSource", "Attribute", "Label", "Caption", "Action",
		"Selection", "Class", "Style", "Editable", "Visible",
		"WidgetType", "DesignProperties", "Association", "CaptionAttribute",
		"Content", "RenderMode", "ContentParams", "CaptionParams",
		"ButtonStyle", "DesktopWidth", "DesktopColumns", "TabletColumns",
		"PhoneColumns", "PageSize", "Pagination", "PagingPosition",
		"ShowPagingButtons", "Attributes", "FilterType", "Width", "Height",
		"Tooltip", "Name":
		return true
	}
	return false
}
