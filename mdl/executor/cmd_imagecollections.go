// SPDX-License-Identifier: Apache-2.0

// Package executor - Image collection commands (CREATE/DROP IMAGE COLLECTION)
package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// execCreateImageCollection handles CREATE IMAGE COLLECTION statements.
func (e *Executor) execCreateImageCollection(s *ast.CreateImageCollectionStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Find or auto-create module
	module, err := e.findOrCreateModule(s.Name.Module)
	if err != nil {
		return err
	}

	// Check if image collection already exists
	existing := e.findImageCollection(s.Name.Module, s.Name.Name)
	if existing != nil {
		return fmt.Errorf("image collection already exists: %s.%s", s.Name.Module, s.Name.Name)
	}

	// Build ImageCollection
	ic := &mpr.ImageCollection{
		ContainerID:   module.ID,
		Name:          s.Name.Name,
		ExportLevel:   s.ExportLevel,
		Documentation: s.Comment,
	}

	// Load image files
	for _, item := range s.Images {
		filePath := item.FilePath
		if !filepath.IsAbs(filePath) {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read image file %q: %w", item.FilePath, err)
		}
		format := extToImageFormat(filepath.Ext(filePath))
		ic.Images = append(ic.Images, mpr.Image{
			Name:   item.Name,
			Data:   data,
			Format: format,
		})
	}

	if err := e.writer.CreateImageCollection(ic); err != nil {
		return fmt.Errorf("failed to create image collection: %w", err)
	}

	// Invalidate hierarchy cache so the new collection's container is visible
	e.invalidateHierarchy()

	fmt.Fprintf(e.output, "Created image collection: %s\n", s.Name)
	return nil
}

// execDropImageCollection handles DROP IMAGE COLLECTION statements.
func (e *Executor) execDropImageCollection(s *ast.DropImageCollectionStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	ic := e.findImageCollection(s.Name.Module, s.Name.Name)
	if ic == nil {
		return fmt.Errorf("image collection not found: %s", s.Name)
	}

	if err := e.writer.DeleteImageCollection(string(ic.ID)); err != nil {
		return fmt.Errorf("failed to delete image collection: %w", err)
	}

	fmt.Fprintf(e.output, "Dropped image collection: %s\n", s.Name)
	return nil
}

// describeImageCollection handles DESCRIBE IMAGE COLLECTION Module.Name.
func (e *Executor) describeImageCollection(name ast.QualifiedName) error {
	ic := e.findImageCollection(name.Module, name.Name)
	if ic == nil {
		return fmt.Errorf("image collection not found: %s", name)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(ic.ContainerID)
	modName := h.GetModuleName(modID)

	if ic.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", ic.Documentation)
	}

	exportLevel := ic.ExportLevel
	if exportLevel == "" {
		exportLevel = "Hidden"
	}

	qualifiedName := fmt.Sprintf("%s.%s", modName, ic.Name)

	if len(ic.Images) == 0 {
		fmt.Fprintf(e.output, "CREATE IMAGE COLLECTION %s", qualifiedName)
		if exportLevel != "Hidden" {
			fmt.Fprintf(e.output, " EXPORT LEVEL '%s'", exportLevel)
		}
		fmt.Fprintln(e.output, ";")
		fmt.Fprintln(e.output, "/")
		return nil
	}

	// Write image data to temp files and output CREATE statement with IMAGE lines
	previewDir := filepath.Join("/tmp/mxcli-preview", qualifiedName)
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	fmt.Fprintf(e.output, "CREATE IMAGE COLLECTION %s", qualifiedName)
	if exportLevel != "Hidden" {
		fmt.Fprintf(e.output, " EXPORT LEVEL '%s'", exportLevel)
	}
	fmt.Fprintln(e.output, " (")

	for i, img := range ic.Images {
		ext := imageFormatToExt(img.Format)
		filePath := filepath.Join(previewDir, img.Name+ext)
		if len(img.Data) > 0 {
			if err := os.WriteFile(filePath, img.Data, 0o644); err != nil {
				return fmt.Errorf("failed to write image %s: %w", img.Name, err)
			}
		}

		comma := ","
		if i == len(ic.Images)-1 {
			comma = ""
		}
		fmt.Fprintf(e.output, "    IMAGE %s FROM FILE '%s'%s\n", img.Name, filePath, comma)
	}

	fmt.Fprintln(e.output, ");")
	fmt.Fprintln(e.output, "/")
	return nil
}

// imageFormatToExt converts a Mendix ImageFormat value to a file extension.
func imageFormatToExt(format string) string {
	switch format {
	case "Svg":
		return ".svg"
	case "Gif":
		return ".gif"
	case "Jpg":
		return ".jpg"
	case "Bmp":
		return ".bmp"
	case "Webp":
		return ".webp"
	default:
		return ".png"
	}
}

// extToImageFormat converts a file extension to a Mendix ImageFormat value.
func extToImageFormat(ext string) string {
	switch strings.ToLower(ext) {
	case ".svg":
		return "Svg"
	case ".gif":
		return "Gif"
	case ".jpg", ".jpeg":
		return "Jpg"
	case ".bmp":
		return "Bmp"
	case ".webp":
		return "Webp"
	default:
		return "Png"
	}
}

// showImageCollections handles SHOW IMAGE COLLECTION [IN module].
func (e *Executor) showImageCollections(moduleName string) error {
	collections, err := e.reader.ListImageCollections()
	if err != nil {
		return fmt.Errorf("failed to list image collections: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	fmt.Fprintf(e.output, "| %-40s | %-12s | %-6s |\n", "Image Collection", "Export Level", "Images")
	fmt.Fprintf(e.output, "|%-42s|%-14s|%-8s|\n", strings.Repeat("-", 42), strings.Repeat("-", 14), strings.Repeat("-", 8))

	count := 0
	for _, ic := range collections {
		modID := h.FindModuleID(ic.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}

		qualifiedName := fmt.Sprintf("%s.%s", modName, ic.Name)
		exportLevel := ic.ExportLevel
		if exportLevel == "" {
			exportLevel = "Hidden"
		}
		fmt.Fprintf(e.output, "| %-40s | %-12s | %6d |\n", qualifiedName, exportLevel, len(ic.Images))
		count++
	}

	fmt.Fprintf(e.output, "\n(%d image collection(s))\n", count)
	return nil
}

// findImageCollection finds an image collection by module and name.
func (e *Executor) findImageCollection(moduleName, collectionName string) *mpr.ImageCollection {
	collections, err := e.reader.ListImageCollections()
	if err != nil {
		return nil
	}

	h, err := e.getHierarchy()
	if err != nil {
		return nil
	}

	for _, ic := range collections {
		modID := h.FindModuleID(ic.ContainerID)
		modName := h.GetModuleName(modID)
		if ic.Name == collectionName && modName == moduleName {
			return ic
		}
	}
	return nil
}
