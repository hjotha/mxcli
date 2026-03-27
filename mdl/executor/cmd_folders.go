// SPDX-License-Identifier: Apache-2.0

// Package executor - DROP FOLDER command
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
)

// execDropFolder handles DROP FOLDER 'path' IN Module statements.
// The folder must be empty (no child documents or sub-folders).
func (e *Executor) execDropFolder(s *ast.DropFolderStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Find the module
	module, err := e.findModule(s.Module)
	if err != nil {
		return fmt.Errorf("module not found: %s", s.Module)
	}

	// List all folders
	folders, err := e.reader.ListFolders()
	if err != nil {
		return fmt.Errorf("failed to list folders: %w", err)
	}

	// Walk the folder path to find the target folder
	parts := strings.Split(s.FolderPath, "/")
	currentContainerID := module.ID

	var targetFolderID model.ID
	for i, part := range parts {
		if part == "" {
			continue
		}

		var found bool
		for _, f := range folders {
			if f.ContainerID == currentContainerID && f.Name == part {
				currentContainerID = f.ID
				if i == len(parts)-1 {
					targetFolderID = f.ID
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("folder not found: '%s' in %s", s.FolderPath, s.Module)
		}
	}

	if targetFolderID == "" {
		return fmt.Errorf("folder not found: '%s' in %s", s.FolderPath, s.Module)
	}

	// Delete the folder (writer checks if empty)
	if err := e.writer.DeleteFolder(targetFolderID); err != nil {
		return fmt.Errorf("failed to delete folder '%s': %w", s.FolderPath, err)
	}

	e.invalidateHierarchy()
	fmt.Fprintf(e.output, "Dropped folder: '%s' in %s\n", s.FolderPath, s.Module)
	return nil
}
