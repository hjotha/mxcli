// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
)

// listDataTransformers handles LIST DATA TRANSFORMERS [IN module].
func listDataTransformers(ctx *ExecContext, moduleName string) error {
	e := ctx.executor

	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return mdlerrors.NewBackend("list data transformers", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
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
		fmt.Fprintln(ctx.Output, "No data transformers found.")
		return nil
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Source", "Steps"},
		Rows:    rows,
		Summary: fmt.Sprintf("(%d data transformers)", len(rows)),
	}
	return writeResult(ctx, result)
}

// describeDataTransformer handles DESCRIBE DATA TRANSFORMER Module.Name.
func describeDataTransformer(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor

	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return mdlerrors.NewBackend("list data transformers", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	for _, dt := range transformers {
		modID := h.FindModuleID(dt.ContainerID)
		modName := h.GetModuleName(modID)
		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(dt.Name, name.Name) {
			continue
		}

		w := ctx.Output

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

	return mdlerrors.NewNotFound("data transformer", name.Module+"."+name.Name)
}

// execCreateDataTransformer creates a new data transformer.
func execCreateDataTransformer(ctx *ExecContext, s *ast.CreateDataTransformerStmt) error {
	e := ctx.executor

	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	if err := checkFeature(ctx, "integration", "data_transformer",
		"CREATE DATA TRANSFORMER",
		"upgrade your project to 11.9+"); err != nil {
		return err
	}

	module, err := findModule(ctx, s.Name.Module)
	if err != nil {
		return mdlerrors.NewNotFound("module", s.Name.Module)
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
		return mdlerrors.NewBackend("create data transformer", err)
	}

	if !e.quiet {
		fmt.Fprintf(ctx.Output, "Created data transformer: %s.%s (%d steps)\n",
			s.Name.Module, s.Name.Name, len(dt.Steps))
	}
	return nil
}

// execDropDataTransformer deletes a data transformer.
func execDropDataTransformer(ctx *ExecContext, s *ast.DropDataTransformerStmt) error {
	e := ctx.executor

	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	transformers, err := e.reader.ListDataTransformers()
	if err != nil {
		return mdlerrors.NewBackend("list data transformers", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return err
	}

	for _, dt := range transformers {
		modID := h.FindModuleID(dt.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == s.Name.Module && dt.Name == s.Name.Name {
			if err := e.writer.DeleteDataTransformer(dt.ID); err != nil {
				return mdlerrors.NewBackend("drop data transformer", err)
			}
			if !e.quiet {
				fmt.Fprintf(ctx.Output, "Dropped data transformer: %s.%s\n", s.Name.Module, s.Name.Name)
			}
			return nil
		}
	}

	return mdlerrors.NewNotFound("data transformer", s.Name.Module+"."+s.Name.Name)
}

// Executor wrappers for unmigrated callers.
