// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// =============================================================================
// formatRestCallAction
// =============================================================================

func TestFormatRestCallAction_GET(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:       microflows.HttpMethodGet,
			LocationTemplate: "https://api.example.com/orders",
		},
		ResultHandling: &microflows.ResultHandlingString{VariableName: "Response"},
	}
	got := e.formatRestCallAction(action)
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	assertContains(t, got, "rest call get")
	assertContains(t, got, "'https://api.example.com/orders'")
	assertContains(t, got, "$Response = ")
	assertContains(t, got, "returns String")
}

func TestFormatRestCallAction_HttpResponse(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:       microflows.HttpMethodGet,
			LocationTemplate: "https://api.example.com/orders",
		},
		ResultHandling: &microflows.ResultHandlingHttpResponse{VariableName: "Response"},
	}
	got := e.formatRestCallAction(action)
	assertContains(t, got, "$Response = ")
	assertContains(t, got, "returns response")
}

func TestFormatRestCallAction_POST_CustomBody(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:       microflows.HttpMethodPost,
			LocationTemplate: "https://api.example.com/orders",
		},
		RequestHandling: &microflows.CustomRequestHandling{
			Template: `{"name": "test"}`,
		},
		ResultHandling: &microflows.ResultHandlingNone{},
	}
	got := e.formatRestCallAction(action)
	assertContains(t, got, "rest call post")
	assertContains(t, got, "body '{\"name\": \"test\"}'")
	assertContains(t, got, "returns Nothing")
}

func TestFormatRestCallAction_WithHeaders(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:       microflows.HttpMethodGet,
			LocationTemplate: "https://api.example.com",
			CustomHeaders: []*microflows.HttpHeader{
				{Name: "Authorization", Value: "'Bearer ' + $Token"},
			},
		},
		ResultHandling: &microflows.ResultHandlingString{VariableName: "Resp"},
	}
	got := e.formatRestCallAction(action)
	assertContains(t, got, "header 'Authorization' = 'Bearer ' + $Token")
}

func TestFormatRestCallAction_WithAuth(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:        microflows.HttpMethodGet,
			LocationTemplate:  "https://api.example.com",
			UseAuthentication: true,
			Username:          "'admin'",
			Password:          "'secret'",
		},
		ResultHandling: &microflows.ResultHandlingString{},
	}
	got := e.formatRestCallAction(action)
	assertContains(t, got, "auth basic 'admin' password 'secret'")
}

func TestFormatRestCallAction_WithTimeout(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RestCallAction{
		HttpConfiguration: &microflows.HttpConfiguration{
			HttpMethod:       microflows.HttpMethodGet,
			LocationTemplate: "https://api.example.com",
		},
		TimeoutExpression: "30",
		ResultHandling:    &microflows.ResultHandlingString{},
	}
	got := e.formatRestCallAction(action)
	assertContains(t, got, "timeout 30")
}
