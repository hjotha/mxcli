// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"io"
	"path/filepath"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend"
	"github.com/mendixlabs/mxcli/mdl/catalog"
	"github.com/mendixlabs/mxcli/mdl/diaglog"
	sqllib "github.com/mendixlabs/mxcli/sql"
)

// ExecContext carries all dependencies a statement handler needs.
//
// Design notes:
//   - Embeds context.Context for cancellation and timeout propagation.
//   - Holds a FullBackend for domain operations (handlers narrow to
//     the sub-interface they need via type assertion or accessor).
//   - Ancillary fields (output, format, cache, etc.) are lifted from
//     the Executor struct so handlers don't depend on *Executor.
type ExecContext struct {
	context.Context

	// Backend provides all domain operations (read/write/connect).
	// Nil when not connected.
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

	// MprPath is the filesystem path to the connected .mpr file.
	// Empty when not connected.
	MprPath string

	// SqlMgr manages external SQL database connections (lazy init).
	SqlMgr *sqllib.Manager

	// ThemeRegistry holds cached theme design property definitions (lazy init).
	ThemeRegistry *ThemeRegistry

	// Settings holds session-scoped key-value settings (SET command).
	Settings map[string]any

	// executor is a temporary back-reference used during incremental migration.
	// Handlers that have not yet been migrated to use Backend can access the
	// original Executor through this field. Remove once all handlers are fully
	// migrated to ctx.Backend (tracked: ~27 e := ctx.executor sites remain).
	executor *Executor
}

// Connected returns true if a project is connected via the Backend.
func (ctx *ExecContext) Connected() bool {
	return ctx.Backend != nil && ctx.Backend.IsConnected()
}

// ConnectedForWrite returns true if a project is connected and the backend
// supports write operations. Currently equivalent to Connected() since
// MprBackend always supports writes.
func (ctx *ExecContext) ConnectedForWrite() bool {
	return ctx.Connected()
}

// InvalidateCache clears the hierarchy/entity/microflow cache on both the
// ExecContext and the underlying Executor so that subsequent statements see
// fresh data.
func (ctx *ExecContext) InvalidateCache() {
	ctx.Cache = nil
	if ctx.executor != nil {
		ctx.executor.cache = nil
	}
}

// GetThemeRegistry returns the cached theme registry, loading it lazily
// from the project's theme sources on first access.
func (ctx *ExecContext) GetThemeRegistry() *ThemeRegistry {
	if ctx.ThemeRegistry != nil {
		return ctx.ThemeRegistry
	}
	if ctx.MprPath == "" {
		return nil
	}
	projectDir := filepath.Dir(ctx.MprPath)
	registry, err := loadThemeRegistry(projectDir)
	if err != nil {
		if ctx.Logger != nil {
			ctx.Logger.Warn("failed to load theme registry", "error", err)
		}
		return nil
	}
	ctx.ThemeRegistry = registry
	// Persist back to Executor so subsequent statements pick up the cached value
	// via newExecContext without reloading.
	if ctx.executor != nil {
		ctx.executor.themeRegistry = registry
	}
	return ctx.ThemeRegistry
}
