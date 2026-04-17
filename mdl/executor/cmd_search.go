// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
)

// execShowCallers handles SHOW CALLERS OF Module.Microflow [TRANSITIVE].
func execShowCallers(ctx *ExecContext, s *ast.ShowStmt) error {
	if s.Name == nil {
		return mdlerrors.NewValidation("target name required for SHOW CALLERS")
	}

	// Ensure catalog is available with full mode for refs
	if err := ensureCatalog(ctx, true); err != nil {
		return err
	}

	targetName := s.Name.String()
	fmt.Fprintf(ctx.Output, "\nCallers of %s", targetName)
	if s.Transitive {
		fmt.Fprintln(ctx.Output, " (transitive)")
	} else {
		fmt.Fprintln(ctx.Output, "")
	}

	var query string
	if s.Transitive {
		// Recursive CTE for transitive callers
		query = `
			WITH RECURSIVE callers_cte AS (
				SELECT SourceName as Caller, 1 as Depth
				FROM refs
				WHERE TargetName = ? AND RefKind = 'call'
				UNION ALL
				SELECT r.SourceName, c.Depth + 1
				FROM refs r
				JOIN callers_cte c ON r.TargetName = c.Caller
				WHERE r.RefKind = 'call' AND c.Depth < 10
			)
			SELECT DISTINCT Caller, MIN(Depth) as Depth
			FROM callers_cte
			GROUP BY Caller
			ORDER BY Depth, Caller
		`
	} else {
		// Direct callers only
		query = `
			SELECT DISTINCT SourceName as Caller, 1 as Depth
			FROM refs
			WHERE TargetName = ? AND RefKind = 'call'
			ORDER BY Caller
		`
	}

	result, err := ctx.Catalog.Query(strings.Replace(query, "?", "'"+targetName+"'", 1))
	if err != nil {
		return mdlerrors.NewBackend("query callers", err)
	}

	if result.Count == 0 {
		fmt.Fprintln(ctx.Output, "(no callers found)")
		return nil
	}

	fmt.Fprintf(ctx.Output, "Found %d caller(s)\n", result.Count)
	outputCatalogResults(ctx, result)
	return nil
}

// execShowCallees handles SHOW CALLEES OF Module.Microflow [TRANSITIVE].
func execShowCallees(ctx *ExecContext, s *ast.ShowStmt) error {
	if s.Name == nil {
		return mdlerrors.NewValidation("target name required for SHOW CALLEES")
	}

	// Ensure catalog is available with full mode for refs
	if err := ensureCatalog(ctx, true); err != nil {
		return err
	}

	sourceName := s.Name.String()
	fmt.Fprintf(ctx.Output, "\nCallees of %s", sourceName)
	if s.Transitive {
		fmt.Fprintln(ctx.Output, " (transitive)")
	} else {
		fmt.Fprintln(ctx.Output, "")
	}

	var query string
	if s.Transitive {
		// Recursive CTE for transitive callees
		query = `
			WITH RECURSIVE callees_cte AS (
				SELECT TargetName as Callee, 1 as Depth
				FROM refs
				WHERE SourceName = ? AND RefKind = 'call'
				UNION ALL
				SELECT r.TargetName, c.Depth + 1
				FROM refs r
				JOIN callees_cte c ON r.SourceName = c.Callee
				WHERE r.RefKind = 'call' AND c.Depth < 10
			)
			SELECT DISTINCT Callee, MIN(Depth) as Depth
			FROM callees_cte
			GROUP BY Callee
			ORDER BY Depth, Callee
		`
	} else {
		// Direct callees only
		query = `
			SELECT DISTINCT TargetName as Callee, 1 as Depth
			FROM refs
			WHERE SourceName = ? AND RefKind = 'call'
			ORDER BY Callee
		`
	}

	result, err := ctx.Catalog.Query(strings.Replace(query, "?", "'"+sourceName+"'", 1))
	if err != nil {
		return mdlerrors.NewBackend("query callees", err)
	}

	if result.Count == 0 {
		fmt.Fprintln(ctx.Output, "(no callees found)")
		return nil
	}

	fmt.Fprintf(ctx.Output, "Found %d callee(s)\n", result.Count)
	outputCatalogResults(ctx, result)
	return nil
}

// execShowReferences handles SHOW REFERENCES TO Module.Entity.
func execShowReferences(ctx *ExecContext, s *ast.ShowStmt) error {
	if s.Name == nil {
		return mdlerrors.NewValidation("target name required for SHOW REFERENCES")
	}

	// Ensure catalog is available with full mode for refs
	if err := ensureCatalog(ctx, true); err != nil {
		return err
	}

	targetName := s.Name.String()
	fmt.Fprintf(ctx.Output, "\nReferences to %s\n", targetName)

	// Find all references to this target
	query := `
		SELECT SourceType, SourceName, RefKind
		FROM refs
		WHERE TargetName = ?
		ORDER BY RefKind, SourceType, SourceName
	`

	result, err := ctx.Catalog.Query(strings.Replace(query, "?", "'"+targetName+"'", 1))
	if err != nil {
		return mdlerrors.NewBackend("query references", err)
	}

	if result.Count == 0 {
		fmt.Fprintln(ctx.Output, "(no references found)")
		return nil
	}

	fmt.Fprintf(ctx.Output, "Found %d reference(s)\n", result.Count)
	outputCatalogResults(ctx, result)
	return nil
}

// execShowImpact handles SHOW IMPACT OF Module.Entity.
// This shows all elements that would be affected by changing the target.
func execShowImpact(ctx *ExecContext, s *ast.ShowStmt) error {
	if s.Name == nil {
		return mdlerrors.NewValidation("target name required for SHOW IMPACT")
	}

	// Ensure catalog is available with full mode for refs
	if err := ensureCatalog(ctx, true); err != nil {
		return err
	}

	targetName := s.Name.String()
	fmt.Fprintf(ctx.Output, "\nImpact analysis for %s\n", targetName)

	// Find all direct references to this target
	directQuery := `
		SELECT SourceType, SourceName, RefKind
		FROM refs
		WHERE TargetName = ?
		ORDER BY SourceType, SourceName
	`

	result, err := ctx.Catalog.Query(strings.Replace(directQuery, "?", "'"+targetName+"'", 1))
	if err != nil {
		return mdlerrors.NewBackend("query impact", err)
	}

	if result.Count == 0 {
		fmt.Fprintln(ctx.Output, "(no impact - element is not referenced)")
		return nil
	}

	// Group by type for summary
	typeCounts := make(map[string]int)
	for _, row := range result.Rows {
		if len(row) > 0 {
			if t, ok := row[0].(string); ok {
				typeCounts[t]++
			}
		}
	}

	fmt.Fprintf(ctx.Output, "\nSummary:\n")
	for t, count := range typeCounts {
		fmt.Fprintf(ctx.Output, "  %s: %d\n", t, count)
	}
	fmt.Fprintln(ctx.Output)

	fmt.Fprintf(ctx.Output, "Found %d affected element(s)\n", result.Count)
	outputCatalogResults(ctx, result)

	return nil
}

// --- Executor method wrappers for backward compatibility ---
