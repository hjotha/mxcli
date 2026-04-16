// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Sample responses mirror what the 2026-04 spike recorded from the real
// API. If Mendix changes the response shape, these sample strings are the
// first place to update.

const sampleContentList = `{
  "items": [
    {
      "contentId": 170,
      "publisher": "Mendix",
      "type": "Module",
      "categories": [{"name": "Utility"}],
      "supportCategory": "Platform",
      "licenseUrl": "http://www.apache.org/licenses/LICENSE-2.0.html",
      "isPrivate": false,
      "latestVersion": {
        "name": "Community Commons",
        "versionId": "0a03e65a-d94f-47fa-ac40-4e8e054fdcd4",
        "versionNumber": "11.5.0",
        "minSupportedMendixVersion": "10.24.0",
        "publicationDate": "2026-01-13T06:57:14.512Z"
      }
    }
  ]
}`

const sampleContent = `{
  "contentId": 170,
  "publisher": "Mendix",
  "type": "Module",
  "categories": [{"name": "Utility"}],
  "supportCategory": "Platform",
  "licenseUrl": "http://www.apache.org/licenses/LICENSE-2.0.html",
  "isPrivate": false,
  "latestVersion": {
    "name": "Community Commons",
    "versionId": "0a03e65a-d94f-47fa-ac40-4e8e054fdcd4",
    "versionNumber": "11.5.0",
    "minSupportedMendixVersion": "10.24.0",
    "publicationDate": "2026-01-13T06:57:14.512Z"
  }
}`

const sampleVersions = `{
  "items": [
    {
      "name": "Community Commons",
      "versionId": "0a03e65a-d94f-47fa-ac40-4e8e054fdcd4",
      "versionNumber": "11.5.0",
      "minSupportedMendixVersion": "10.24.0",
      "publicationDate": "2026-01-13T06:57:14.512Z",
      "releaseNotes": "<p>We upgraded guava to 33.5.0-jre</p>"
    }
  ]
}`

func newMockServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return NewWithBaseURL(ts.Client(), ts.URL), ts
}

func TestSearch_PassesQueryAndLimit(t *testing.T) {
	var gotPath, gotQuery string
	client, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleContentList))
	})

	result, err := client.Search(context.Background(), "database", 3)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if gotPath != "/v1/content" {
		t.Errorf("path: got %q, want /v1/content", gotPath)
	}
	if !strings.Contains(gotQuery, "search=database") {
		t.Errorf("query missing search: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=3") {
		t.Errorf("query missing limit: %q", gotQuery)
	}
	if len(result.Items) != 1 || result.Items[0].ContentID != 170 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestSearch_NoQueryOrLimit(t *testing.T) {
	var gotQuery string
	client, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(sampleContentList))
	})

	if _, err := client.Search(context.Background(), "", 0); err != nil {
		t.Fatal(err)
	}
	if gotQuery != "" {
		t.Errorf("expected empty query when no search or limit, got %q", gotQuery)
	}
}

func TestGet_ParsesContentDetail(t *testing.T) {
	client, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/content/170" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(sampleContent))
	})

	got, err := client.Get(context.Background(), 170)
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentID != 170 || got.Publisher != "Mendix" {
		t.Errorf("unexpected content: %+v", got)
	}
	if got.LatestVersion == nil || got.LatestVersion.VersionNumber != "11.5.0" {
		t.Errorf("latestVersion not parsed: %+v", got.LatestVersion)
	}
	if len(got.Categories) != 1 || got.Categories[0].Name != "Utility" {
		t.Errorf("categories not parsed: %+v", got.Categories)
	}
}

func TestVersions_ParsesList(t *testing.T) {
	client, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/content/170/versions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(sampleVersions))
	})

	got, err := client.Versions(context.Background(), 170)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 version, got %d", len(got.Items))
	}
	v := got.Items[0]
	if v.VersionNumber != "11.5.0" {
		t.Errorf("versionNumber: %q", v.VersionNumber)
	}
	if v.MinSupportedMendixVersion != "10.24.0" {
		t.Errorf("minSupportedMendixVersion: %q", v.MinSupportedMendixVersion)
	}
	if !strings.Contains(v.ReleaseNotes, "guava") {
		t.Errorf("releaseNotes: %q", v.ReleaseNotes)
	}
	if v.PublicationDate.IsZero() {
		t.Error("publicationDate did not parse")
	}
	expected := time.Date(2026, 1, 13, 6, 57, 14, 512000000, time.UTC)
	if !v.PublicationDate.Equal(expected) {
		t.Errorf("publicationDate: got %v, want %v", v.PublicationDate, expected)
	}
}

func TestGet_HTTPErrorIsReported(t *testing.T) {
	client, _ := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})

	_, err := client.Get(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status: %v", err)
	}
}

func TestGet_InvalidJSONReported(t *testing.T) {
	client, _ := newMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})

	_, err := client.Get(context.Background(), 170)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error should mention decode: %v", err)
	}
}

func TestNew_UsesDefaultBaseURL(t *testing.T) {
	c := New(http.DefaultClient)
	if c.baseURL != BaseURL {
		t.Errorf("expected default BaseURL %q, got %q", BaseURL, c.baseURL)
	}
	if c.baseURL != "https://marketplace-api.mendix.com" {
		t.Errorf("default BaseURL unexpected: %q", c.baseURL)
	}
}
