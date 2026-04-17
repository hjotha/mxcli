// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

func execConnect(ctx *ExecContext, s *ast.ConnectStmt) error {
	e := ctx.executor
	if e.writer != nil {
		e.writer.Close()
	}

	writer, err := mpr.NewWriter(s.Path)
	if err != nil {
		return mdlerrors.NewBackend("connect", err)
	}

	e.writer = writer
	e.reader = writer.Reader()
	e.mprPath = s.Path
	e.cache = &executorCache{} // Initialize fresh cache

	// Display connection info with version
	pv := e.reader.ProjectVersion()
	if !ctx.Quiet {
		fmt.Fprintf(ctx.Output, "Connected to: %s (Mendix %s)\n", s.Path, pv.ProductVersion)
	}
	if ctx.Logger != nil {
		ctx.Logger.Connect(s.Path, pv.ProductVersion, pv.FormatVersion)
	}
	return nil
}

// reconnect closes the current connection and reopens it.
// This is needed when the project file has been modified externally.
func reconnect(ctx *ExecContext) error {
	e := ctx.executor
	if e.mprPath == "" {
		return mdlerrors.NewNotConnected()
	}

	// Close existing connection
	if e.writer != nil {
		e.writer.Close()
	}

	// Reopen connection
	writer, err := mpr.NewWriter(e.mprPath)
	if err != nil {
		return mdlerrors.NewBackend("reconnect", err)
	}

	e.writer = writer
	e.reader = writer.Reader()
	e.cache = &executorCache{} // Reset cache
	return nil
}

func execDisconnect(ctx *ExecContext) error {
	e := ctx.executor
	if e.writer == nil {
		fmt.Fprintln(ctx.Output, "Not connected")
		return nil
	}

	// Reconcile any pending security changes before closing
	if err := e.finalizeProgramExecution(); err != nil {
		fmt.Fprintf(ctx.Output, "Warning: finalization error: %v\n", err)
	}

	e.writer.Close()
	fmt.Fprintf(ctx.Output, "Disconnected from: %s\n", e.mprPath)
	e.writer = nil
	e.reader = nil
	e.mprPath = ""
	e.cache = nil
	return nil
}

// Executor method wrappers — kept during migration for callers not yet
// converted to free functions. Remove once all callers are migrated.

func execStatus(ctx *ExecContext) error {
	e := ctx.executor
	if e.writer == nil {
		fmt.Fprintln(ctx.Output, "Status: Not connected")
		return nil
	}

	pv := e.reader.ProjectVersion()
	fmt.Fprintf(ctx.Output, "Status: Connected\n")
	fmt.Fprintf(ctx.Output, "Project: %s\n", e.mprPath)
	fmt.Fprintf(ctx.Output, "Mendix Version: %s\n", pv.ProductVersion)
	fmt.Fprintf(ctx.Output, "MPR Format: v%d\n", pv.FormatVersion)

	// Show module count
	modules, err := e.reader.ListModules()
	if err == nil {
		fmt.Fprintf(ctx.Output, "Modules: %d\n", len(modules))
	}

	return nil
}
