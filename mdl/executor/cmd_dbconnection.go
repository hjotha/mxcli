// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
)

// createDatabaseConnection handles CREATE DATABASE CONNECTION command.
func createDatabaseConnection(ctx *ExecContext, stmt *ast.CreateDatabaseConnectionStmt) error {
	e := ctx.executor

	if e.writer == nil {
		return mdlerrors.NewNotConnectedWrite()
	}

	if stmt.Name.Module == "" {
		return mdlerrors.NewValidation("module name required: use CREATE DATABASE CONNECTION Module.ConnectionName")
	}

	module, err := findModule(ctx, stmt.Name.Module)
	if err != nil {
		return err
	}

	// Check for existing connection
	existing, _ := e.reader.ListDatabaseConnections()
	h, _ := getHierarchy(ctx)

	for _, ex := range existing {
		modID := h.FindModuleID(ex.ContainerID)
		modName := h.GetModuleName(modID)
		if strings.EqualFold(modName, stmt.Name.Module) && strings.EqualFold(ex.Name, stmt.Name.Name) {
			if stmt.CreateOrModify {
				if err := e.writer.DeleteDatabaseConnection(ex.ID); err != nil {
					return mdlerrors.NewBackend("delete existing connection", err)
				}
			} else {
				return mdlerrors.NewAlreadyExistsMsg("database connection", modName+"."+ex.Name, fmt.Sprintf("database connection already exists: %s.%s (use CREATE OR MODIFY to update)", modName, ex.Name))
			}
		}
	}

	// Build connection string ref
	connStr := stmt.ConnectionString
	userName := stmt.UserName
	password := stmt.Password

	// Resolve ConnectionInput.Value from constant default (for Studio Pro dev)
	connInputValue := ""
	if stmt.ConnectionStringIsRef {
		connInputValue = resolveConstantDefault(ctx, connStr)
	}

	conn := &model.DatabaseConnection{
		ContainerID:          module.ID,
		Name:                 stmt.Name.Name,
		DatabaseType:         stmt.DatabaseType,
		ConnectionString:     connStr,
		ConnectionInputValue: connInputValue,
		UserName:             userName,
		Password:             password,
		ExportLevel:          "Hidden",
	}

	// Build queries
	for _, qDef := range stmt.Queries {
		q := &model.DatabaseQuery{
			Name:      qDef.Name,
			QueryType: 1, // custom SQL
			SQL:       qDef.SQL,
		}
		q.TypeName = "DatabaseConnector$DatabaseQuery"

		// Build parameters
		for _, pDef := range qDef.Parameters {
			p := &model.DatabaseQueryParameter{
				ParameterName:         pDef.Name,
				DefaultValue:          pDef.DefaultValue,
				EmptyValueBecomesNull: pDef.TestWithNull,
				DataType:              astDataTypeToDBType(pDef.DataType),
			}
			p.TypeName = "DatabaseConnector$QueryParameter"
			q.Parameters = append(q.Parameters, p)
		}

		// Build table mapping
		if qDef.Returns.String() != "" {
			tm := &model.DatabaseTableMapping{
				Entity: qDef.Returns.String(),
			}
			tm.TypeName = "DatabaseConnector$TableMapping"

			// Build column mappings from MAP clause
			for _, m := range qDef.Mappings {
				cm := &model.DatabaseColumnMapping{
					Attribute:  qDef.Returns.String() + "." + m.AttributeName,
					ColumnName: m.ColumnName,
				}
				cm.TypeName = "DatabaseConnector$ColumnMapping"
				tm.Columns = append(tm.Columns, cm)
			}

			q.TableMappings = append(q.TableMappings, tm)
		}

		conn.Queries = append(conn.Queries, q)
	}

	if err := e.writer.CreateDatabaseConnection(conn); err != nil {
		return mdlerrors.NewBackend("create database connection", err)
	}

	invalidateHierarchy(ctx)
	fmt.Fprintf(ctx.Output, "Created database connection: %s.%s\n", stmt.Name.Module, stmt.Name.Name)
	return nil
}

