// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"
)

func TestParseOQLString_SimpleQuery(t *testing.T) {
	oql := "from CRM.Account as a select a.Name, a.Email"
	parsed := ParseOQLString(oql)
	if parsed == nil {
		t.Fatal("Expected non-nil parsed result")
	}

	if len(parsed.Tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(parsed.Tables))
	}
	if parsed.Tables[0].Entity != "CRM.Account" {
		t.Errorf("Expected entity CRM.Account, got %q", parsed.Tables[0].Entity)
	}
	if parsed.Tables[0].Alias != "a" {
		t.Errorf("Expected alias 'a', got %q", parsed.Tables[0].Alias)
	}
	if len(parsed.Select) != 2 {
		t.Errorf("Expected 2 select items, got %d", len(parsed.Select))
	}
}

func TestParseOQLString_SubqueryInFROM(t *testing.T) {
	oql := "from (from MultipleAggregate.Policy as p2 select count(p2.PolicyNumber) as TotalPolicies) as p select p.TotalPolicies"
	parsed := ParseOQLString(oql)
	if parsed == nil {
		t.Fatal("Expected non-nil parsed result")
	}

	if len(parsed.Tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(parsed.Tables))
	}

	table := parsed.Tables[0]
	if table.Alias != "p" {
		t.Errorf("Expected alias 'p', got %q", table.Alias)
	}
	if table.JoinType != "from" {
		t.Errorf("Expected joinType 'from', got %q", table.JoinType)
	}
	if table.Entity != "MultipleAggregate.Policy" {
		t.Errorf("Expected entity from inner query 'MultipleAggregate.Policy', got %q", table.Entity)
	}
	if table.Subquery == nil {
		t.Fatal("Expected Subquery to be set")
	}
	if table.SubqueryText == "" {
		t.Error("Expected SubqueryText to be non-empty")
	}

	// Verify inner query structure
	sub := table.Subquery
	if len(sub.Tables) != 1 {
		t.Fatalf("Expected 1 inner table, got %d", len(sub.Tables))
	}
	if sub.Tables[0].Entity != "MultipleAggregate.Policy" {
		t.Errorf("Expected inner entity 'MultipleAggregate.Policy', got %q", sub.Tables[0].Entity)
	}
	if sub.Tables[0].Alias != "p2" {
		t.Errorf("Expected inner alias 'p2', got %q", sub.Tables[0].Alias)
	}
	if len(sub.Select) != 1 {
		t.Fatalf("Expected 1 inner select item, got %d", len(sub.Select))
	}
	if sub.Select[0].Alias != "TotalPolicies" {
		t.Errorf("Expected inner select alias 'TotalPolicies', got %q", sub.Select[0].Alias)
	}
}

func TestParseOQLString_SubqueryInJOIN(t *testing.T) {
	oql := `from MultipleAggregate.Policy as p left join (from MultipleAggregate.Claim as c2 select c2.PolicyNumber as PN, count(c2.ClaimID) as ClaimCount) as c on p.PolicyNumber = c.PN select p.PolicyNumber, c.ClaimCount`
	parsed := ParseOQLString(oql)
	if parsed == nil {
		t.Fatal("Expected non-nil parsed result")
	}

	if len(parsed.Tables) != 2 {
		t.Fatalf("Expected 2 tables, got %d", len(parsed.Tables))
	}

	// First table: direct entity reference
	from := parsed.Tables[0]
	if from.Entity != "MultipleAggregate.Policy" {
		t.Errorf("Expected from entity 'MultipleAggregate.Policy', got %q", from.Entity)
	}
	if from.Alias != "p" {
		t.Errorf("Expected from alias 'p', got %q", from.Alias)
	}
	if from.Subquery != nil {
		t.Error("Expected from table to have no subquery")
	}

	// Second table: subquery in LEFT JOIN
	join := parsed.Tables[1]
	if join.JoinType != "left join" {
		t.Errorf("Expected joinType 'left join', got %q", join.JoinType)
	}
	if join.Alias != "c" {
		t.Errorf("Expected alias 'c', got %q", join.Alias)
	}
	if join.Entity != "MultipleAggregate.Claim" {
		t.Errorf("Expected entity from inner query 'MultipleAggregate.Claim', got %q", join.Entity)
	}
	if join.OnExpr != "p.PolicyNumber = c.PN" {
		t.Errorf("Expected on expr 'p.PolicyNumber = c.PN', got %q", join.OnExpr)
	}
	if join.Subquery == nil {
		t.Fatal("Expected Subquery to be set on join table")
	}

	sub := join.Subquery
	if len(sub.Tables) != 1 {
		t.Fatalf("Expected 1 inner table, got %d", len(sub.Tables))
	}
	if len(sub.Select) != 2 {
		t.Fatalf("Expected 2 inner select items, got %d", len(sub.Select))
	}
}

