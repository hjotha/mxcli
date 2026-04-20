// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowDatabaseConnections_Mock(t *testing.T) {
	mod := mkModule("DataMod")
	conn := &model.DatabaseConnection{
		BaseElement:  model.BaseElement{ID: nextID("dbc")},
		ContainerID:  mod.ID,
		Name:         "MyDB",
		DatabaseType: "PostgreSQL",
	}

	h := mkHierarchy(mod)
	withContainer(h, conn.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:             func() bool { return true },
		ListDatabaseConnectionsFunc: func() ([]*model.DatabaseConnection, error) { return []*model.DatabaseConnection{conn}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showDatabaseConnections(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "DataMod.MyDB")
}

func TestDescribeDatabaseConnection_Mock(t *testing.T) {
	mod := mkModule("DataMod")
	conn := &model.DatabaseConnection{
		BaseElement:  model.BaseElement{ID: nextID("dbc")},
		ContainerID:  mod.ID,
		Name:         "MyDB",
		DatabaseType: "PostgreSQL",
	}

	h := mkHierarchy(mod)
	withContainer(h, conn.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:             func() bool { return true },
		ListDatabaseConnectionsFunc: func() ([]*model.DatabaseConnection, error) { return []*model.DatabaseConnection{conn}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeDatabaseConnection(ctx, ast.QualifiedName{Module: "DataMod", Name: "MyDB"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE DATABASE CONNECTION")
}
