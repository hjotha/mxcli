// SPDX-License-Identifier: Apache-2.0

// Package executor - DROP/MOVE FOLDER commands
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// findFolderByPath walks a folder path under a module and returns the folder ID.
func (e *Executor) findFolderByPath(moduleID model.ID, folderPath string, folders []*mpr.FolderInfo) (model.ID, error) {
	parts := strings.Split(folderPath, "/")
	currentContainerID := moduleID

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
			return "", fmt.Errorf("folder not found: '%s'", folderPath)
		}
	}

	if targetFolderID == "" {
		return "", fmt.Errorf("folder not found: '%s'", folderPath)
	}

	return targetFolderID, nil
}

// execDropFolder handles DROP FOLDER 'path' IN Module statements.
// The folder must be empty (no child documents or sub-folders).
func (e *Executor) execDropFolder(s *ast.DropFolderStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project")
	}

	module, err := e.findModule(s.Module)
	if err != nil {
		return fmt.Errorf("module not found: %s", s.Module)
	}

	folders, err := e.reader.ListFolders()
	if err != nil {
		return fmt.Errorf("failed to list folders: %w", err)
	}

	folderID, err := e.findFolderByPath(module.ID, s.FolderPath, folders)
	if err != nil {
		return fmt.Errorf("%w in %s", err, s.Module)
	}

	if err := e.writer.DeleteFolder(folderID); err != nil {
		return fmt.Errorf("failed to delete folder '%s': %w", s.FolderPath, err)
	}

	e.invalidateHierarchy()
	fmt.Fprintf(e.output, "Dropped folder: '%s' in %s\n", s.FolderPath, s.Module)
	return nil
}

// execMoveFolder handles MOVE FOLDER Module.FolderName TO ... statements.
func (e *Executor) execMoveFolder(s *ast.MoveFolderStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Find the source module
	sourceModule, err := e.findModule(s.Name.Module)
	if err != nil {
		return fmt.Errorf("source module not found: %s", s.Name.Module)
	}

	// Find the source folder
	folders, err := e.reader.ListFolders()
	if err != nil {
		return fmt.Errorf("failed to list folders: %w", err)
	}

	folderID, err := e.findFolderByPath(sourceModule.ID, s.Name.Name, folders)
	if err != nil {
		return fmt.Errorf("%w in %s", err, s.Name.Module)
	}

	// Determine target module
	var targetModule *model.Module
	if s.TargetModule != "" {
		targetModule, err = e.findModule(s.TargetModule)
		if err != nil {
			return fmt.Errorf("target module not found: %s", s.TargetModule)
		}
	} else {
		targetModule = sourceModule
	}

	// Resolve target container
	var targetContainerID model.ID
	if s.TargetFolder != "" {
		targetContainerID, err = e.resolveFolder(targetModule.ID, s.TargetFolder)
		if err != nil {
			return fmt.Errorf("failed to resolve target folder: %w", err)
		}
	} else {
		targetContainerID = targetModule.ID
	}

	// Move the folder
	if err := e.writer.MoveFolder(folderID, targetContainerID); err != nil {
		return fmt.Errorf("failed to move folder: %w", err)
	}

	e.invalidateHierarchy()

	target := targetModule.Name
	if s.TargetFolder != "" {
		target += "/" + s.TargetFolder
	}
	fmt.Fprintf(e.output, "Moved folder %s to %s\n", s.Name.String(), target)
	return nil
}
