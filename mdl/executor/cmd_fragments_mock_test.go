// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestShowFragments_Mock(t *testing.T) {
	ctx, buf := newMockCtx(t)
	ctx.Fragments = map[string]*ast.DefineFragmentStmt{
		"myFrag": {Name: "myFrag"},
	}

	assertNoError(t, showFragments(ctx))

	out := buf.String()
	assertContainsStr(t, out, "myFrag")
}

func TestShowFragments_Empty_Mock(t *testing.T) {
	ctx, buf := newMockCtx(t)
	ctx.Fragments = map[string]*ast.DefineFragmentStmt{}

	assertNoError(t, showFragments(ctx))

	out := buf.String()
	assertContainsStr(t, out, "No fragments defined.")
}
