// SPDX-License-Identifier: Apache-2.0

// Package executor - Commands for agent-editor Model documents.
//
// Handles `SHOW MODELS [IN module]` and `DESCRIBE MODEL Module.Name`.
// The underlying BSON is a CustomBlobDocuments$CustomBlobDocument with
// CustomDocumentType = "agenteditor.model". See
// docs/11-proposals/PROPOSAL_agent_document_support.md for schema.
package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
)

// showAgentEditorModels handles SHOW MODELS [IN module].
func showAgentEditorModels(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	models, err := e.reader.ListAgentEditorModels()
	if err != nil {
		return mdlerrors.NewBackend("list models", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return err
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Provider", "Key Constant", "Display Name"},
	}

	for _, m := range models {
		modID := h.FindModuleID(m.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}

		keyConstant := ""
		if m.Key != nil {
			keyConstant = m.Key.QualifiedName
		}

		result.Rows = append(result.Rows, []any{
			fmt.Sprintf("%s.%s", modName, m.Name),
			modName,
			m.Name,
			m.Provider,
			keyConstant,
			m.DisplayName,
		})
	}

	result.Summary = fmt.Sprintf("(%d model(s))", len(result.Rows))
	return writeResult(ctx, result)
}

// describeAgentEditorModel handles DESCRIBE MODEL Module.Name.
// Emits a round-trippable CREATE MODEL statement.
func describeAgentEditorModel(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	m := findAgentEditorModel(ctx, name.Module, name.Name)
	if m == nil {
		return mdlerrors.NewNotFound("model", name.String())
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return err
	}
	modID := h.FindModuleID(m.ContainerID)
	modName := h.GetModuleName(modID)
	qualifiedName := fmt.Sprintf("%s.%s", modName, m.Name)

	if m.Documentation != "" {
		fmt.Fprintf(ctx.Output, "/**\n * %s\n */\n", m.Documentation)
	}

	fmt.Fprintf(ctx.Output, "CREATE MODEL %s (\n", qualifiedName)

	// Emit properties in stable order. User-set properties (Provider, Key)
	// come first; Portal-populated metadata comes last and only if non-empty.
	var lines []string
	if m.Provider != "" {
		lines = append(lines, fmt.Sprintf("  Provider: %s", m.Provider))
	}
	if m.Key != nil && m.Key.QualifiedName != "" {
		lines = append(lines, fmt.Sprintf("  Key: %s", m.Key.QualifiedName))
	}
	// Portal-populated fields — round-tripped but flagged read-only in MDL.
	if m.DisplayName != "" {
		lines = append(lines, fmt.Sprintf("  DisplayName: '%s'", escapeSQLString(m.DisplayName)))
	}
	if m.KeyName != "" {
		lines = append(lines, fmt.Sprintf("  KeyName: '%s'", escapeSQLString(m.KeyName)))
	}
	if m.KeyID != "" {
		lines = append(lines, fmt.Sprintf("  KeyId: '%s'", escapeSQLString(m.KeyID)))
	}
	if m.Environment != "" {
		lines = append(lines, fmt.Sprintf("  Environment: '%s'", escapeSQLString(m.Environment)))
	}
	if m.ResourceName != "" {
		lines = append(lines, fmt.Sprintf("  ResourceName: '%s'", escapeSQLString(m.ResourceName)))
	}
	if m.DeepLinkURL != "" {
		lines = append(lines, fmt.Sprintf("  DeepLinkURL: '%s'", escapeSQLString(m.DeepLinkURL)))
	}

	for i, line := range lines {
		if i < len(lines)-1 {
			fmt.Fprintln(ctx.Output, line+",")
		} else {
			fmt.Fprintln(ctx.Output, line)
		}
	}

	fmt.Fprintln(ctx.Output, ");")
	fmt.Fprintln(ctx.Output, "/")
	return nil
}

// findAgentEditorModel looks up a model by module and name.
func findAgentEditorModel(ctx *ExecContext, moduleName, modelName string) *agenteditor.Model {
	e := ctx.executor
	models, err := e.reader.ListAgentEditorModels()
	if err != nil {
		return nil
	}
	h, err := getHierarchy(ctx)
	if err != nil {
		return nil
	}
	for _, m := range models {
		modID := h.FindModuleID(m.ContainerID)
		modName := h.GetModuleName(modID)
		if m.Name == modelName && modName == moduleName {
			return m
		}
	}
	return nil
}

// --- Executor method wrappers for backward compatibility ---
