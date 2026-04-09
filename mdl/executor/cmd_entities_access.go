// SPDX-License-Identifier: Apache-2.0

// Package executor - Entity access control (GRANT/REVOKE output and resolution)
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

// outputEntityAccessGrants outputs GRANT statements for entity access rules.
func (e *Executor) outputEntityAccessGrants(entity *domainmodel.Entity, moduleName, entityName string) {
	if len(entity.AccessRules) == 0 {
		return
	}

	// Build attribute name map for resolving member accesses
	attrNames := make(map[string]string)
	for _, attr := range entity.Attributes {
		attrNames[string(attr.ID)] = attr.Name
	}

	for _, rule := range entity.AccessRules {
		// Get role names
		var roleStrs []string
		for _, rn := range rule.ModuleRoleNames {
			roleStrs = append(roleStrs, rn)
		}
		if len(roleStrs) == 0 {
			for _, rid := range rule.ModuleRoles {
				roleStrs = append(roleStrs, string(rid))
			}
		}
		if len(roleStrs) == 0 {
			continue
		}

		rightsStr := e.formatAccessRuleRights(rule, attrNames)
		if rightsStr == "" {
			continue
		}

		grantLine := fmt.Sprintf("\nGRANT %s ON %s.%s (%s)",
			strings.Join(roleStrs, ", "), moduleName, entityName, rightsStr)

		if rule.XPathConstraint != "" {
			grantLine += fmt.Sprintf(" WHERE '%s'", rule.XPathConstraint)
		}
		grantLine += ";"

		fmt.Fprintln(e.output, grantLine)
	}
}

// resolveEntityMemberAccess determines per-member READ/WRITE access.
// Returns nil slices for "all members" (*), or specific member name lists.
func (e *Executor) resolveEntityMemberAccess(rule *domainmodel.AccessRule, attrNames map[string]string) (readMembers []string, writeMembers []string) {
	if len(rule.MemberAccesses) == 0 {
		// No per-member overrides: use default
		return nil, nil
	}

	// Check if all member accesses match the default — if so, treat as "*"
	allMatchDefault := true
	for _, ma := range rule.MemberAccesses {
		if ma.AccessRights != rule.DefaultMemberAccessRights {
			allMatchDefault = false
			break
		}
	}
	if allMatchDefault {
		return nil, nil
	}

	// Collect members by access level
	var readOnly, readWrite []string
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

		switch ma.AccessRights {
		case domainmodel.MemberAccessRightsReadWrite:
			readWrite = append(readWrite, memberName)
		case domainmodel.MemberAccessRightsReadOnly:
			readOnly = append(readOnly, memberName)
		}
	}

	// If there are overrides, list specific members for READ and WRITE
	// READ includes both ReadOnly and ReadWrite members
	allReadable := append(readOnly, readWrite...)
	if len(allReadable) == 0 {
		readMembers = nil // all via default
	} else {
		readMembers = allReadable
	}

	if len(readWrite) == 0 {
		writeMembers = []string{} // no write members
	} else {
		writeMembers = readWrite
	}

	return readMembers, writeMembers
}

// formatAccessRuleRights formats the rights portion of an access rule as a string.
// Returns a string like "CREATE, DELETE, READ (Name, Price), WRITE (Price)" or empty if no rights.
func (e *Executor) formatAccessRuleRights(rule *domainmodel.AccessRule, attrNames map[string]string) string {
	var rights []string
	if rule.AllowCreate {
		rights = append(rights, "CREATE")
	}
	if rule.AllowDelete {
		rights = append(rights, "DELETE")
	}

	hasRead := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadOnly ||
		rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
	hasWrite := rule.DefaultMemberAccessRights == domainmodel.MemberAccessRightsReadWrite
	if !hasRead || !hasWrite {
		for _, ma := range rule.MemberAccesses {
			if ma.AccessRights == domainmodel.MemberAccessRightsReadOnly ||
				ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
				hasRead = true
			}
			if ma.AccessRights == domainmodel.MemberAccessRightsReadWrite {
				hasWrite = true
			}
		}
	}

	readMembers, writeMembers := e.resolveEntityMemberAccess(rule, attrNames)

	if hasRead {
		if readMembers == nil {
			rights = append(rights, "READ *")
		} else {
			rights = append(rights, fmt.Sprintf("READ (%s)", strings.Join(readMembers, ", ")))
		}
	}
	if hasWrite {
		if writeMembers == nil {
			rights = append(rights, "WRITE *")
		} else if len(writeMembers) > 0 {
			rights = append(rights, fmt.Sprintf("WRITE (%s)", strings.Join(writeMembers, ", ")))
		}
	}

	return strings.Join(rights, ", ")
}

// formatAccessRuleResult re-reads the entity and formats the resulting access state
// for the given roles. Returns a string like "  Result: CREATE, READ (Name, Price)\n".
func (e *Executor) formatAccessRuleResult(moduleName, entityName string, roleNames []string) string {
	e.invalidateDomainModelsCache()

	module, err := e.findModule(moduleName)
	if err != nil {
		return ""
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return ""
	}

	entity := dm.FindEntityByName(entityName)
	if entity == nil {
		return ""
	}

	attrNames := make(map[string]string)
	for _, attr := range entity.Attributes {
		attrNames[string(attr.ID)] = attr.Name
	}

	// Build role set for matching
	roleSet := make(map[string]bool)
	for _, rn := range roleNames {
		roleSet[rn] = true
	}

	for _, rule := range entity.AccessRules {
		// Check if this rule matches the given roles
		matchCount := 0
		for _, rn := range rule.ModuleRoleNames {
			if roleSet[rn] {
				matchCount++
			}
		}
		if matchCount == 0 {
			continue
		}
		// Found a matching rule
		rightsStr := e.formatAccessRuleRights(rule, attrNames)
		if rightsStr == "" {
			return "  Result: (no access)\n"
		}
		return fmt.Sprintf("  Result: %s\n", rightsStr)
	}

	return "  Result: (no access)\n"
}
