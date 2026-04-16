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
