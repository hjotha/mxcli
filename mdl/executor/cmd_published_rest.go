// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
)

// showPublishedRestServices handles SHOW PUBLISHED REST SERVICES [IN module] command.
func (e *Executor) showPublishedRestServices(moduleName string) error {
	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return mdlerrors.NewBackend("list published REST services", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	type row struct {
		module        string
		qualifiedName string
		path          string
		version       string
		resources     int
		operations    int
	}
	var rows []row

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && !strings.EqualFold(modName, moduleName) {
			continue
		}

		qn := modName + "." + svc.Name
		opCount := 0
		for _, res := range svc.Resources {
			opCount += len(res.Operations)
		}

		path := svc.Path
		if len(path) > 50 {
			path = path[:47] + "..."
		}

		rows = append(rows, row{modName, qn, path, svc.Version, len(svc.Resources), opCount})
	}

	if len(rows) == 0 {
		fmt.Fprintln(e.output, "No published REST services found.")
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Module", "QualifiedName", "Path", "Version", "Resources", "Operations"},
		Summary: fmt.Sprintf("(%d published REST services)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.module, r.qualifiedName, r.path, r.version, r.resources, r.operations})
	}
	return e.writeResult(result)
}

// describePublishedRestService handles DESCRIBE PUBLISHED REST SERVICE command.
func (e *Executor) describePublishedRestService(name ast.QualifiedName) error {
	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return mdlerrors.NewBackend("list published REST services", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		qualifiedName := modName + "." + svc.Name

		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(svc.Name, name.Name) {
			continue
		}

		// Output as re-executable MDL
		fmt.Fprintf(e.output, "CREATE PUBLISHED REST SERVICE %s (\n", qualifiedName)
		fmt.Fprintf(e.output, "  Path: '%s'", svc.Path)
		if svc.Version != "" {
			fmt.Fprintf(e.output, ",\n  Version: '%s'", svc.Version)
		}
		if svc.ServiceName != "" {
			fmt.Fprintf(e.output, ",\n  ServiceName: '%s'", svc.ServiceName)
		}
		folderPath := h.BuildFolderPath(svc.ContainerID)
		if folderPath != "" {
			fmt.Fprintf(e.output, ",\n  Folder: '%s'", folderPath)
		}
		fmt.Fprintln(e.output, "\n)")

		if len(svc.Resources) > 0 {
			fmt.Fprintln(e.output, "{")
			for _, res := range svc.Resources {
				fmt.Fprintf(e.output, "  RESOURCE '%s' {\n", res.Name)
				for _, op := range res.Operations {
					deprecated := ""
					if op.Deprecated {
						deprecated = " DEPRECATED"
					}
					mf := ""
					if op.Microflow != "" {
						mf = fmt.Sprintf(" MICROFLOW %s", op.Microflow)
					}
					summary := ""
					if op.Summary != "" {
						summary = fmt.Sprintf(" -- %s", op.Summary)
					}
					opPath := ""
					if op.Path != "" {
						opPath = fmt.Sprintf(" '%s'", op.Path)
					}
					fmt.Fprintf(e.output, "    %s%s%s%s;%s\n",
						strings.ToUpper(op.HTTPMethod), opPath, mf, deprecated, summary)
				}
				fmt.Fprintln(e.output, "  }")
			}
			fmt.Fprintln(e.output, "};")
		} else {
			fmt.Fprintln(e.output, ";")
		}
		fmt.Fprintln(e.output, "/")

		// Emit GRANT statements for any module roles with access.
		if len(svc.AllowedRoles) > 0 {
			fmt.Fprintf(e.output, "\nGRANT ACCESS ON PUBLISHED REST SERVICE %s.%s TO %s;\n",
				modName, svc.Name, strings.Join(svc.AllowedRoles, ", "))
		}

		return nil
	}

	return mdlerrors.NewNotFound("published REST service", name.String())
}

// findPublishedRestService looks up a published REST service by module and name.
func (e *Executor) findPublishedRestService(moduleName, name string) (*model.PublishedRestService, error) {
	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return nil, err
	}
	h, err := e.getHierarchy()
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == moduleName && svc.Name == name {
			return svc, nil
		}
	}
	return nil, mdlerrors.NewNotFound("published REST service", moduleName+"."+name)
}

