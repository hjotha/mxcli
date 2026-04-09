// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"github.com/mendixlabs/mxcli/sdk/pages"
	"go.mongodb.org/mongo-driver/bson"
)

// =============================================================================
// Default Object List Population
// =============================================================================

// ensureRequiredObjectLists populates empty Object list properties with one default
// entry. This prevents CE0642 "Property 'X' is required" errors for widget properties
// like Accordion groups, AreaChart series, etc.
func ensureRequiredObjectLists(obj bson.D, propertyTypeIDs map[string]pages.PropertyTypeIDEntry) bson.D {
	for propKey, entry := range propertyTypeIDs {
		if entry.ObjectTypeID == "" || len(entry.NestedPropertyIDs) == 0 {
			continue
		}
		// Skip non-required object lists that have nested DataSource properties —
		// auto-populating these creates entries that trigger widget-level validation errors.
		// Required object lists (like AreaChart series) are populated even with nested DataSource
		// because the DataSource is conditional (e.g., depends on dataSet enum).
		if !entry.Required {
			hasNestedDS := false
			for _, nested := range entry.NestedPropertyIDs {
				if nested.ValueType == "DataSource" {
					hasNestedDS = true
					break
				}
			}
			if hasNestedDS {
				continue
			}
		}
		// Skip if any Required nested property is Attribute (needs entity context)
		hasRequiredAttr := false
		for _, nested := range entry.NestedPropertyIDs {
			if nested.Required && nested.ValueType == "Attribute" {
				hasRequiredAttr = true
				break
			}
		}
		if hasRequiredAttr {
			continue
		}
		obj = updateWidgetPropertyValue(obj, propertyTypeIDs, propKey, func(val bson.D) bson.D {
			for _, elem := range val {
				if elem.Key == "Objects" {
					if arr, ok := elem.Value.(bson.A); ok && len(arr) <= 1 {
						// Empty Objects array — create one default entry
						defaultObj := createDefaultWidgetObject(entry.ObjectTypeID, entry.NestedPropertyIDs)
						newArr := bson.A{int32(2), defaultObj}
						result := make(bson.D, 0, len(val))
						for _, e := range val {
							if e.Key == "Objects" {
								result = append(result, bson.E{Key: "Objects", Value: newArr})
							} else {
								result = append(result, e)
							}
						}
						return result
					}
				}
			}
			return val
		})
	}
	return obj
}

// createDefaultWidgetObject creates a minimal WidgetObject BSON entry for an object list.
func createDefaultWidgetObject(objectTypeID string, nestedProps map[string]pages.PropertyTypeIDEntry) bson.D {
	propsArr := bson.A{int32(2)} // version marker
	for _, entry := range nestedProps {
		prop := createDefaultWidgetProperty(entry)
		propsArr = append(propsArr, prop)
	}
	return bson.D{
		{Key: "$ID", Value: generateBinaryID()},
		{Key: "$Type", Value: "CustomWidgets$WidgetObject"},
		{Key: "TypePointer", Value: hexIDToBlob(objectTypeID)},
		{Key: "Properties", Value: propsArr},
	}
}

// createDefaultWidgetProperty creates a WidgetProperty with default WidgetValue.
func createDefaultWidgetProperty(entry pages.PropertyTypeIDEntry) bson.D {
	return bson.D{
		{Key: "$ID", Value: generateBinaryID()},
		{Key: "$Type", Value: "CustomWidgets$WidgetProperty"},
		{Key: "TypePointer", Value: hexIDToBlob(entry.PropertyTypeID)},
		{Key: "Value", Value: createDefaultWidgetValue(entry)},
	}
}

// createDefaultWidgetValue creates a WidgetValue with standard default fields.
// Sets type-specific defaults: Expression→Expression field, TextTemplate→template, etc.
func createDefaultWidgetValue(entry pages.PropertyTypeIDEntry) bson.D {
	primitiveVal := entry.DefaultValue
	expressionVal := ""
	var textTemplate interface{} // nil by default

	// Route default value to the correct field based on ValueType
	switch entry.ValueType {
	case "Expression":
		expressionVal = primitiveVal
		primitiveVal = ""
	case "TextTemplate":
		// Create a ClientTemplate with a placeholder translation to satisfy CE4899
		text := primitiveVal
		if text == "" {
			text = " " // non-empty to satisfy "required" translation check
		}
		textTemplate = createDefaultClientTemplateBSON(text)
	case "String":
		if primitiveVal == "" {
			primitiveVal = " " // non-empty to satisfy required String properties
		}
	}

	return bson.D{
		{Key: "$ID", Value: generateBinaryID()},
		{Key: "$Type", Value: "CustomWidgets$WidgetValue"},
		{Key: "Action", Value: bson.D{
			{Key: "$ID", Value: generateBinaryID()},
			{Key: "$Type", Value: "Forms$NoAction"},
			{Key: "DisabledDuringExecution", Value: true},
		}},
		{Key: "AttributeRef", Value: nil},
		{Key: "DataSource", Value: nil},
		{Key: "EntityRef", Value: nil},
		{Key: "Expression", Value: expressionVal},
		{Key: "Form", Value: ""},
		{Key: "Icon", Value: nil},
		{Key: "Image", Value: ""},
		{Key: "Microflow", Value: ""},
		{Key: "Nanoflow", Value: ""},
		{Key: "Objects", Value: bson.A{int32(2)}},
		{Key: "PrimitiveValue", Value: primitiveVal},
		{Key: "Selection", Value: "None"},
		{Key: "SourceVariable", Value: nil},
		{Key: "TextTemplate", Value: textTemplate},
		{Key: "TranslatableValue", Value: nil},
		{Key: "TypePointer", Value: hexIDToBlob(entry.ValueTypeID)},
		{Key: "Widgets", Value: bson.A{int32(2)}},
		{Key: "XPathConstraint", Value: ""},
	}
}
