// SPDX-License-Identifier: Apache-2.0

package ast

// ============================================================================
// Data Types
// ============================================================================

// DataType represents an MDL attribute data type.
type DataType struct {
	Kind            DataTypeKind
	Length          int            // For String(length), -1 for unlimited
	Precision       int            // For Decimal(p,s)
	Scale           int            // For Decimal(p,s)
	EnumRef         *QualifiedName // For Enumeration(Module.EnumName)
	EntityRef       *QualifiedName // For Entity or List of Entity types
	TemplateContext string         // For StringTemplate(Sql), stores "Sql", "OQL", etc.
	TypeParamName   string         // For TypeEntityTypeParam: the declared name (e.g., "pEntity")
}

// DataTypeKind represents the kind of data type.
type DataTypeKind int

const (
	TypeUnknown DataTypeKind = iota // Unknown or unresolvable type
	TypeString
	TypeInteger
	TypeLong
	TypeDecimal
	TypeBoolean
	TypeDateTime
	TypeDate
	TypeAutoNumber
	TypeBinary
	TypeEnumeration
	TypeEntity          // Entity reference (for microflow parameters)
	TypeListOf          // List of entity (for microflow parameters)
	TypeVoid            // Void return type (for microflows)
	TypeStringTemplate  // StringTemplate(Sql) etc. for Java actions
	TypeEntityTypeParam // ENTITY <pEntity> type parameter declaration for Java actions
)

func (k DataTypeKind) String() string {
	switch k {
	case TypeString:
		return "String"
	case TypeInteger:
		return "Integer"
	case TypeLong:
		return "Long"
	case TypeDecimal:
		return "Decimal"
	case TypeBoolean:
		return "Boolean"
	case TypeDateTime:
		return "DateTime"
	case TypeDate:
		return "Date"
	case TypeAutoNumber:
		return "AutoNumber"
	case TypeBinary:
		return "Binary"
	case TypeEnumeration:
		return "Enumeration"
	case TypeEntity:
		return "Entity"
	case TypeListOf:
		return "List"
	case TypeVoid:
		return "Void"
	case TypeStringTemplate:
		return "StringTemplate"
	case TypeEntityTypeParam:
		return "EntityTypeParam"
	default:
		return "Unknown"
	}
}
