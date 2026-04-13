// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
)

// listDataTransformers handles LIST DATA TRANSFORMERS [IN module].
func (e *Executor) listDataTransformers(moduleName string) error {
	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return fmt.Errorf("failed to list data transformers: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	var rows [][]any
	for _, dt := range transformers {
		modID := h.FindModuleID(dt.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && !strings.EqualFold(modName, moduleName) {
			continue
		}
		qn := modName + "." + dt.Name
		steps := ""
		for _, s := range dt.Steps {
			if steps != "" {
				steps += " → "
			}
			steps += s.Technology
		}
		rows = append(rows, []any{qn, modName, dt.Name, dt.SourceType, steps})
	}

	if len(rows) == 0 {
		fmt.Fprintln(e.output, "No data transformers found.")
		return nil
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Source", "Steps"},
		Rows:    rows,
		Summary: fmt.Sprintf("(%d data transformers)", len(rows)),
	}
	return e.writeResult(result)
}

// describeDataTransformer handles DESCRIBE DATA TRANSFORMER Module.Name.
func (e *Executor) describeDataTransformer(name ast.QualifiedName) error {
	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return fmt.Errorf("failed to list data transformers: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, dt := range transformers {
		modID := h.FindModuleID(dt.ContainerID)
		modName := h.GetModuleName(modID)
		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(dt.Name, name.Name) {
			continue
		}

		w := e.output

		// Emit re-executable MDL
		fmt.Fprintf(w, "CREATE DATA TRANSFORMER %s.%s\n", modName, dt.Name)

		// Source — collapse newlines into spaces for single-line string
		sourceContent := strings.ReplaceAll(dt.SourceJSON, "\n", " ")
		sourceContent = strings.ReplaceAll(sourceContent, "'", "''")
		fmt.Fprintf(w, "SOURCE %s '%s'\n", dt.SourceType, sourceContent)
		fmt.Fprintln(w, "{")

		for _, step := range dt.Steps {
			if strings.Contains(step.Expression, "\n") {
				// Multi-line: use $$ quoting
				fmt.Fprintf(w, "  %s $$\n%s\n  $$;\n", step.Technology, step.Expression)
			} else {
				// Single-line: use regular string
				expr := strings.ReplaceAll(step.Expression, "'", "''")
				fmt.Fprintf(w, "  %s '%s';\n", step.Technology, expr)
			}
		}

		fmt.Fprintln(w, "};")
		return nil
	}

	return fmt.Errorf("data transformer not found: %s.%s", name.Module, name.Name)
}

// execCreateDataTransformer creates a new data transformer.
func (e *Executor) execCreateDataTransformer(s *ast.CreateDataTransformerStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	if err := e.checkFeature("integration", "data_transformer",
		"CREATE DATA TRANSFORMER",
		"upgrade your project to 11.9+"); err != nil {
		return err
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return fmt.Errorf("module %s not found", s.Name.Module)
	}

	dt := &model.DataTransformer{
		ContainerID: module.ID,
		Name:        s.Name.Name,
		SourceType:  s.SourceType,
		SourceJSON:  s.SourceJSON,
	}

	for _, step := range s.Steps {
		dt.Steps = append(dt.Steps, &model.DataTransformerStep{
			Technology: step.Technology,
			Expression: step.Expression,
		})
	}

	if err := e.writer.CreateDataTransformer(dt); err != nil {
		return fmt.Errorf("failed to create data transformer: %w", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created data transformer: %s.%s (%d steps)\n",
			s.Name.Module, s.Name.Name, len(dt.Steps))
	}
	return nil
}

// execDropDataTransformer deletes a data transformer.
func (e *Executor) execDropDataTransformer(s *ast.DropDataTransformerStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return fmt.Errorf("failed to list data transformers: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	for _, dt := range transformers {
		modID := h.FindModuleID(dt.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == s.Name.Module && dt.Name == s.Name.Name {
			if err := e.writer.DeleteDataTransformer(dt.ID); err != nil {
				return fmt.Errorf("failed to drop data transformer: %w", err)
			}
			if !e.quiet {
				fmt.Fprintf(e.output, "Dropped data transformer: %s.%s\n", s.Name.Module, s.Name.Name)
			}
			return nil
		}
	}

	return fmt.Errorf("data transformer %s.%s not found", s.Name.Module, s.Name.Name)
}
