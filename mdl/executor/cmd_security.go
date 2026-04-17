// SPDX-License-Identifier: Apache-2.0

// Package executor - Security commands (SHOW/DESCRIBE/GRANT/REVOKE/CREATE/ALTER/DROP)
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/security"
)

// showProjectSecurity handles SHOW PROJECT SECURITY.
func showProjectSecurity(ctx *ExecContext) error {
	e := ctx.executor
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read project security", err)
	}

	fmt.Fprintf(ctx.Output, "Security Level: %s\n", security.SecurityLevelDisplay(ps.SecurityLevel))
	fmt.Fprintf(ctx.Output, "Check Security: %v\n", ps.CheckSecurity)
	fmt.Fprintf(ctx.Output, "Strict Mode: %v\n", ps.StrictMode)
	fmt.Fprintf(ctx.Output, "Demo Users Enabled: %v\n", ps.EnableDemoUsers)
	fmt.Fprintf(ctx.Output, "Guest Access: %v\n", ps.EnableGuestAccess)
	if ps.AdminUserName != "" {
		fmt.Fprintf(ctx.Output, "Admin User: %s\n", ps.AdminUserName)
	}
	if ps.GuestUserRole != "" {
		fmt.Fprintf(ctx.Output, "Guest User Role: %s\n", ps.GuestUserRole)
	}
	fmt.Fprintf(ctx.Output, "User Roles: %d\n", len(ps.UserRoles))
	fmt.Fprintf(ctx.Output, "Demo Users: %d\n", len(ps.DemoUsers))

	if ps.PasswordPolicy != nil {
		pp := ps.PasswordPolicy
		fmt.Fprintf(ctx.Output, "\nPassword Policy:\n")
		fmt.Fprintf(ctx.Output, "  Minimum Length: %d\n", pp.MinimumLength)
		fmt.Fprintf(ctx.Output, "  Require Digit: %v\n", pp.RequireDigit)
		fmt.Fprintf(ctx.Output, "  Require Mixed Case: %v\n", pp.RequireMixedCase)
		fmt.Fprintf(ctx.Output, "  Require Symbol: %v\n", pp.RequireSymbol)
	}

	return nil
}

// showModuleRoles handles SHOW MODULE ROLES [IN module].
func showModuleRoles(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read module security", err)
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Role", "Description"},
	}

	for _, ms := range allMS {
		modName := h.GetModuleName(ms.ContainerID)
		if modName == "" {
			continue
		}
		if moduleName != "" && modName != moduleName {
			continue
		}
		for _, mr := range ms.ModuleRoles {
			qn := modName + "." + mr.Name
			result.Rows = append(result.Rows, []any{qn, modName, mr.Name, mr.Description})
		}
	}

	result.Summary = fmt.Sprintf("(%d module roles)", len(result.Rows))
	return writeResult(ctx, result)
}

// showUserRoles handles SHOW USER ROLES.
func showUserRoles(ctx *ExecContext) error {
	e := ctx.executor
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read project security", err)
	}

	result := &TableResult{
		Columns: []string{"Name", "Module Roles", "Manage All", "Check Security"},
	}

	for _, ur := range ps.UserRoles {
		ma := "No"
		if ur.ManageAllRoles {
			ma = "Yes"
		}
		cs := "No"
		if ur.CheckSecurity {
			cs = "Yes"
		}
		result.Rows = append(result.Rows, []any{ur.Name, len(ur.ModuleRoles), ma, cs})
	}

	result.Summary = fmt.Sprintf("(%d user roles)", len(result.Rows))
	return writeResult(ctx, result)
}

// showDemoUsers handles SHOW DEMO USERS.
func showDemoUsers(ctx *ExecContext) error {
	e := ctx.executor
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read project security", err)
	}

	if !ps.EnableDemoUsers {
		fmt.Fprintln(ctx.Output, "Demo users are disabled.")
		fmt.Fprintln(ctx.Output, "Enable with: ALTER PROJECT SECURITY DEMO USERS ON;")
		return nil
	}

	result := &TableResult{
		Columns: []string{"User Name", "User Roles"},
	}

	for _, du := range ps.DemoUsers {
		rolesStr := strings.Join(du.UserRoles, ", ")
		result.Rows = append(result.Rows, []any{du.UserName, rolesStr})
	}

	result.Summary = fmt.Sprintf("(%d demo users)", len(result.Rows))
	return writeResult(ctx, result)
}

