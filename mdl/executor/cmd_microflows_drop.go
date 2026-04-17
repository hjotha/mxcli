// SPDX-License-Identifier: Apache-2.0

// Package executor - DROP MICROFLOW command
package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
)

// execDropMicroflow handles DROP MICROFLOW statements.
func execDropMicroflow(ctx *ExecContext, s *ast.DropMicroflowStmt) error {
	e := ctx.executor
	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	// Get hierarchy for module/folder resolution
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Find and delete the microflow
	mfs, err := e.reader.ListMicroflows()
	if err != nil {
		return mdlerrors.NewBackend("list microflows", err)
	}

	for _, mf := range mfs {
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == s.Name.Module && mf.Name == s.Name.Name {
			if err := e.writer.DeleteMicroflow(mf.ID); err != nil {
				return mdlerrors.NewBackend("delete microflow", err)
			}
			// Clear executor-level caches so subsequent CREATE sees fresh state
			qualifiedName := s.Name.Module + "." + s.Name.Name
			if e.cache != nil && e.cache.createdMicroflows != nil {
				delete(e.cache.createdMicroflows, qualifiedName)
			}
			invalidateHierarchy(ctx)
			fmt.Fprintf(ctx.Output, "Dropped microflow: %s.%s\n", s.Name.Module, s.Name.Name)
			return nil
		}
	}

	return mdlerrors.NewNotFound("microflow", s.Name.Module+"."+s.Name.Name)
}
