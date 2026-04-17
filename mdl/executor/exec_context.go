// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"io"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend"
	"github.com/mendixlabs/mxcli/mdl/catalog"
	"github.com/mendixlabs/mxcli/mdl/diaglog"
)

// ExecContext carries all dependencies a statement handler needs.
// It will replace the direct *Executor receiver once handlers are migrated.
//
// Design notes:
//   - Embeds context.Context for cancellation and timeout propagation.
//   - Holds a FullBackend for domain operations (handlers narrow to
//     the sub-interface they need via type assertion or accessor).
//   - Ancillary fields (output, format, cache, etc.) are lifted from
//     the Executor struct so handlers don't depend on *Executor.
type ExecContext struct {
	context.Context

	// Backend provides all domain operations.
	Backend backend.FullBackend

	// Output is the writer for user-visible output (with line-limit guard).
	Output io.Writer

	// Format controls output formatting (table, json, etc.).
	Format OutputFormat

	// Quiet suppresses connection and status messages.
	Quiet bool

	// Logger is the session diagnostics logger (nil = no logging).
	Logger *diaglog.Logger

	// Fragments holds script-scoped fragment definitions.
	Fragments map[string]*ast.DefineFragmentStmt

	// Catalog provides MDL name resolution.
	Catalog *catalog.Catalog

	// Cache holds per-session cached data for performance.
	Cache *executorCache

	// executor is a temporary back-reference used during incremental migration.
	// Handlers that have not yet been migrated to use Backend can access the
	// original Executor through this field. It will be removed once all handlers
	// are migrated.
	executor *Executor
}
