// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// showBusinessEventServices displays a table of all business event service documents.
func (e *Executor) showBusinessEventServices(inModule string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	services, err := e.reader.ListBusinessEventServices()
	if err != nil {
		return fmt.Errorf("failed to list business event services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	var filtered []*model.BusinessEventService
	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		moduleName := h.GetModuleName(modID)
		if inModule != "" && !strings.EqualFold(moduleName, inModule) {
			continue
		}
		filtered = append(filtered, svc)
	}

	if len(filtered) == 0 {
		if inModule != "" {
			fmt.Fprintf(e.output, "No business event services found in module %s\n", inModule)
		} else {
			fmt.Fprintln(e.output, "No business event services found")
		}
		return nil
	}

	type row struct {
		module, qualifiedName, name            string
		msgCount, publishCount, subscribeCount int
	}
	var rows []row
	modWidth, qnWidth, nameWidth := len("Module"), len("QualifiedName"), len("Service")

	for _, svc := range filtered {
		modID := h.FindModuleID(svc.ContainerID)
		moduleName := h.GetModuleName(modID)
		qn := moduleName + "." + svc.Name
		r := row{module: moduleName, qualifiedName: qn, name: svc.Name}

		if svc.Definition != nil {
			for _, ch := range svc.Definition.Channels {
				r.msgCount += len(ch.Messages)
			}
		}
		for _, op := range svc.OperationImplementations {
			switch op.Operation {
			case "publish":
				r.publishCount++
			case "subscribe":
				r.subscribeCount++
			}
		}

		if len(moduleName) > modWidth {
			modWidth = len(moduleName)
		}
		if len(qn) > qnWidth {
			qnWidth = len(qn)
		}
		if len(svc.Name) > nameWidth {
			nameWidth = len(svc.Name)
		}
		rows = append(rows, r)
	}

	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-10s | %-10s | %-10s |\n",
		modWidth, "Module", qnWidth, "QualifiedName", nameWidth, "Service", "Messages", "Publish", "Subscribe")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", modWidth), strings.Repeat("-", qnWidth), strings.Repeat("-", nameWidth),
		strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 10))

	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %10d | %10d | %10d |\n",
			modWidth, r.module, qnWidth, r.qualifiedName, nameWidth, r.name,
			r.msgCount, r.publishCount, r.subscribeCount)
	}

	fmt.Fprintf(e.output, "\n(%d business event services)\n", len(filtered))
	return nil
}

// showBusinessEventClients displays a table of all business event client documents.
func (e *Executor) showBusinessEventClients(inModule string) error {
	fmt.Fprintln(e.output, "Business event clients are not yet implemented.")
	return nil
}

// showBusinessEvents displays a table of individual messages across all business event services.
func (e *Executor) showBusinessEvents(inModule string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	services, err := e.reader.ListBusinessEventServices()
	if err != nil {
		return fmt.Errorf("failed to list business event services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	type row struct {
		service, message, operation, entity string
		attrs                               int
	}
	var rows []row
	svcWidth := len("Service")
	msgWidth := len("Message")
	opWidth := len("Operation")
	entityWidth := len("Entity")

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		moduleName := h.GetModuleName(modID)
		if inModule != "" && !strings.EqualFold(moduleName, inModule) {
			continue
		}

		svcQN := moduleName + "." + svc.Name

		// Build operation map: messageName -> ServiceOperation
		opMap := make(map[string]*model.ServiceOperation)
		for _, op := range svc.OperationImplementations {
			opMap[op.MessageName] = op
		}

		if svc.Definition != nil {
			for _, ch := range svc.Definition.Channels {
				for _, msg := range ch.Messages {
					opStr := ""
					entityStr := ""
					if op, ok := opMap[msg.MessageName]; ok {
						opStr = strings.ToUpper(op.Operation)
						entityStr = op.Entity
					}
					r := row{
						service:   svcQN,
						message:   msg.MessageName,
						operation: opStr,
						entity:    entityStr,
						attrs:     len(msg.Attributes),
					}
					if len(svcQN) > svcWidth {
						svcWidth = len(svcQN)
					}
					if len(msg.MessageName) > msgWidth {
						msgWidth = len(msg.MessageName)
					}
					if len(opStr) > opWidth {
						opWidth = len(opStr)
					}
					if len(entityStr) > entityWidth {
						entityWidth = len(entityStr)
					}
					rows = append(rows, r)
				}
			}
		}
	}

	if len(rows) == 0 {
		if inModule != "" {
			fmt.Fprintf(e.output, "No business events found in module %s\n", inModule)
		} else {
			fmt.Fprintln(e.output, "No business events found")
		}
		return nil
	}

	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-10s |\n",
		svcWidth, "Service", msgWidth, "Message", opWidth, "Operation", entityWidth, "Entity", "Attributes")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", svcWidth), strings.Repeat("-", msgWidth),
		strings.Repeat("-", opWidth), strings.Repeat("-", entityWidth), strings.Repeat("-", 10))

	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %10d |\n",
			svcWidth, r.service, msgWidth, r.message, opWidth, r.operation,
			entityWidth, r.entity, r.attrs)
	}

	fmt.Fprintf(e.output, "\n(%d business events)\n", len(rows))
	return nil
}