// showAccessOnEntity handles SHOW ACCESS ON Module.Entity.
func showAccessOnEntity(ctx *ExecContext, name *ast.QualifiedName) error {
	e := ctx.executor
	if name == nil {
		return mdlerrors.NewValidation("entity name required")
	}

	module, err := findModule(ctx, name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return mdlerrors.NewBackend("get domain model", err)
	}

	var entity *domainmodel.Entity
	for _, ent := range dm.Entities {
		if ent.Name == name.Name {
			entity = ent
			break
		}
	}
	if entity == nil {
		return mdlerrors.NewNotFound("entity", name.String())
	}

	if len(entity.AccessRules) == 0 {
		fmt.Fprintf(ctx.Output, "No access rules on %s\n", name)
		return nil
	}

	// Build attribute name map
	attrNames := make(map[string]string)
	for _, attr := range entity.Attributes {
		attrNames[string(attr.ID)] = attr.Name
	}

	fmt.Fprintf(ctx.Output, "Access rules for %s.%s:\n\n", name.Module, name.Name)

	for i, rule := range entity.AccessRules {
		// Show roles
		var roleStrs []string
		for _, rn := range rule.ModuleRoleNames {
			roleStrs = append(roleStrs, rn)
		}
		if len(roleStrs) == 0 {
			for _, rid := range rule.ModuleRoles {
				roleStrs = append(roleStrs, string(rid))
			}
		}
		fmt.Fprintf(ctx.Output, "Rule %d: %s\n", i+1, strings.Join(roleStrs, ", "))

		// Show CRUD rights (READ/WRITE inferred from DefaultMemberAccessRights + MemberAccesses)
		var rights []string
		if rule.AllowCreate {
			rights = append(rights, "CREATE")
		}
		hasRead := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadOnly ||
			rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
		hasWrite := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
		for _, ma := range rule.MemberAccesses {
			if ma.AccessRights == domainmodel.MemberAccessRightsReadOnly || ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
				hasRead = true
			}
			if ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
				hasWrite = true
			}
		}
		if hasRead {
			rights = append(rights, "READ")
		}
		if hasWrite {
			rights = append(rights, "WRITE")
		}
		if rule.AllowDelete {
			rights = append(rights, "DELETE")
		}
		fmt.Fprintf(ctx.Output, "  Rights: %s\n", strings.Join(rights, ", "))

		// Show default member access
		if rule.DefaultMemberAccessRights != "" {
			fmt.Fprintf(ctx.Output, "  Default member access: %s\n", rule.DefaultMemberAccessRights)
		}

		// Show member-level access
		for _, ma := range rule.MemberAccesses {
			memberName := ma.AttributeName
			if memberName == "" {
				memberName = ma.AssociationName
			}
			if memberName == "" {
				if an, ok := attrNames[string(ma.AttributeID)]; ok {
					memberName = an
				} else {
					memberName = string(ma.AttributeID)
				}
			}
			fmt.Fprintf(ctx.Output, "  %s: %s\n", memberName, ma.AccessRights)
		}

		// Show XPath constraint
		if rule.XPathConstraint != "" {
			fmt.Fprintf(ctx.Output, "  WHERE '%s'\n", rule.XPathConstraint)
		}
		fmt.Fprintln(ctx.Output)
	}

	return nil
}

