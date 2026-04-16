// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"encoding/json"
	"testing"
)

func TestSearchResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"uuid": "a7f3c2d1-4b5e-6c7f-8d9e-0a1b2c3d4e5f",
				"name": "CustomerService",
				"version": "1.2.0",
				"description": "Customer data service",
				"serviceType": "OData",
				"environment": {
					"name": "Production",
					"location": "EU",
					"type": "Production",
					"uuid": "env-uuid"
				},
				"application": {
					"name": "CRM App",
					"description": "CRM system",
					"uuid": "app-uuid",
					"businessOwner": {
						"name": "Business Owner",
						"email": "owner@example.com",
						"uuid": "owner-uuid"
					},
					"technicalOwner": {
						"name": "Technical Owner",
						"email": "tech@example.com",
						"uuid": "tech-uuid"
					}
				},
				"securityClassification": "Internal",
				"lastUpdated": "2026-04-16T10:00:00Z",
				"validated": true
			}
		],
		"totalResults": 42,
		"limit": 20,
		"offset": 0
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.TotalResults != 42 {
		t.Errorf("expected totalResults=42, got %d", resp.TotalResults)
	}
	if resp.Limit != 20 {
		t.Errorf("expected limit=20, got %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Errorf("expected offset=0, got %d", resp.Offset)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Data))
	}

	item := resp.Data[0]
	if item.UUID != "a7f3c2d1-4b5e-6c7f-8d9e-0a1b2c3d4e5f" {
		t.Errorf("unexpected uuid: %s", item.UUID)
	}
	if item.Name != "CustomerService" {
		t.Errorf("unexpected name: %s", item.Name)
	}
	if item.ServiceType != "OData" {
		t.Errorf("unexpected serviceType: %s", item.ServiceType)
	}
	if item.Environment.Type != "Production" {
		t.Errorf("unexpected environment type: %s", item.Environment.Type)
	}
	if item.Application.Name != "CRM App" {
		t.Errorf("unexpected application name: %s", item.Application.Name)
	}
}

func TestSearchResponse_EmptyResults(t *testing.T) {
	jsonData := `{
		"data": [],
		"totalResults": 0,
		"limit": 20,
		"offset": 0
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(resp.Data))
	}
	if resp.TotalResults != 0 {
		t.Errorf("expected totalResults=0, got %d", resp.TotalResults)
	}
}