// describeBusinessEventService outputs the full MDL description of a business event service.
func (e *Executor) describeBusinessEventService(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	services, err := e.reader.ListBusinessEventServices()
	if err != nil {
		return fmt.Errorf("failed to list business event services: %w", err)
	}

	// Use hierarchy to resolve container IDs to module names
	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	// Find the service by qualified name
	var found *model.BusinessEventService
	var foundModule string
	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		moduleName := h.GetModuleName(modID)
		if strings.EqualFold(moduleName, name.Module) && strings.EqualFold(svc.Name, name.Name) {
			found = svc
			foundModule = moduleName
			break
		}
	}

	if found == nil {
		return fmt.Errorf("business event service not found: %s", name)
	}

	// Output MDL CREATE statement
	if found.Documentation != "" {
		outputJavadoc(e.output, found.Documentation)
	}
	fmt.Fprintf(e.output, "CREATE OR REPLACE BUSINESS EVENT SERVICE %s.%s\n", foundModule, found.Name)

	if found.Definition != nil {
		fmt.Fprintf(e.output, "(\n")
		fmt.Fprintf(e.output, "  ServiceName: '%s'", found.Definition.ServiceName)
		if found.Definition.EventNamePrefix != "" {
			fmt.Fprintf(e.output, ",\n  EventNamePrefix: '%s'", found.Definition.EventNamePrefix)
		} else {
			fmt.Fprintf(e.output, ",\n  EventNamePrefix: ''")
		}
		fmt.Fprintf(e.output, "\n)\n")

		fmt.Fprintf(e.output, "{\n")

		// Build operation map: messageName -> operation info
		opMap := make(map[string]*model.ServiceOperation)
		for _, op := range found.OperationImplementations {
			opMap[op.MessageName] = op
		}

		// Output messages
		for _, ch := range found.Definition.Channels {
			for _, msg := range ch.Messages {
				// Format attributes
				var attrs []string
				for _, a := range msg.Attributes {
					attrs = append(attrs, fmt.Sprintf("%s: %s", a.AttributeName, a.AttributeType))
				}

				// Determine operation from OperationImplementations
				opStr := "PUBLISH"
				entityStr := ""
				if op, ok := opMap[msg.MessageName]; ok {
					if op.Operation == "subscribe" {
						opStr = "SUBSCRIBE"
					}
					if op.Entity != "" {
						entityStr = fmt.Sprintf("\n    ENTITY %s", op.Entity)
					}
				}

				fmt.Fprintf(e.output, "  MESSAGE %s (%s) %s%s;\n",
					msg.MessageName, strings.Join(attrs, ", "), opStr, entityStr)
			}
		}

		fmt.Fprintf(e.output, "};\n")
	}

	return nil
}