// showAccessOnMicroflow handles SHOW ACCESS ON MICROFLOW Module.MF.
func showAccessOnMicroflow(ctx *ExecContext, name *ast.QualifiedName) error {
	e := ctx.executor
	if name == nil {
		return mdlerrors.NewValidation("microflow name required")
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	mfs, err := e.reader.ListMicroflows()
	if err != nil {
		return mdlerrors.NewBackend("list microflows", err)
	}

	for _, mf := range mfs {
		modName := h.GetModuleName(h.FindModuleID(mf.ContainerID))
		if modName == name.Module && mf.Name == name.Name {
			if len(mf.AllowedModuleRoles) == 0 {
				fmt.Fprintf(ctx.Output, "No module roles granted execute access on %s.%s\n", modName, mf.Name)
				return nil
			}
			fmt.Fprintf(ctx.Output, "Allowed module roles for %s.%s:\n", modName, mf.Name)
			for _, role := range mf.AllowedModuleRoles {
				fmt.Fprintf(ctx.Output, "  %s\n", string(role))
			}
			return nil
		}
	}

	return mdlerrors.NewNotFound("microflow", name.String())
}

// showAccessOnPage handles SHOW ACCESS ON PAGE Module.Page.
func showAccessOnPage(ctx *ExecContext, name *ast.QualifiedName) error {
	e := ctx.executor
	if name == nil {
		return mdlerrors.NewValidation("page name required")
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	pages, err := e.reader.ListPages()
	if err != nil {
		return mdlerrors.NewBackend("list pages", err)
	}

	for _, pg := range pages {
		modName := h.GetModuleName(h.FindModuleID(pg.ContainerID))
		if modName == name.Module && pg.Name == name.Name {
			if len(pg.AllowedRoles) == 0 {
				fmt.Fprintf(ctx.Output, "No module roles granted view access on %s.%s\n", modName, pg.Name)
				return nil
			}
			fmt.Fprintf(ctx.Output, "Allowed module roles for %s.%s:\n", modName, pg.Name)
			for _, role := range pg.AllowedRoles {
				fmt.Fprintf(ctx.Output, "  %s\n", string(role))
			}
			return nil
		}
	}

	return mdlerrors.NewNotFound("page", name.String())
}

// showAccessOnWorkflow handles SHOW ACCESS ON WORKFLOW Module.WF.
func showAccessOnWorkflow(ctx *ExecContext, name *ast.QualifiedName) error {
	return mdlerrors.NewUnsupported("SHOW ACCESS ON WORKFLOW is not supported: Mendix workflows do not have document-level AllowedModuleRoles (unlike microflows and pages). Workflow access is controlled through the microflow that triggers the workflow and UserTask targeting")
}

// showSecurityMatrix handles SHOW SECURITY MATRIX [IN module].
func showSecurityMatrix(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	if e.format == FormatJSON {
		return showSecurityMatrixJSON(ctx, moduleName)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Collect all module roles
	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read module security", err)
	}

	// Build role list for the target module(s)
	type moduleRoleInfo struct {
		moduleName string
		roleName   string
	}
	var roles []moduleRoleInfo
	for _, ms := range allMS {
		modName := h.GetModuleName(ms.ContainerID)
		if modName == "" {
			continue
		}
		if moduleName != "" && modName != moduleName {
			continue
		}
		for _, mr := range ms.ModuleRoles {
			roles = append(roles, moduleRoleInfo{modName, mr.Name})
		}
	}

	if len(roles) == 0 {
		if moduleName != "" {
			fmt.Fprintf(ctx.Output, "No module roles found in %s\n", moduleName)
		} else {
			fmt.Fprintln(ctx.Output, "No module roles found")
		}
		return nil
	}

	// Build role column headers
	var roleHeaders []string
	for _, r := range roles {
		roleHeaders = append(roleHeaders, r.moduleName+"."+r.roleName)
	}

	// Collect entities with access rules
	dms, err := e.reader.ListDomainModels()
	if err != nil {
		return mdlerrors.NewBackend("list domain models", err)
	}

	fmt.Fprintf(ctx.Output, "Security Matrix")
	if moduleName != "" {
		fmt.Fprintf(ctx.Output, " for %s", moduleName)
	}
	fmt.Fprintln(ctx.Output, ":")
	fmt.Fprintln(ctx.Output)

	// Entities section
	fmt.Fprintln(ctx.Output, "## Entity Access")
	fmt.Fprintln(ctx.Output)

	entityFound := false
	for _, dm := range dms {
		modID := h.FindModuleID(dm.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}

		for _, entity := range dm.Entities {
			if len(entity.AccessRules) == 0 {
				continue
			}
			entityFound = true
			fmt.Fprintf(ctx.Output, "### %s.%s\n", modName, entity.Name)

			for _, rule := range entity.AccessRules {
				var roleStrs []string
				for _, rn := range rule.ModuleRoleNames {
					roleStrs = append(roleStrs, rn)
				}
				if len(roleStrs) == 0 {
					for _, rid := range rule.ModuleRoles {
						roleStrs = append(roleStrs, string(rid))
					}
				}

				var rights []string
				if rule.AllowCreate {
					rights = append(rights, "C")
				}
				rr := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadOnly ||
					rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
				rw := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
				for _, ma := range rule.MemberAccesses {
					if ma.AccessRights == domainmodel.MemberAccessRightsReadOnly || ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
						rr = true
					}
					if ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
						rw = true
					}
				}
				if rr {
					rights = append(rights, "R")
				}
				if rw {
					rights = append(rights, "W")
				}
				if rule.AllowDelete {
					rights = append(rights, "D")
				}

				fmt.Fprintf(ctx.Output, "  %s: %s\n", strings.Join(roleStrs, ", "), strings.Join(rights, ""))
			}
			fmt.Fprintln(ctx.Output)
		}
	}
	if !entityFound {
		fmt.Fprintln(ctx.Output, "(no entity access rules configured)")
		fmt.Fprintln(ctx.Output)
	}

	// Microflow section
	fmt.Fprintln(ctx.Output, "## Microflow Access")
	fmt.Fprintln(ctx.Output)

	mfs, err := e.reader.ListMicroflows()
	if err != nil {
		return mdlerrors.NewBackend("list microflows", err)
	}

	mfFound := false
	for _, mf := range mfs {
		if len(mf.AllowedModuleRoles) == 0 {
			continue
		}
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		mfFound = true
		var roleStrs []string
		for _, r := range mf.AllowedModuleRoles {
			roleStrs = append(roleStrs, string(r))
		}
		fmt.Fprintf(ctx.Output, "  %s.%s: %s\n", modName, mf.Name, strings.Join(roleStrs, ", "))
	}
	if !mfFound {
		fmt.Fprintln(ctx.Output, "(no microflow access rules configured)")
	}
	fmt.Fprintln(ctx.Output)

	// Page section
	fmt.Fprintln(ctx.Output, "## Page Access")
	fmt.Fprintln(ctx.Output)

	pages, err := e.reader.ListPages()
	if err != nil {
		return mdlerrors.NewBackend("list pages", err)
	}

	pgFound := false
	for _, pg := range pages {
		if len(pg.AllowedRoles) == 0 {
			continue
		}
		modID := h.FindModuleID(pg.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		pgFound = true
		var roleStrs []string
		for _, r := range pg.AllowedRoles {
			roleStrs = append(roleStrs, string(r))
		}
		fmt.Fprintf(ctx.Output, "  %s.%s: %s\n", modName, pg.Name, strings.Join(roleStrs, ", "))
	}
	if !pgFound {
		fmt.Fprintln(ctx.Output, "(no page access rules configured)")
	}
	fmt.Fprintln(ctx.Output)

	// Workflow section — workflows don't have document-level AllowedModuleRoles
	fmt.Fprintln(ctx.Output, "## Workflow Access")
	fmt.Fprintln(ctx.Output)
	fmt.Fprintln(ctx.Output, "(workflow access is controlled through triggering microflows and UserTask targeting, not document-level roles)")
	fmt.Fprintln(ctx.Output)

	return nil
}

// showSecurityMatrixJSON emits the security matrix as a JSON table
// with one row per access rule across entities, microflows, pages, and workflows.
func showSecurityMatrixJSON(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	tr := &TableResult{
		Columns: []string{"ObjectType", "QualifiedName", "Roles", "Rights"},
	}

	// Entities
	dms, _ := e.reader.ListDomainModels()
	for _, dm := range dms {
		modID := h.FindModuleID(dm.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		for _, entity := range dm.Entities {
			for _, rule := range entity.AccessRules {
				var roleStrs []string
				for _, rn := range rule.ModuleRoleNames {
					roleStrs = append(roleStrs, rn)
				}
				if len(roleStrs) == 0 {
					for _, rid := range rule.ModuleRoles {
						roleStrs = append(roleStrs, string(rid))
					}
				}

				var rights []string
				if rule.AllowCreate {
					rights = append(rights, "C")
				}
				rr := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadOnly ||
					rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
				rw := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
				for _, ma := range rule.MemberAccesses {
					if ma.AccessRights == domainmodel.MemberAccessRightsReadOnly || ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
						rr = true
					}
					if ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
						rw = true
					}
				}
				if rr {
					rights = append(rights, "R")
				}
				if rw {
					rights = append(rights, "W")
				}
				if rule.AllowDelete {
					rights = append(rights, "D")
				}

				tr.Rows = append(tr.Rows, []any{
					"Entity",
					modName + "." + entity.Name,
					strings.Join(roleStrs, ", "),
					strings.Join(rights, ""),
				})
			}
		}
	}

	// Microflows
	mfs, _ := e.reader.ListMicroflows()
	for _, mf := range mfs {
		if len(mf.AllowedModuleRoles) == 0 {
			continue
		}
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		var roleStrs []string
		for _, r := range mf.AllowedModuleRoles {
			roleStrs = append(roleStrs, string(r))
		}
		tr.Rows = append(tr.Rows, []any{
			"Microflow",
			modName + "." + mf.Name,
			strings.Join(roleStrs, ", "),
			"X",
		})
	}

	// Pages
	pages, _ := e.reader.ListPages()
	for _, pg := range pages {
		if len(pg.AllowedRoles) == 0 {
			continue
		}
		modID := h.FindModuleID(pg.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		var roleStrs []string
		for _, r := range pg.AllowedRoles {
			roleStrs = append(roleStrs, string(r))
		}
		tr.Rows = append(tr.Rows, []any{
			"Page",
			modName + "." + pg.Name,
			strings.Join(roleStrs, ", "),
			"X",
		})
	}

	return writeResult(ctx, tr)
}

// describeModuleRole handles DESCRIBE MODULE ROLE Module.RoleName.
func describeModuleRole(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read module security", err)
	}

	for _, ms := range allMS {
		modName := h.GetModuleName(ms.ContainerID)
		if name.Module != "" && modName != name.Module {
			continue
		}
		for _, mr := range ms.ModuleRoles {
			if mr.Name == name.Name {
				fmt.Fprintf(ctx.Output, "CREATE MODULE ROLE %s.%s", modName, mr.Name)
				if mr.Description != "" {
					fmt.Fprintf(ctx.Output, " DESCRIPTION '%s'", mr.Description)
				}
				fmt.Fprintln(ctx.Output, ";")
				fmt.Fprintln(ctx.Output, "/")

				// Show which user roles include this module role
				qualifiedRole := modName + "." + mr.Name
				ps, err := e.reader.GetProjectSecurity()
				if err == nil {
					var includedBy []string
					for _, ur := range ps.UserRoles {
						for _, mrRef := range ur.ModuleRoles {
							if mrRef == qualifiedRole {
								includedBy = append(includedBy, ur.Name)
							}
						}
					}
					if len(includedBy) > 0 {
						fmt.Fprintf(ctx.Output, "\n-- Included in user roles: %s\n", strings.Join(includedBy, ", "))
					}
				}

				return nil
			}
		}
	}

	return mdlerrors.NewNotFound("module role", name.String())
}

