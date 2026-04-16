// SPDX-License-Identifier: Apache-2.0

// Package executor - Security commands (SHOW/DESCRIBE/GRANT/REVOKE/CREATE/ALTER/DROP)
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/security"
)

// showProjectSecurity handles SHOW PROJECT SECURITY.
func (e *Executor) showProjectSecurity() error {
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return fmt.Errorf("failed to read project security: %w", err)
	}

	fmt.Fprintf(e.output, "Security Level: %s\n", security.SecurityLevelDisplay(ps.SecurityLevel))
	fmt.Fprintf(e.output, "Check Security: %v\n", ps.CheckSecurity)
	fmt.Fprintf(e.output, "Strict Mode: %v\n", ps.StrictMode)
	fmt.Fprintf(e.output, "Demo Users Enabled: %v\n", ps.EnableDemoUsers)
	fmt.Fprintf(e.output, "Guest Access: %v\n", ps.EnableGuestAccess)
	if ps.AdminUserName != "" {
		fmt.Fprintf(e.output, "Admin User: %s\n", ps.AdminUserName)
	}
	if ps.GuestUserRole != "" {
		fmt.Fprintf(e.output, "Guest User Role: %s\n", ps.GuestUserRole)
	}
	fmt.Fprintf(e.output, "User Roles: %d\n", len(ps.UserRoles))
	fmt.Fprintf(e.output, "Demo Users: %d\n", len(ps.DemoUsers))

	if ps.PasswordPolicy != nil {
		pp := ps.PasswordPolicy
		fmt.Fprintf(e.output, "\nPassword Policy:\n")
		fmt.Fprintf(e.output, "  Minimum Length: %d\n", pp.MinimumLength)
		fmt.Fprintf(e.output, "  Require Digit: %v\n", pp.RequireDigit)
		fmt.Fprintf(e.output, "  Require Mixed Case: %v\n", pp.RequireMixedCase)
		fmt.Fprintf(e.output, "  Require Symbol: %v\n", pp.RequireSymbol)
	}

	return nil
}

// showModuleRoles handles SHOW MODULE ROLES [IN module].
func (e *Executor) showModuleRoles(moduleName string) error {
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return fmt.Errorf("failed to read module security: %w", err)
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
	return e.writeResult(result)
}

// showUserRoles handles SHOW USER ROLES.
func (e *Executor) showUserRoles() error {
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return fmt.Errorf("failed to read project security: %w", err)
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
	return e.writeResult(result)
}

// showDemoUsers handles SHOW DEMO USERS.
func (e *Executor) showDemoUsers() error {
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return fmt.Errorf("failed to read project security: %w", err)
	}

	if !ps.EnableDemoUsers {
		fmt.Fprintln(e.output, "Demo users are disabled.")
		fmt.Fprintln(e.output, "Enable with: ALTER PROJECT SECURITY DEMO USERS ON;")
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
	return e.writeResult(result)
}

// showAccessOnEntity handles SHOW ACCESS ON Module.Entity.
func (e *Executor) showAccessOnEntity(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("entity name required")
	}

	module, err := e.findModule(name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	var entity *domainmodel.Entity
	for _, ent := range dm.Entities {
		if ent.Name == name.Name {
			entity = ent
			break
		}
	}
	if entity == nil {
		return fmt.Errorf("entity not found: %s", name)
	}

	if len(entity.AccessRules) == 0 {
		fmt.Fprintf(e.output, "No access rules on %s\n", name)
		return nil
	}

	// Build attribute name map
	attrNames := make(map[string]string)
	for _, attr := range entity.Attributes {
		attrNames[string(attr.ID)] = attr.Name
	}

	fmt.Fprintf(e.output, "Access rules for %s.%s:\n\n", name.Module, name.Name)

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
		fmt.Fprintf(e.output, "Rule %d: %s\n", i+1, strings.Join(roleStrs, ", "))

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
		fmt.Fprintf(e.output, "  Rights: %s\n", strings.Join(rights, ", "))

		// Show default member access
		if rule.DefaultMemberAccessRights != "" {
			fmt.Fprintf(e.output, "  Default member access: %s\n", rule.DefaultMemberAccessRights)
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
			fmt.Fprintf(e.output, "  %s: %s\n", memberName, ma.AccessRights)
		}

		// Show XPath constraint
		if rule.XPathConstraint != "" {
			fmt.Fprintf(e.output, "  WHERE '%s'\n", rule.XPathConstraint)
		}
		fmt.Fprintln(e.output)
	}

	return nil
}

