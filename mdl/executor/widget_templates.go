// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/mendixlabs/mxcli/sdk/mpr"
	"go.mongodb.org/mongo-driver/bson"
)

// createClientTemplateBSONWithParams creates a Forms$ClientTemplate that supports
// attribute parameter binding. Syntax: '{AttrName} - {OtherAttr}' extracts attribute
// names from curly braces, replaces them with {1}, {2}, etc., and generates
// TemplateParameter entries with AttributeRef bindings.
// If no {AttrName} patterns are found, creates a static text template.
func createClientTemplateBSONWithParams(text string, entityContext string) bson.D {
	// Extract {AttributeName} patterns and build parameter list
	re := regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)
	matches := re.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		// No attribute references — static text
		return createDefaultClientTemplateBSON(text)
	}

	// Replace {AttrName} with {1}, {2}, etc. and collect attribute names
	var attrNames []string
	paramText := text
	// Process in reverse to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		attrName := text[match[2]:match[3]]
		// Check if it's a pure number (like {1}) — keep as-is
		if _, err := fmt.Sscanf(attrName, "%d", new(int)); err == nil {
			continue
		}
		attrNames = append([]string{attrName}, attrNames...) // prepend
		paramText = paramText[:match[0]] + fmt.Sprintf("{%d}", len(attrNames)) + paramText[match[1]:]
	}

	// Rebuild paramText with sequential numbering
	paramText = text
	attrNames = nil
	for i := 0; i < len(matches); i++ {
		match := matches[i]
		attrName := text[match[2]:match[3]]
		if _, err := fmt.Sscanf(attrName, "%d", new(int)); err == nil {
			continue
		}
		attrNames = append(attrNames, attrName)
	}
	paramText = re.ReplaceAllStringFunc(text, func(s string) string {
		name := s[1 : len(s)-1]
		if _, err := fmt.Sscanf(name, "%d", new(int)); err == nil {
			return s // keep numeric {1} as-is
		}
		for i, an := range attrNames {
			if an == name {
				return fmt.Sprintf("{%d}", i+1)
			}
		}
		return s
	})

	// Build parameters BSON
	params := bson.A{int32(2)} // version marker for non-empty array
	for _, attrName := range attrNames {
		attrPath := attrName
		if entityContext != "" && !strings.Contains(attrName, ".") {
			attrPath = entityContext + "." + attrName
		}
		params = append(params, bson.D{
			{Key: "$ID", Value: generateBinaryID()},
			{Key: "$Type", Value: "Forms$ClientTemplateParameter"},
			{Key: "AttributeRef", Value: bson.D{
				{Key: "$ID", Value: generateBinaryID()},
				{Key: "$Type", Value: "DomainModels$AttributeRef"},
				{Key: "Attribute", Value: attrPath},
				{Key: "EntityRef", Value: nil},
			}},
			{Key: "Expression", Value: ""},
			{Key: "FormattingInfo", Value: bson.D{
				{Key: "$ID", Value: generateBinaryID()},
				{Key: "$Type", Value: "Forms$FormattingInfo"},
				{Key: "CustomDateFormat", Value: ""},
				{Key: "DateFormat", Value: "Date"},
				{Key: "DecimalPrecision", Value: int64(2)},
				{Key: "EnumFormat", Value: "Text"},
				{Key: "GroupDigits", Value: false},
				{Key: "TimeFormat", Value: "HoursMinutes"},
			}},
			{Key: "SourceVariable", Value: nil},
		})
	}

	makeText := func(t string) bson.D {
		return bson.D{
			{Key: "$ID", Value: generateBinaryID()},
			{Key: "$Type", Value: "Texts$Text"},
			{Key: "Items", Value: bson.A{int32(3), bson.D{
				{Key: "$ID", Value: generateBinaryID()},
				{Key: "$Type", Value: "Texts$Translation"},
				{Key: "LanguageCode", Value: "en_US"},
				{Key: "Text", Value: t},
			}}},
		}
	}

	return bson.D{
		{Key: "$ID", Value: generateBinaryID()},
		{Key: "$Type", Value: "Forms$ClientTemplate"},
		{Key: "Fallback", Value: makeText(paramText)},
		{Key: "Parameters", Value: params},
		{Key: "Template", Value: makeText(paramText)},
	}
}

// createDefaultClientTemplateBSON creates a Forms$ClientTemplate with an en_US translation.
func createDefaultClientTemplateBSON(text string) bson.D {
	makeText := func(t string) bson.D {
		return bson.D{
			{Key: "$ID", Value: generateBinaryID()},
			{Key: "$Type", Value: "Texts$Text"},
			{Key: "Items", Value: bson.A{int32(3), bson.D{
				{Key: "$ID", Value: generateBinaryID()},
				{Key: "$Type", Value: "Texts$Translation"},
				{Key: "LanguageCode", Value: "en_US"},
				{Key: "Text", Value: t},
			}}},
		}
	}
	return bson.D{
		{Key: "$ID", Value: generateBinaryID()},
		{Key: "$Type", Value: "Forms$ClientTemplate"},
		{Key: "Fallback", Value: makeText(text)},
		{Key: "Parameters", Value: bson.A{int32(2)}},
		{Key: "Template", Value: makeText(text)},
	}
}

// generateBinaryID creates a new random 16-byte UUID in Microsoft GUID binary format.
func generateBinaryID() []byte {
	return hexIDToBlob(mpr.GenerateID())
}

// hexIDToBlob converts a hex UUID string to a 16-byte binary blob in Microsoft GUID format.
func hexIDToBlob(hexStr string) []byte {
	hexStr = strings.ReplaceAll(hexStr, "-", "")
	data, err := hex.DecodeString(hexStr)
	if err != nil || len(data) != 16 {
		return data
	}
	// Swap bytes to match Microsoft GUID format (little-endian for first 3 segments)
	data[0], data[1], data[2], data[3] = data[3], data[2], data[1], data[0]
	data[4], data[5] = data[5], data[4]
	data[6], data[7] = data[7], data[6]
	return data
}