func TestParseOQLString_MultipleSubqueries(t *testing.T) {
	oql := `from (from MyMod.A as a1 select a1.X as AX) as a left join (from MyMod.B as b1 select b1.Y as by) as b on a.AX = b.BY select a.AX, b.BY`
	parsed := ParseOQLString(oql)
	if parsed == nil {
		t.Fatal("Expected non-nil parsed result")
	}

	if len(parsed.Tables) != 2 {
		t.Fatalf("Expected 2 tables, got %d", len(parsed.Tables))
	}

	// Both tables should have subqueries
	if parsed.Tables[0].Subquery == nil {
		t.Error("Expected from table to have subquery")
	}
	if parsed.Tables[1].Subquery == nil {
		t.Error("Expected join table to have subquery")
	}
	if parsed.Tables[0].Entity != "MyMod.A" {
		t.Errorf("Expected from entity 'MyMod.A', got %q", parsed.Tables[0].Entity)
	}
	if parsed.Tables[1].Entity != "MyMod.B" {
		t.Errorf("Expected join entity 'MyMod.B', got %q", parsed.Tables[1].Entity)
	}
}

func TestQueryPlan_SubqueryInFROM(t *testing.T) {
	oql := "from (from Insurance.Policy as p2 select count(p2.Number) as Total) as p select p.Total"
	plan := parseOqlPlan("Insurance.MyView", oql)

	if len(plan.Tables) != 1 {
		t.Fatalf("Expected 1 plan table, got %d", len(plan.Tables))
	}

	table := plan.Tables[0]
	if table.Entity != "(From) Insurance.Policy" {
		t.Errorf("Expected entity '(From) Insurance.Policy', got %q", table.Entity)
	}
	if table.Alias != "p" {
		t.Errorf("Expected alias 'p', got %q", table.Alias)
	}
	// Subquery SELECT items should populate attributes
	if len(table.Attributes) < 1 {
		t.Errorf("Expected at least 1 attribute from subquery select, got %d", len(table.Attributes))
	}
}

func TestQueryPlan_SubqueryInJOIN(t *testing.T) {
	oql := `from Insurance.Policy as p left join (from Insurance.Claim as c2 where c2.Active = true select c2.PolicyNum as PN, count(c2.ID) as Cnt) as c on p.Num = c.PN select p.Num, c.Cnt`
	plan := parseOqlPlan("Insurance.MyView", oql)

	if len(plan.Tables) != 2 {
		t.Fatalf("Expected 2 plan tables, got %d", len(plan.Tables))
	}

	// First table: regular entity
	if plan.Tables[0].Entity != "Insurance.Policy" {
		t.Errorf("Expected entity 'Insurance.Policy', got %q", plan.Tables[0].Entity)
	}

	// Second table: subquery-derived
	joinTable := plan.Tables[1]
	if joinTable.Entity != "(From) Insurance.Claim" {
		t.Errorf("Expected entity '(From) Insurance.Claim', got %q", joinTable.Entity)
	}
	if joinTable.Alias != "c" {
		t.Errorf("Expected alias 'c', got %q", joinTable.Alias)
	}

	// Subquery attributes from inner SELECT
	if len(joinTable.Attributes) < 2 {
		t.Errorf("Expected at least 2 attributes from subquery select, got %d", len(joinTable.Attributes))
	}

	// Subquery filters from inner WHERE
	if len(joinTable.Filters) < 1 {
		t.Errorf("Expected at least 1 filter from subquery where, got %d", len(joinTable.Filters))
	}

	// Join edges should be created
	if len(plan.Joins) != 1 {
		t.Fatalf("Expected 1 join, got %d", len(plan.Joins))
	}
	if plan.Joins[0].JoinType != "left join" {
		t.Errorf("Expected join type 'left join', got %q", plan.Joins[0].JoinType)
	}
	if plan.Joins[0].Condition != "p.Num = c.PN" {
		t.Errorf("Expected join condition 'p.Num = c.PN', got %q", plan.Joins[0].Condition)
	}
}

