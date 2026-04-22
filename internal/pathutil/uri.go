// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// URIToPath converts a file:// URI to a filesystem path.
// If the input is not a valid URI or has a scheme other than "file",
// returns the input unchanged (treating it as a raw path).
func URIToPath(rawURI string) string {
	u, err := url.Parse(rawURI)
	if err != nil {
		return ""
	}
	if u.Scheme == "file" {
		return filepath.FromSlash(u.Path)
	}
	// If no scheme, treat as a raw path
	return rawURI
}

// NormalizeURL converts relative paths to absolute file:// URLs, while preserving HTTP(S) URLs.
// This is useful for storing URLs in a way that external tools (like Mendix Studio Pro) can
// reliably distinguish between local files and HTTP endpoints.
//
// Supported input formats:
//   - https://... or http://... → returned as-is
//   - file:///abs/path → returned as-is
//   - ./path or path/file.xml → converted to file:///absolute/path
//
// If baseDir is provided, relative paths are resolved against it.
// Otherwise, they're resolved against the current working directory.
//
// Returns an error if the path cannot be resolved to an absolute path.
func NormalizeURL(rawURL string, baseDir string) (string, error) {
	if rawURL == "" {
		return "", nil
	}

	// HTTP(S) URLs are already normalized
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL, nil
	}

	// Extract file path from file:// URLs or use raw input
	filePath := rawURL
	if strings.HasPrefix(rawURL, "file://") {
		filePath = URIToPath(rawURL)
		if filePath == "" {
			return "", fmt.Errorf("invalid file:// URI: %s", rawURL)
		}
	}

	// Convert relative paths to absolute
	if !filepath.IsAbs(filePath) {
		if baseDir != "" {
			filePath = filepath.Join(baseDir, filePath)
		} else {
			// No base directory - use cwd
			cwd, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("failed to resolve relative path: %w", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}
	}

	// Convert to absolute path (clean up ./ and ../)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Return as file:// URL with forward slashes (cross-platform)
	// RFC 8089 requires three slashes: file:///path or file:///C:/path
	slashed := filepath.ToSlash(absPath)
	if !strings.HasPrefix(slashed, "/") {
		// Windows path like C:/Users/x needs leading slash: file:///C:/Users/x
		slashed = "/" + slashed
	}
	return "file://" + slashed, nil
}

// PathFromURL extracts a filesystem path from a URL, handling both file:// URLs and HTTP(S) URLs.
// For file:// URLs, returns the local filesystem path.
// For HTTP(S) URLs or other schemes, returns an empty string.
// This is the inverse of converting a path to a file:// URL.
func PathFromURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "file://") {
		return URIToPath(rawURL)
	}
	// Not a file:// URL
	return ""
}
