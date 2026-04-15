// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"net/url"
	"path/filepath"
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
