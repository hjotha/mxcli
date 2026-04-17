// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"

	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
)

// showLanguages lists all languages found in the project's translatable strings.
// Requires REFRESH CATALOG FULL to populate the strings table.
func showLanguages(ctx *ExecContext) error {
	if ctx.Catalog == nil {
		return mdlerrors.NewValidation("no catalog available — run REFRESH CATALOG FULL first")
	}

	result, err := ctx.Catalog.Query(`
		SELECT Language, COUNT(*) as StringCount
		FROM strings
		WHERE Language != ''
		GROUP BY Language
		ORDER BY StringCount DESC
	`)
	if err != nil {
		return mdlerrors.NewBackend("query languages", err)
	}

	if len(result.Rows) == 0 {
		fmt.Fprintln(ctx.Output, "No translatable strings found. Run REFRESH CATALOG FULL to populate the strings table.")
		return nil
	}

	tr := &TableResult{
		Columns: []string{"Language", "Strings"},
		Summary: fmt.Sprintf("(%d languages)", len(result.Rows)),
	}
	for _, row := range result.Rows {
		lang := ""
		count := ""
		if len(row) > 0 {
			lang = fmt.Sprintf("%v", row[0])
		}
		if len(row) > 1 {
			count = fmt.Sprintf("%v", row[1])
		}
		tr.Rows = append(tr.Rows, []any{lang, count})
	}
	return writeResult(ctx, tr)
}

// --- Executor method wrapper for backward compatibility ---