// showDatabaseConnections handles SHOW DATABASE CONNECTIONS command.
func showDatabaseConnections(ctx *ExecContext, moduleName string) error {
	e := ctx.executor

	connections, err := e.reader.ListDatabaseConnections()
	if err != nil {
		return mdlerrors.NewBackend("list database connections", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		dbType        string
		queries       int
	}
	var rows []row

	for _, conn := range connections {
		modID := h.FindModuleID(conn.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && !strings.EqualFold(modName, moduleName) {
			continue
		}

		qualifiedName := modName + "." + conn.Name
		folderPath := h.BuildFolderPath(conn.ContainerID)

		rows = append(rows, row{qualifiedName, modName, conn.Name, folderPath, conn.DatabaseType, len(conn.Queries)})
	}

	if len(rows) == 0 {
		fmt.Fprintln(ctx.Output, "No database connections found.")
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Folder", "Type", "Queries"},
		Summary: fmt.Sprintf("(%d database connections)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.folderPath, r.dbType, r.queries})
	}
	return writeResult(ctx, result)
}

// describeDatabaseConnection handles DESCRIBE DATABASE CONNECTION command.
func describeDatabaseConnection(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor

	connections, err := e.reader.ListDatabaseConnections()
	if err != nil {
		return mdlerrors.NewBackend("list database connections", err)
	}

	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	for _, conn := range connections {
		modID := h.FindModuleID(conn.ContainerID)
		modName := h.GetModuleName(modID)
		if strings.EqualFold(modName, name.Module) && strings.EqualFold(conn.Name, name.Name) {
			return outputDatabaseConnectionMDL(ctx, conn, modName)
		}
	}

	return mdlerrors.NewNotFound("database connection", name.String())
}

// outputDatabaseConnectionMDL outputs a database connection definition in MDL format.
func outputDatabaseConnectionMDL(ctx *ExecContext, conn *model.DatabaseConnection, moduleName string) error {
	fmt.Fprintf(ctx.Output, "CREATE DATABASE CONNECTION %s.%s\n", moduleName, conn.Name)
	fmt.Fprintf(ctx.Output, "TYPE '%s'\n", conn.DatabaseType)

	// Connection string
	fmt.Fprintf(ctx.Output, "CONNECTION STRING @%s\n", conn.ConnectionString)

	// Username
	fmt.Fprintf(ctx.Output, "USERNAME @%s\n", conn.UserName)

	// Password
	fmt.Fprintf(ctx.Output, "PASSWORD @%s\n", conn.Password)

	// Queries
	if len(conn.Queries) > 0 {
		fmt.Fprintln(ctx.Output, "BEGIN")
		for _, q := range conn.Queries {
			fmt.Fprintf(ctx.Output, "  QUERY %s\n", q.Name)

			// SQL string
			if q.SQL != "" {
				escaped := strings.ReplaceAll(q.SQL, "'", "''")
				fmt.Fprintf(ctx.Output, "    SQL '%s'\n", escaped)
			}

			// PARAMETER clauses
			for _, p := range q.Parameters {
				typeName := dbTypeToMDLType(p.DataType)
				if p.EmptyValueBecomesNull {
					fmt.Fprintf(ctx.Output, "    PARAMETER %s: %s NULL\n", p.ParameterName, typeName)
				} else if p.DefaultValue != "" {
					escaped := strings.ReplaceAll(p.DefaultValue, "'", "''")
					fmt.Fprintf(ctx.Output, "    PARAMETER %s: %s DEFAULT '%s'\n", p.ParameterName, typeName, escaped)
				} else {
					fmt.Fprintf(ctx.Output, "    PARAMETER %s: %s\n", p.ParameterName, typeName)
				}
			}

			// RETURNS and MAP from table mapping
			if len(q.TableMappings) > 0 {
				tm := q.TableMappings[0]
				fmt.Fprintf(ctx.Output, "    RETURNS %s\n", tm.Entity)

				// MAP clause
				if len(tm.Columns) > 0 {
					fmt.Fprintln(ctx.Output, "    MAP (")
					for i, c := range tm.Columns {
						// Extract attribute name from qualified ref (Module.Entity.Attr → Attr)
						attrName := c.Attribute
						if parts := strings.Split(attrName, "."); len(parts) >= 3 {
							attrName = parts[len(parts)-1]
						}
						sep := ","
						if i == len(tm.Columns)-1 {
							sep = ""
						}
						fmt.Fprintf(ctx.Output, "      %s AS %s%s\n", c.ColumnName, attrName, sep)
					}
					fmt.Fprintln(ctx.Output, "    )")
				}
			}
			fmt.Fprintln(ctx.Output, "  ;")
		}
		fmt.Fprintln(ctx.Output, "END")
	}

	fmt.Fprintln(ctx.Output, ";")
	fmt.Fprintln(ctx.Output, "/")

	return nil
}

// resolveConstantDefault looks up a constant by qualified name and returns its default value.
func resolveConstantDefault(ctx *ExecContext, qualifiedName string) string {
	e := ctx.executor

	constants, err := e.reader.ListConstants()
	if err != nil {
		return ""
	}
	h, err := getHierarchy(ctx)
	if err != nil {
		return ""
	}
	for _, c := range constants {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		fqn := modName + "." + c.Name
		if strings.EqualFold(fqn, qualifiedName) {
			return c.DefaultValue
		}
	}
	return ""
}

// astDataTypeToDBType converts an AST DataType to a BSON DataType string for DatabaseConnector.
func astDataTypeToDBType(dt ast.DataType) string {
	switch dt.Kind {
	case ast.TypeString:
		return "DataTypes$StringType"
	case ast.TypeInteger:
		return "DataTypes$IntegerType"
	case ast.TypeLong:
		return "DataTypes$IntegerType" // Long maps to IntegerType in DataTypes
	case ast.TypeDecimal:
		return "DataTypes$DecimalType"
	case ast.TypeBoolean:
		return "DataTypes$BooleanType"
	case ast.TypeDateTime, ast.TypeDate:
		return "DataTypes$DateTimeType"
	default:
		return "DataTypes$StringType"
	}
}

// dbTypeToMDLType converts a BSON DataType string to an MDL type name.
func dbTypeToMDLType(bsonType string) string {
	switch bsonType {
	case "DataTypes$StringType":
		return "String"
	case "DataTypes$IntegerType":
		return "Integer"
	case "DataTypes$DecimalType":
		return "Decimal"
	case "DataTypes$BooleanType":
		return "Boolean"
	case "DataTypes$DateTimeType":
		return "DateTime"
	default:
		return "String"
	}
}

// Executor wrappers for unmigrated callers.