// showAccessOnMicroflow handles SHOW ACCESS ON MICROFLOW Module.MF.
func (e *Executor) showAccessOnMicroflow(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("microflow name required")
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	mfs, err := e.reader.ListMicroflows()
	if err != nil {
		return fmt.Errorf("failed to list microflows: %w", err)
	}

	for _, mf := range mfs {
		modName := h.GetModuleName(h.FindModuleID(mf.ContainerID))
		if modName == name.Module && mf.Name == name.Name {
			if len(mf.AllowedModuleRoles) == 0 {
				fmt.Fprintf(e.output, "No module roles granted execute access on %s.%s\n", modName, mf.Name)
				return nil
			}
			fmt.Fprintf(e.output, "Allowed module roles for %s.%s:\n", modName, mf.Name)
			for _, role := range mf.AllowedModuleRoles {
				fmt.Fprintf(e.output, "  %s\n", string(role))
			}
			return nil
		}
	}

	return fmt.Errorf("microflow not found: %s", name)
}

// showAccessOnPage handles SHOW ACCESS ON PAGE Module.Page.
func (e *Executor) showAccessOnPage(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("page name required")
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	pages, err := e.reader.ListPages()
	if err != nil {
		return fmt.Errorf("failed to list pages: %w", err)
	}

	for _, pg := range pages {
		modName := h.GetModuleName(h.FindModuleID(pg.ContainerID))
		if modName == name.Module && pg.Name == name.Name {
			if len(pg.AllowedRoles) == 0 {
				fmt.Fprintf(e.output, "No module roles granted view access on %s.%s\n", modName, pg.Name)
				return nil
			}
			fmt.Fprintf(e.output, "Allowed module roles for %s.%s:\n", modName, pg.Name)
			for _, role := range pg.AllowedRoles {
				fmt.Fprintf(e.output, "  %s\n", string(role))
			}
			return nil
		}
	}

	return fmt.Errorf("page not found: %s", name)
}

// showAccessOnWorkflow handles SHOW ACCESS ON WORKFLOW Module.WF.
func (e *Executor) showAccessOnWorkflow(name *ast.QualifiedName) error {
	return fmt.Errorf("SHOW ACCESS ON WORKFLOW is not supported: Mendix workflows do not have document-level AllowedModuleRoles (unlike microflows and pages). Workflow access is controlled through the microflow that triggers the workflow and UserTask targeting")
}

// showSecurityMatrix handles SHOW SECURITY MATRIX [IN module].
func (e *Executor) showSecurityMatrix(moduleName string) error {
	if e.format == FormatJSON {
		return e.showSecurityMatrixJSON(moduleName)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Collect all module roles
	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return fmt.Errorf("failed to read module security: %w", err)
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
			fmt.Fprintf(e.output, "No module roles found in %s\n", moduleName)
		} else {
			fmt.Fprintln(e.output, "No module roles found")
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
		return fmt.Errorf("failed to list domain models: %w", err)
	}

	fmt.Fprintf(e.output, "Security Matrix")
	if moduleName != "" {
		fmt.Fprintf(e.output, " for %s", moduleName)
	}
	fmt.Fprintln(e.output, ":")
	fmt.Fprintln(e.output)

	// Entities section
	fmt.Fprintln(e.output, "## Entity Access")
	fmt.Fprintln(e.output)

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
			fmt.Fprintf(e.output, "### %s.%s\n", modName, entity.Name)

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

				fmt.Fprintf(e.output, "  %s: %s\n", strings.Join(roleStrs, ", "), strings.Join(rights, ""))
			}
			fmt.Fprintln(e.output)
		}
	}
	if !entityFound {
		fmt.Fprintln(e.output, "(no entity access rules configured)")
		fmt.Fprintln(e.output)
	}

	// Microflow section
	fmt.Fprintln(e.output, "## Microflow Access")
	fmt.Fprintln(e.output)

	mfs, err := e.reader.ListMicroflows()
	if err != nil {
		return fmt.Errorf("failed to list microflows: %w", err)
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
		fmt.Fprintf(e.output, "  %s.%s: %s\n", modName, mf.Name, strings.Join(roleStrs, ", "))
	}
	if !mfFound {
		fmt.Fprintln(e.output, "(no microflow access rules configured)")
	}
	fmt.Fprintln(e.output)

	// Page section
	fmt.Fprintln(e.output, "## Page Access")
	fmt.Fprintln(e.output)

	pages, err := e.reader.ListPages()
	if err != nil {
		return fmt.Errorf("failed to list pages: %w", err)
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
		fmt.Fprintf(e.output, "  %s.%s: %s\n", modName, pg.Name, strings.Join(roleStrs, ", "))
	}
	if !pgFound {
		fmt.Fprintln(e.output, "(no page access rules configured)")
	}
	fmt.Fprintln(e.output)

	// Workflow section — workflows don't have document-level AllowedModuleRoles
	fmt.Fprintln(e.output, "## Workflow Access")
	fmt.Fprintln(e.output)
	fmt.Fprintln(e.output, "(workflow access is controlled through triggering microflows and UserTask targeting, not document-level roles)")
	fmt.Fprintln(e.output)

	return nil
}

