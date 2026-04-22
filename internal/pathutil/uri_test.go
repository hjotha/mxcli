// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestURIToPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "file URI with absolute path",
			input:    "file:///home/user/file.txt",
			expected: "/home/user/file.txt",
		},
		{
			name:     "raw absolute path",
			input:    "/home/user/file.txt",
			expected: "/home/user/file.txt",
		},
		{
			name:     "raw relative path",
			input:    "./metadata/file.xml",
			expected: "./metadata/file.xml",
		},
		{
			name:     "http URL returns unchanged",
			input:    "https://example.com/metadata",
			expected: "https://example.com/metadata",
		},
		{
			name:     "invalid URI returns empty",
			input:    "://invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URIToPath(tt.input)

			// Skip path separator tests on Windows
			if runtime.GOOS != "windows" {
				if result != tt.expected {
					t.Errorf("URIToPath(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestURIToPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows file URI",
			input:    "file:///C:/Users/test/file.txt",
			expected: "C:\\Users\\test\\file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URIToPath(tt.input)
			if result != tt.expected {
				t.Errorf("URIToPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		baseDir    string
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "HTTP URL unchanged",
			input:      "https://api.example.com/$metadata",
			baseDir:    "",
			wantPrefix: "https://",
			wantErr:    false,
		},
		{
			name:       "HTTPS URL unchanged",
			input:      "http://localhost:8080/odata/$metadata",
			baseDir:    "",
			wantPrefix: "http://",
			wantErr:    false,
		},
		{
			name:       "Absolute file:// URL unchanged",
			input:      "file:///tmp/metadata.xml",
			baseDir:    "",
			wantPrefix: "file://",
			wantErr:    false,
		},
		{
			name:       "Relative path with ./ normalized",
			input:      "./metadata.xml",
			baseDir:    tmpDir,
			wantPrefix: "file://",
			wantErr:    false,
		},
		{
			name:       "Bare relative path normalized",
			input:      "metadata.xml",
			baseDir:    tmpDir,
			wantPrefix: "file://",
			wantErr:    false,
		},
		{
			name:       "Absolute path normalized to file://",
			input:      "/tmp/metadata.xml",
			baseDir:    "",
			wantPrefix: "file://",
			wantErr:    false,
		},
		{
			name:       "Subdirectory relative path",
			input:      "contracts/metadata.xml",
			baseDir:    tmpDir,
			wantPrefix: "file://",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input, tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("Result %q does not start with %q", result, tt.wantPrefix)
			}

			// Verify file:// URLs contain absolute paths
			if strings.HasPrefix(result, "file://") {
				path := strings.TrimPrefix(result, "file://")
				if !filepath.IsAbs(path) {
					t.Errorf("file:// URL contains relative path: %q", result)
				}
			}

			// Verify relative paths are resolved correctly
			if tt.baseDir != "" && !strings.HasPrefix(tt.input, "http") && !strings.HasPrefix(tt.input, "file://") {
				path := strings.TrimPrefix(result, "file://")
				if !strings.Contains(path, filepath.ToSlash(tmpDir)) {
					t.Errorf("Relative path not resolved against baseDir. Got: %q", result)
				}
			}
		})
	}
}

func TestPathFromURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "file:// URL extracts path",
			input:    "file:///tmp/metadata.xml",
			expected: "/tmp/metadata.xml",
		},
		{
			name:     "HTTP URL returns empty",
			input:    "https://api.example.com/$metadata",
			expected: "",
		},
		{
			name:     "HTTPS URL returns empty",
			input:    "http://localhost:8080/metadata",
			expected: "",
		},
		{
			name:     "bare path returns empty",
			input:    "/tmp/file.xml",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathFromURL(tt.input)
			if runtime.GOOS != "windows" && result != tt.expected {
				t.Errorf("PathFromURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
