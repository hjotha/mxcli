// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"runtime"
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