// showSecurityMatrixJSON emits the security matrix as a JSON table
// with one row per access rule across entities, microflows, pages, and workflows.
func (e *Executor) showSecurityMatrixJSON(moduleName string) error {
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
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

	return e.writeResult(tr)
}

// describeModuleRole handles DESCRIBE MODULE ROLE Module.RoleName.
func (e *Executor) describeModuleRole(name ast.QualifiedName) error {
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	allMS, err := e.reader.ListModuleSecurity()
	if err != nil {
		return fmt.Errorf("failed to read module security: %w", err)
	}

	for _, ms := range allMS {
		modName := h.GetModuleName(ms.ContainerID)
		if name.Module != "" && modName != name.Module {
			continue
		}
		for _, mr := range ms.ModuleRoles {
			if mr.Name == name.Name {
				fmt.Fprintf(e.output, "CREATE MODULE ROLE %s.%s", modName, mr.Name)
				if mr.Description != "" {
					fmt.Fprintf(e.output, " DESCRIPTION '%s'", mr.Description)
				}
				fmt.Fprintln(e.output, ";")
				fmt.Fprintln(e.output, "/")

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
						fmt.Fprintf(e.output, "\n-- Included in user roles: %s\n", strings.Join(includedBy, ", "))
					}
				}

				return nil
			}
		}
	}

	return fmt.Errorf("module role not found: %s", name)
}

// describeDemoUser handles DESCRIBE DEMO USER 'name'.
func (e *Executor) describeDemoUser(userName string) error {
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return fmt.Errorf("failed to read project security: %w", err)
	}

	for _, du := range ps.DemoUsers {
		if du.UserName == userName {
			fmt.Fprintf(e.output, "CREATE DEMO USER '%s' PASSWORD '***'", du.UserName)
			if du.Entity != "" {
				fmt.Fprintf(e.output, " ENTITY %s", du.Entity)
			}
			if len(du.UserRoles) > 0 {
				fmt.Fprintf(e.output, " (%s)", strings.Join(du.UserRoles, ", "))
			}
			fmt.Fprintln(e.output, ";")
			fmt.Fprintln(e.output, "/")
			return nil
		}
	}

	return fmt.Errorf("demo user not found: %s", userName)
}

// describeUserRole handles DESCRIBE USER ROLE Name.
func (e *Executor) describeUserRole(name ast.QualifiedName) error {
	ps, err := e.reader.GetProjectSecurity()
	if err != nil {
		return fmt.Errorf("failed to read project security: %w", err)
	}

	for _, ur := range ps.UserRoles {
		if ur.Name == name.Name {
			fmt.Fprintf(e.output, "CREATE USER ROLE %s", ur.Name)

			// Module roles
			if len(ur.ModuleRoles) > 0 {
				fmt.Fprintf(e.output, " (%s)", strings.Join(ur.ModuleRoles, ", "))
			}

			if ur.ManageAllRoles {
				fmt.Fprint(e.output, " MANAGE ALL ROLES")
			}

			fmt.Fprintln(e.output, ";")
			fmt.Fprintln(e.output, "/")

			// Show description if present
			if ur.Description != "" {
				fmt.Fprintf(e.output, "\n-- Description: %s\n", ur.Description)
			}

			// Show check security flag
			if ur.CheckSecurity {
				fmt.Fprintln(e.output, "-- Check security: enabled")
			}

			return nil
		}
	}

	return fmt.Errorf("user role not found: %s", name.Name)
}
