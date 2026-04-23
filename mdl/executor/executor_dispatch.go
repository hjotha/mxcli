// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

// executeInner dispatches a statement to its registered handler.
func (e *Executor) executeInner(ctx context.Context, stmt ast.Statement) error {
	ectx := e.newExecContext(ctx)
	err := e.registry.Dispatch(ectx, stmt)
	// Only sync back when the context has not been cancelled. Execute() runs
	// executeInner in a goroutine with a wall-clock timeout; if the timeout
	// fires, the goroutine keeps running but Execute() has already returned.
	// Syncing stale state back at that point would race with subsequent calls.
	if ctx.Err() == nil {
		e.syncBack(ectx)
	}
	return err
}

// syncBack copies mutated ExecContext fields back to the Executor so that
// the next newExecContext call picks up handler-side state changes.
//
// Fields intentionally NOT synced back (read-only from handler perspective):
//   - Output, Format, Quiet, Logger — set once at Executor construction
//   - BackendFactory — set once at Executor construction
//   - OutputGuard — removed; writeDescribeJSON captures via Output swap only
//   - ExecuteFn, ExecuteProgramFn, FinalizeFn — bound to Executor methods, immutable
func (e *Executor) syncBack(ctx *ExecContext) {
	e.backend = ctx.Backend
	e.mprPath = ctx.MprPath
	e.cache = ctx.Cache
	e.catalog = ctx.Catalog
	e.settings = ctx.Settings
	e.fragments = ctx.Fragments
	e.sqlMgr = ctx.SqlMgr
	e.themeRegistry = ctx.ThemeRegistry
}

// newExecContext builds an ExecContext from the current Executor state.
func (e *Executor) newExecContext(ctx context.Context) *ExecContext {
	return &ExecContext{
		Context:          ctx,
		Backend:          e.backend,
		Output:           e.output,
		Format:           e.format,
		Quiet:            e.quiet,
		Logger:           e.logger,
		Fragments:        e.fragments,
		Catalog:          e.catalog,
		Cache:            e.cache,
		MprPath:          e.mprPath,
		SqlMgr:           e.sqlMgr,
		ThemeRegistry:    e.themeRegistry,
		Settings:         e.settings,
		BackendFactory:   e.backendFactory,
		ExecuteFn:        e.Execute,
		ExecuteProgramFn: e.ExecuteProgram,
		FinalizeFn:       e.finalizeProgramExecution,
	}
}

// Ensure ast import is used via executeInner's stmt parameter.
var _ ast.Statement = (*ast.HelpStmt)(nil)