// execCreatePublishedRestService creates a new published REST service.
func (e *Executor) execCreatePublishedRestService(s *ast.CreatePublishedRestServiceStmt) error {
	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	if err := e.checkFeature("integration", "published_rest_service",
		"CREATE PUBLISHED REST SERVICE",
		"upgrade your project to 10.0+"); err != nil {
		return err
	}

	// Handle CREATE OR REPLACE — delete existing if found
	if s.CreateOrReplace {
		existing, findErr := e.findPublishedRestService(s.Name.Module, s.Name.Name)
		var nfe *mdlerrors.NotFoundError
		if findErr != nil && !errors.As(findErr, &nfe) {
			return mdlerrors.NewBackend("find existing service", findErr)
		}
		if existing != nil {
			if err := e.writer.DeletePublishedRestService(existing.ID); err != nil {
				return mdlerrors.NewBackend("replace existing service", err)
			}
		}
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return mdlerrors.NewNotFound("module", s.Name.Module)
	}

	containerID := module.ID
	if s.Folder != "" {
		folderID, err := e.resolveFolder(module.ID, s.Folder)
		if err != nil {
			return mdlerrors.NewBackend(fmt.Sprintf("resolve folder '%s'", s.Folder), err)
		}
		containerID = folderID
	}

	svc := &model.PublishedRestService{
		ContainerID: containerID,
		Name:        s.Name.Name,
		Path:        s.Path,
		Version:     s.Version,
		ServiceName: s.ServiceName,
	}

	for _, resDef := range s.Resources {
		resource := &model.PublishedRestResource{
			Name: resDef.Name,
		}
		for _, opDef := range resDef.Operations {
			op := &model.PublishedRestOperation{
				HTTPMethod: opDef.HTTPMethod,
				Path:       opDef.Path,
				Microflow:  opDef.Microflow.String(),
				Summary:    "",
				Deprecated: opDef.Deprecated,
			}
			resource.Operations = append(resource.Operations, op)
		}
		svc.Resources = append(svc.Resources, resource)
	}

	if err := e.writer.CreatePublishedRestService(svc); err != nil {
		return mdlerrors.NewBackend("create published REST service", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Created published REST service %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}

// execDropPublishedRestService deletes a published REST service.
func (e *Executor) execDropPublishedRestService(s *ast.DropPublishedRestServiceStmt) error {
	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return mdlerrors.NewBackend("list published REST services", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == s.Name.Module && svc.Name == s.Name.Name {
			if err := e.writer.DeletePublishedRestService(svc.ID); err != nil {
				return mdlerrors.NewBackend("drop published REST service", err)
			}
			if !e.quiet {
				fmt.Fprintf(e.output, "Dropped published REST service %s.%s\n", s.Name.Module, s.Name.Name)
			}
			return nil
		}
	}

	return mdlerrors.NewNotFound("published REST service", s.Name.Module+"."+s.Name.Name)
}

// astResourceDefToModel converts an AST PublishedRestResourceDef to the
// runtime model type used by the writer.
func astResourceDefToModel(def *ast.PublishedRestResourceDef) *model.PublishedRestResource {
	resource := &model.PublishedRestResource{Name: def.Name}
	for _, opDef := range def.Operations {
		resource.Operations = append(resource.Operations, &model.PublishedRestOperation{
			HTTPMethod: opDef.HTTPMethod,
			Path:       opDef.Path,
			Microflow:  opDef.Microflow.String(),
			Deprecated: opDef.Deprecated,
		})
	}
	return resource
}

// execAlterPublishedRestService applies SET / ADD RESOURCE / DROP RESOURCE
// actions to an existing published REST service.
func (e *Executor) execAlterPublishedRestService(s *ast.AlterPublishedRestServiceStmt) error {
	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	if err := e.checkFeature("integration", "published_rest_alter",
		"ALTER PUBLISHED REST SERVICE",
		"upgrade your project to 10.0+"); err != nil {
		return err
	}

	svc, err := e.findPublishedRestService(s.Name.Module, s.Name.Name)
	if err != nil {
		return err
	}

	for _, action := range s.Actions {
		switch a := action.(type) {
		case *ast.PublishedRestSetAction:
			for key, val := range a.Changes {
				switch strings.ToLower(key) {
				case "path":
					svc.Path = val
				case "version":
					svc.Version = val
				case "servicename":
					svc.ServiceName = val
				default:
					return mdlerrors.NewUnsupported(fmt.Sprintf("unknown published REST service property: %s (allowed: Path, Version, ServiceName)", key))
				}
			}

		case *ast.PublishedRestAddResourceAction:
			// Reject duplicate resource names
			for _, existing := range svc.Resources {
				if existing.Name == a.Resource.Name {
					return mdlerrors.NewAlreadyExistsMsg("resource", a.Resource.Name, fmt.Sprintf("resource '%s' already exists on %s.%s", a.Resource.Name, s.Name.Module, s.Name.Name))
				}
			}
			svc.Resources = append(svc.Resources, astResourceDefToModel(a.Resource))

		case *ast.PublishedRestDropResourceAction:
			idx := -1
			for i, existing := range svc.Resources {
				if existing.Name == a.Name {
					idx = i
					break
				}
			}
			if idx == -1 {
				return mdlerrors.NewNotFoundMsg("resource", a.Name, fmt.Sprintf("resource '%s' not found on %s.%s", a.Name, s.Name.Module, s.Name.Name))
			}
			svc.Resources = append(svc.Resources[:idx], svc.Resources[idx+1:]...)

		default:
			return mdlerrors.NewUnsupported(fmt.Sprintf("unsupported alter action: %T", action))
		}
	}

	if err := e.writer.UpdatePublishedRestService(svc); err != nil {
		return mdlerrors.NewBackend("alter published REST service", err)
	}

	if !e.quiet {
		fmt.Fprintf(e.output, "Altered published REST service %s.%s\n", s.Name.Module, s.Name.Name)
	}
	return nil
}
