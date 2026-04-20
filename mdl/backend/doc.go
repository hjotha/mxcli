// SPDX-License-Identifier: Apache-2.0

// Package backend defines domain-specific interfaces that decouple the
// executor from concrete storage (e.g. .mpr files). Each interface
// groups related read/write operations by domain concept.
//
// Shared value types live in mdl/types to keep this package free of
// sdk/mpr dependencies. Conversion between types.* and sdk/mpr.*
// structs is handled inside mdl/backend/mpr (MprBackend).
package backend
