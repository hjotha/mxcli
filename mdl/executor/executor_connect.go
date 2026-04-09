// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

func (e *Executor) execConnect(s *ast.ConnectStmt) error {
	if e.writer != nil {
		e.writer.Close()
	}

	writer, err := mpr.NewWriter(s.Path)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	e.writer = writer
	e.reader = writer.Reader()
	e.mprPath = s.Path
	e.cache = &executorCache{} // Initialize fresh cache

	// Display connection info with version
	pv := e.reader.ProjectVersion()
	if !e.quiet {
		fmt.Fprintf(e.output, "Connected to: %s (Mendix %s)\n", s.Path, pv.ProductVersion)
	}
	if e.logger != nil {
		e.logger.Connect(s.Path, pv.ProductVersion, pv.FormatVersion)
	}
	return nil
}

// reconnect closes the current connection and reopens it.
// This is needed when the project file has been modified externally.
func (e *Executor) reconnect() error {
	if e.mprPath == "" {
		return fmt.Errorf("no project path to reconnect to")
	}

	// Close existing connection
	if e.writer != nil {
		e.writer.Close()
	}

	// Reopen connection
	writer, err := mpr.NewWriter(e.mprPath)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	e.writer = writer
	e.reader = writer.Reader()
	e.cache = &executorCache{} // Reset cache
	return nil
}

func (e *Executor) execDisconnect() error {
	if e.writer == nil {
		fmt.Fprintln(e.output, "Not connected")
		return nil
	}

	// Reconcile any pending security changes before closing
	if err := e.finalizeProgramExecution(); err != nil {
		fmt.Fprintf(e.output, "Warning: finalization error: %v\n", err)
	}

	e.writer.Close()
	fmt.Fprintf(e.output, "Disconnected from: %s\n", e.mprPath)
	e.writer = nil
	e.reader = nil
	e.mprPath = ""
	e.cache = nil
	return nil
}

func (e *Executor) execStatus() error {
	if e.writer == nil {
		fmt.Fprintln(e.output, "Status: Not connected")
		return nil
	}

	pv := e.reader.ProjectVersion()
	fmt.Fprintf(e.output, "Status: Connected\n")
	fmt.Fprintf(e.output, "Project: %s\n", e.mprPath)
	fmt.Fprintf(e.output, "Mendix Version: %s\n", pv.ProductVersion)
	fmt.Fprintf(e.output, "MPR Format: v%d\n", pv.FormatVersion)

	// Show module count
	modules, err := e.reader.ListModules()
	if err == nil {
		fmt.Fprintf(e.output, "Modules: %d\n", len(modules))
	}

	return nil
}