// createBusinessEventService creates a new business event service from an AST statement.
func (e *Executor) createBusinessEventService(stmt *ast.CreateBusinessEventServiceStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project (read-only mode)")
	}

	moduleName := stmt.Name.Module
	module, err := e.findModule(moduleName)
	if err != nil {
		return fmt.Errorf("module not found: %s", moduleName)
	}

	// Check for existing service with same name (if not CREATE OR REPLACE)
	existingServices, _ := e.reader.ListBusinessEventServices()
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, existing := range existingServices {
		existModID := h.FindModuleID(existing.ContainerID)
		existModName := h.GetModuleName(existModID)
		if strings.EqualFold(existModName, moduleName) && strings.EqualFold(existing.Name, stmt.Name.Name) {
			if stmt.CreateOrReplace {
				// Delete existing
				if err := e.writer.DeleteBusinessEventService(existing.ID); err != nil {
					return fmt.Errorf("failed to delete existing service: %w", err)
				}
			} else {
				return fmt.Errorf("business event service already exists: %s.%s (use CREATE OR REPLACE to overwrite)", moduleName, stmt.Name.Name)
			}
		}
	}

	// Resolve folder if specified
	containerID := module.ID
	if stmt.Folder != "" {
		folderID, err := e.resolveFolder(module.ID, stmt.Folder)
		if err != nil {
			return fmt.Errorf("failed to resolve folder '%s': %w", stmt.Folder, err)
		}
		containerID = folderID
	}

	// Build the service from AST
	svc := &model.BusinessEventService{
		ContainerID:   containerID,
		Name:          stmt.Name.Name,
		Documentation: stmt.Documentation,
		ExportLevel:   "Hidden",
	}

	// Build definition
	def := &model.BusinessEventDefinition{
		ServiceName:     stmt.ServiceName,
		EventNamePrefix: stmt.EventNamePrefix,
	}
	def.TypeName = "BusinessEvents$BusinessEventDefinition"

	// Create channels (one per message in our simplified model)
	for _, msgDef := range stmt.Messages {
		ch := &model.BusinessEventChannel{
			ChannelName: generateChannelName(),
		}
		ch.TypeName = "BusinessEvents$Channel"

		msg := &model.BusinessEventMessage{
			MessageName: msgDef.MessageName,
		}
		msg.TypeName = "BusinessEvents$Message"

		// Set publish/subscribe based on operation
		switch strings.ToUpper(msgDef.Operation) {
		case "PUBLISH":
			msg.CanSubscribe = true // Service publishes → others subscribe
		case "SUBSCRIBE":
			msg.CanPublish = true // Service subscribes → others publish
		}

		// Build attributes
		for _, attrDef := range msgDef.Attributes {
			attr := &model.BusinessEventAttribute{
				AttributeName: attrDef.Name,
				AttributeType: attrDef.TypeName,
			}
			attr.TypeName = "BusinessEvents$MessageAttribute"
			msg.Attributes = append(msg.Attributes, attr)
		}

		ch.Messages = append(ch.Messages, msg)
		def.Channels = append(def.Channels, ch)

		// Create operation implementation
		op := &model.ServiceOperation{
			MessageName: msgDef.MessageName,
			Operation:   strings.ToLower(msgDef.Operation),
			Entity:      msgDef.Entity,
			Microflow:   msgDef.Microflow,
		}
		op.TypeName = "BusinessEvents$ServiceOperation"
		svc.OperationImplementations = append(svc.OperationImplementations, op)
	}

	svc.Definition = def

	// Write to project
	if err := e.writer.CreateBusinessEventService(svc); err != nil {
		return fmt.Errorf("failed to create business event service: %w", err)
	}

	fmt.Fprintf(e.output, "Created business event service: %s.%s\n", moduleName, stmt.Name.Name)
	return nil
}

// dropBusinessEventService deletes a business event service.
func (e *Executor) dropBusinessEventService(stmt *ast.DropBusinessEventServiceStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project (read-only mode)")
	}

	services, err := e.reader.ListBusinessEventServices()
	if err != nil {
		return fmt.Errorf("failed to list business event services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		moduleName := h.GetModuleName(modID)
		if strings.EqualFold(moduleName, stmt.Name.Module) && strings.EqualFold(svc.Name, stmt.Name.Name) {
			if err := e.writer.DeleteBusinessEventService(svc.ID); err != nil {
				return fmt.Errorf("failed to delete business event service: %w", err)
			}
			fmt.Fprintf(e.output, "Dropped business event service: %s.%s\n", moduleName, svc.Name)
			return nil
		}
	}

	return fmt.Errorf("business event service not found: %s", stmt.Name)
}

// generateChannelName generates a hex channel name (similar to Mendix Studio Pro).
func generateChannelName() string {
	// Generate a UUID-like hex string
	uuid := mpr.GenerateID()
	return strings.ReplaceAll(uuid, "-", "")
}