// describeDemoUser handles DESCRIBE DEMO USER 'name'.
func describeDemoUser(ctx *ExecContext, userName string) error {
	e := ctx.executor
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read project security", err)
	}

	for _, du := range ps.DemoUsers {
		if du.UserName == userName {
			fmt.Fprintf(ctx.Output, "CREATE DEMO USER '%s' PASSWORD '***'", du.UserName)
			if du.Entity != "" {
				fmt.Fprintf(ctx.Output, " ENTITY %s", du.Entity)
			}
			if len(du.UserRoles) > 0 {
				fmt.Fprintf(ctx.Output, " (%s)", strings.Join(du.UserRoles, ", "))
			}
			fmt.Fprintln(ctx.Output, ";")
			fmt.Fprintln(ctx.Output, "/")
			return nil
		}
	}

	return mdlerrors.NewNotFound("demo user", userName)
}

// describeUserRole handles DESCRIBE USER ROLE Name.
func describeUserRole(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return mdlerrors.NewBackend("read project security", err)
	}

	for _, ur := range ps.UserRoles {
		if ur.Name == name.Name {
			fmt.Fprintf(ctx.Output, "CREATE USER ROLE %s", ur.Name)

			// Module roles
			if len(ur.ModuleRoles) > 0 {
				fmt.Fprintf(ctx.Output, " (%s)", strings.Join(ur.ModuleRoles, ", "))
			}

			if ur.ManageAllRoles {
				fmt.Fprint(ctx.Output, " MANAGE ALL ROLES")
			}

			fmt.Fprintln(ctx.Output, ";")
			fmt.Fprintln(ctx.Output, "/")

			// Show description if present
			if ur.Description != "" {
				fmt.Fprintf(ctx.Output, "\n-- Description: %s\n", ur.Description)
			}

			// Show check security flag
			if ur.CheckSecurity {
				fmt.Fprintln(ctx.Output, "-- Check security: enabled")
			}

			return nil
		}
	}

	return mdlerrors.NewNotFound("user role", name.Name)
}

// Executor method wrappers — delegate to free functions for callers that
// still use the Executor receiver (e.g. executor_query.go).
