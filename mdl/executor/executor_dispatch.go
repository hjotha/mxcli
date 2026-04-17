// SPDX-License-Identifier: Apache-2.0

package executor

import "github.com/mendixlabs/mxcli/mdl/ast"

// executeInner dispatches a statement to its registered handler.
func (e *Executor) executeInner(stmt ast.Statement) error {
	return e.registry.Dispatch(e, stmt)
}
