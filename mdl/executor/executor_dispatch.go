// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

// executeInner dispatches a statement to its registered handler.
func (e *Executor) executeInner(ctx context.Context, stmt ast.Statement) error {
	ectx := e.newExecContext(ctx)
	return e.registry.Dispatch(ectx, stmt)
}

// newExecContext builds an ExecContext from the current Executor state.
// During the migration period, Backend may be nil; handlers access the
// Executor via ctx.executor for operations not yet routed through Backend.
func (e *Executor) newExecContext(ctx context.Context) *ExecContext {
	return &ExecContext{
		Context:   ctx,
		Output:    e.output,
		Format:    e.format,
		Quiet:     e.quiet,
		Logger:    e.logger,
		Fragments: e.fragments,
		Catalog:   e.catalog,
		Cache:     e.cache,
		executor:  e,
	}
}
