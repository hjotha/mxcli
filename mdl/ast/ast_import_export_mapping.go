// SPDX-License-Identifier: Apache-2.0

package ast

// ============================================================================
// Import Mapping Statements
// ============================================================================

// CreateImportMappingStmt represents:
//
//	CREATE IMPORT MAPPING Module.Name
//	  [FROM JSON STRUCTURE Module.JsonStructure | FROM XML SCHEMA Module.Schema]
//	{ root AS Module.Entity (Create) { ... } }
type CreateImportMappingStmt struct {
	Name        QualifiedName
	SchemaKind  string        // "JSON_STRUCTURE" or "XML_SCHEMA" or ""
	SchemaRef   QualifiedName // qualified name of the schema source
	RootElement *ImportMappingElementDef
}

func (s *CreateImportMappingStmt) isStatement() {}

// DropImportMappingStmt represents: DROP IMPORT MAPPING Module.Name
type DropImportMappingStmt struct {
	Name QualifiedName
}

func (s *DropImportMappingStmt) isStatement() {}

// ImportMappingElementDef represents one element in the mapping tree.
// It may be an object mapping (AS entity) or a value mapping (AS attribute).
type ImportMappingElementDef struct {
	// JSON field name (or "root" for the root element)
	JsonName string
	// Object mapping fields (set when mapping to an entity)
	Entity         string // qualified entity name (e.g. "Module.Customer")
	ObjectHandling string // "Create", "Find", "FindOrCreate", "Custom"
	Association    string // qualified association name for via clause
	Children       []*ImportMappingElementDef
	// Value mapping fields (set when mapping to an attribute)
	Attribute string // attribute name (unqualified, e.g. "Name")
	DataType  string // "String", "Integer", "Boolean", "Decimal", "DateTime"
	IsKey     bool
}

// ============================================================================
// Export Mapping Statements
// ============================================================================

// CreateExportMappingStmt represents:
//
//	CREATE EXPORT MAPPING Module.Name
//	  [TO JSON STRUCTURE Module.JsonStructure | TO XML SCHEMA Module.Schema]
//	  [NULL VALUES LeaveOutElement | SendAsNil]
//	{ Module.Entity AS root { ... } }
type CreateExportMappingStmt struct {
	Name            QualifiedName
	SchemaKind      string        // "JSON_STRUCTURE" or "XML_SCHEMA" or ""
	SchemaRef       QualifiedName // qualified name of the schema source
	NullValueOption string        // "LeaveOutElement" or "SendAsNil" (default: "LeaveOutElement")
	RootElement     *ExportMappingElementDef
}

func (s *CreateExportMappingStmt) isStatement() {}

// DropExportMappingStmt represents: DROP EXPORT MAPPING Module.Name
type DropExportMappingStmt struct {
	Name QualifiedName
}

func (s *DropExportMappingStmt) isStatement() {}

// ExportMappingElementDef represents one element in an export mapping tree.
// It may be an object mapping (entity AS JSON key) or a value mapping (attribute AS JSON key).
type ExportMappingElementDef struct {
	// JSON field name (the RHS of AS)
	JsonName string
	// Object mapping fields (set when mapping from an entity)
	Entity      string // qualified entity name (e.g. "Module.Customer")
	Association string // qualified association name for VIA clause
	Children    []*ExportMappingElementDef
	// Value mapping fields (set when mapping from an attribute)
	Attribute string // attribute name (unqualified, e.g. "Name")
	DataType  string // "String", "Integer", "Boolean", "Decimal", "DateTime"
}
