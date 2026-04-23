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
	e.syncBack(ectx)
	return err
}

// syncBack copies mutated ExecContext fields back to the Executor so that
// the next newExecContext call picks up handler-side state changes.
// This replaces the scattered ctx.executor.field = ctx.Field write-backs
// that previously lived inside individual handlers.
func (e *Executor) syncBack(ctx *ExecContext) {
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
		Context:       ctx,
		Backend:       e.backend,
		Output:        e.output,
		Format:        e.format,
		Quiet:         e.quiet,
		Logger:        e.logger,
		Fragments:     e.fragments,
		Catalog:       e.catalog,
		Cache:         e.cache,
		MprPath:       e.mprPath,
		SqlMgr:        e.sqlMgr,
		ThemeRegistry: e.themeRegistry,
		Settings:      e.settings,
		executor:      e,
	}
}

// Ensure ast import is used via executeInner's stmt parameter.
var _ ast.Statement = (*ast.HelpStmt)(nil)
