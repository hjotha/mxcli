// SPDX-License-Identifier: Apache-2.0

package catalog

// SearchOptions contains parameters for catalog search requests.
type SearchOptions struct {
	Query                   string
	ServiceType             string // "OData", "REST", "SOAP", "" (all)
	ProductionEndpointsOnly bool
	OwnedContentOnly        bool
	Limit                   int
	Offset                  int
}

// SearchResponse represents the response from GET /data endpoint.
type SearchResponse struct {
	Data         []SearchResult `json:"data"`
	TotalResults int            `json:"totalResults"`
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
}

// SearchResult represents a single data source/service in the catalog.
type SearchResult struct {
	UUID                   string      `json:"uuid"`
	Name                   string      `json:"name"`
	Version                string      `json:"version"`
	Description            string      `json:"description"`
	ServiceType            string      `json:"serviceType"`
	Environment            Environment `json:"environment"`
	Application            Application `json:"application"`
	SecurityClassification string      `json:"securityClassification"`
	LastUpdated            string      `json:"lastUpdated"`
	Validated              bool        `json:"validated"`
}

// Environment represents the deployment environment of a service.
type Environment struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Type     string `json:"type"` // "Production", "Acceptance", "Test"
	UUID     string `json:"uuid"`
}

// Application represents the Mendix app hosting a service.
type Application struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	UUID           string `json:"uuid"`
	BusinessOwner  *Owner `json:"businessOwner,omitempty"`
	TechnicalOwner *Owner `json:"technicalOwner,omitempty"`
}

// Owner represents a business or technical owner.
type Owner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	UUID  string `json:"uuid"`
}

// EndpointDetails represents the full response from GET /endpoints/{uuid}.
type EndpointDetails struct {
	UUID                   string             `json:"uuid"`
	Path                   string             `json:"path"`
	Location               string             `json:"location"`
	Discoverable           bool               `json:"discoverable"`
	Validated              bool               `json:"validated"`
	SecurityClassification string             `json:"securityClassification"`
	Connections            int                `json:"connections"`
	LastUpdated            string             `json:"lastUpdated"`
	ServiceVersion         ServiceVersion     `json:"serviceVersion"`
	Environment            EnvironmentWithApp `json:"environment"`
}

// EnvironmentWithApp extends Environment with nested Application (used in endpoint details).
type EnvironmentWithApp struct {
	Name        string      `json:"name"`
	Location    string      `json:"location"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	Application Application `json:"application"`
}

// ServiceVersion contains version-specific metadata and contract.
type ServiceVersion struct {
	Version        string          `json:"version"`
	Description    string          `json:"description"`
	UUID           string          `json:"uuid"`
	PublishDate    string          `json:"publishDate"`
	Type           string          `json:"type"` // "OData", "REST", "SOAP"
	Contracts      []Contract      `json:"contracts"`
	SecurityScheme *SecurityScheme `json:"securityScheme,omitempty"`
	TotalEntities  int             `json:"totalEntities"`
	Entities       []Entity        `json:"entities"`
	TotalActions   int             `json:"totalActions"`
	Actions        []Action        `json:"actions"`
}

// Contract represents an API contract (metadata, OpenAPI, WSDL, etc).
type Contract struct {
	Type                 string     `json:"type"` // "CSDL", "OpenAPI", "WSDL"
	SpecificationVersion string     `json:"specificationVersion"`
	DocumentBaseURL      string     `json:"documentBaseURL"`
	Documents            []Document `json:"documents"`
}

// Document contains the actual contract content.
type Document struct {
	IsPrimary bool   `json:"isPrimary"`
	URI       string `json:"uri"`
	Contents  string `json:"contents"` // Embedded XML/JSON contract
}

// SecurityScheme describes authentication requirements.
type SecurityScheme struct {
	SecurityTypes  []SecurityType `json:"securityTypes"`
	MxAllowedRoles []Role         `json:"mxAllowedRoles,omitempty"`
}

// SecurityType represents an authentication method.
type SecurityType struct {
	Name                string `json:"name"` // "Basic", "MxID", etc.
	MarketplaceModuleID string `json:"marketplaceModuleID,omitempty"`
}

// Role represents a Mendix module role.
type Role struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// Entity represents an OData entity set.
type Entity struct {
	Name              string        `json:"name"`
	EntitySetName     string        `json:"entitySetName"`
	EntityTypeName    string        `json:"entityTypeName"`
	Namespace         string        `json:"namespace"`
	Validated         bool          `json:"validated"`
	Updatable         bool          `json:"updatable"`
	Insertable        bool          `json:"insertable"`
	Deletable         bool          `json:"deletable"`
	TotalAttributes   int           `json:"totalAttributes"`
	Attributes        []Attribute   `json:"attributes"`
	TotalAssociations int           `json:"totalAssociations"`
	Associations      []Association `json:"associations"`
}

// Attribute represents an entity attribute.
type Attribute struct {
	Name       string `json:"name"`
	TypeName   string `json:"typeName"`
	TypeKind   string `json:"typeKind"`
	Updatable  bool   `json:"updatable"`
	Insertable bool   `json:"insertable"`
	Filterable bool   `json:"filterable"`
	Sortable   bool   `json:"sortable"`
}

// Association represents an entity association.
type Association struct {
	Name              string `json:"name"`
	ReferencedDataset string `json:"referencedDataset"`
	Multiplicity      string `json:"multiplicity"`
	EntitySetName     string `json:"entitySetName"`
	EntityTypeName    string `json:"entityTypeName"`
	Namespace         string `json:"namespace"`
}

// Action represents an OData action or function.
type Action struct {
	Name               string      `json:"name"`
	FullyQualifiedName string      `json:"fullyQualifiedName"`
	Summary            string      `json:"summary"`
	Description        string      `json:"description"`
	TotalParameters    int         `json:"totalParameters"`
	Parameters         []Parameter `json:"parameters"`
	ReturnType         *ReturnType `json:"returnType,omitempty"`
}

// Parameter represents an action/function parameter.
type Parameter struct {
	Name         string `json:"name"`
	TypeKind     string `json:"typekind"`
	TypeName     string `json:"typeName"`
	IsCollection bool   `json:"isCollection"`
	Nullable     bool   `json:"nullable"`
	Summary      string `json:"summary"`
	Description  string `json:"description"`
}

// ReturnType represents an action/function return type.
type ReturnType struct {
	TypeKind     string `json:"typekind"`
	TypeName     string `json:"typeName"`
	IsCollection bool   `json:"isCollection"`
	Nullable     bool   `json:"nullable"`
}