func TestParseOQLString_ScalarSubqueryInSELECT(t *testing.T) {
	oql := `from Shop.TagValue as tv select tv.Tag, (from Shop.TagValue as tv2 where tv2.Tag = tv.Tag select count(tv2.Value)) as ValueCount`
	parsed := ParseOQLString(oql)
	if parsed == nil {
		t.Fatal("Expected non-nil parsed result")
	}

	if len(parsed.Select) != 2 {
		t.Fatalf("Expected 2 select items, got %d", len(parsed.Select))
	}

	// First item: simple column reference
	if parsed.Select[0].Subquery != nil {
		t.Error("Expected first select item to have no subquery")
	}

	// Second item: scalar subquery
	sel := parsed.Select[1]
	if sel.Alias != "ValueCount" {
		t.Errorf("Expected alias 'ValueCount', got %q", sel.Alias)
	}
	if sel.Subquery == nil {
		t.Fatal("Expected second select item to have a subquery")
	}
	if len(sel.Subquery.Tables) != 1 {
		t.Fatalf("Expected 1 inner table, got %d", len(sel.Subquery.Tables))
	}
	if sel.Subquery.Tables[0].Entity != "Shop.TagValue" {
		t.Errorf("Expected inner entity 'Shop.TagValue', got %q", sel.Subquery.Tables[0].Entity)
	}
	if sel.Subquery.Where == "" {
		t.Error("Expected inner where clause to be non-empty")
	}
}

func TestQueryPlan_ScalarSubqueryInSELECT(t *testing.T) {
	oql := `from Shop.TagValue as tv select tv.Tag, (from Shop.TagValue as tv2 where tv2.Tag = tv.Tag select count(tv2.Value)) as ValueCount`
	plan := parseOqlPlan("Shop.MyView", oql)

	// Should have 2 tables: the FROM table + the scalar subquery table
	if len(plan.Tables) != 2 {
		t.Fatalf("Expected 2 plan tables, got %d", len(plan.Tables))
	}

	// First table: regular FROM entity
	if plan.Tables[0].Entity != "Shop.TagValue" {
		t.Errorf("Expected entity 'Shop.TagValue', got %q", plan.Tables[0].Entity)
	}

	// Second table: scalar subquery
	subTable := plan.Tables[1]
	if subTable.JoinType != "scalar" {
		t.Errorf("Expected joinType 'scalar', got %q", subTable.JoinType)
	}
	if subTable.Entity != "(Select) Shop.TagValue" {
		t.Errorf("Expected entity '(Select) Shop.TagValue', got %q", subTable.Entity)
	}
	if subTable.Alias != "ValueCount" {
		t.Errorf("Expected alias 'ValueCount', got %q", subTable.Alias)
	}
	// Should have the inner SELECT as attributes
	if len(subTable.Attributes) < 1 {
		t.Errorf("Expected at least 1 attribute from inner select, got %d", len(subTable.Attributes))
	}
	// Should have the inner WHERE as a filter
	if len(subTable.Filters) < 1 {
		t.Errorf("Expected at least 1 filter from inner where, got %d", len(subTable.Filters))
	}
}

func TestQueryPlan_NoSubquery(t *testing.T) {
	// Ensure regular queries still work unchanged
	oql := "from CRM.Account as a join CRM.Contact as c on a.ID = c.AccountID select a.Name, c.Email"
	plan := parseOqlPlan("CRM.MyView", oql)

	if len(plan.Tables) != 2 {
		t.Fatalf("Expected 2 plan tables, got %d", len(plan.Tables))
	}
	if plan.Tables[0].Entity != "CRM.Account" {
		t.Errorf("Expected entity 'CRM.Account', got %q", plan.Tables[0].Entity)
	}
	if plan.Tables[1].Entity != "CRM.Contact" {
		t.Errorf("Expected entity 'CRM.Contact', got %q", plan.Tables[1].Entity)
	}
	if len(plan.Joins) != 1 {
		t.Fatalf("Expected 1 join, got %d", len(plan.Joins))
	}
}
