// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
)

// emptyRegistry creates a Registry with no handlers registered.
func emptyRegistry() *Registry {
	return &Registry{handlers: make(map[reflect.Type]StmtHandler)}
}

func TestNewRegistry_NoPanic(t *testing.T) {
	// Smoke test: constructing a registry with all stub registrations
	// must not panic.
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
}

func TestRegistry_Dispatch_UnknownStatement(t *testing.T) {
	r := emptyRegistry()

	// ConnectStmt is not registered — Dispatch must return UnsupportedError.
	err := r.Dispatch(nil, &ast.ConnectStmt{Path: "/tmp/test.mpr"})
	if err == nil {
		t.Fatal("expected error for unregistered statement, got nil")
	}
	var unsupported *mdlerrors.UnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected UnsupportedError, got %T: %v", err, err)
	}
}

func TestRegistry_Register_Duplicate_Panics(t *testing.T) {
	r := emptyRegistry()
	handler := func(e *Executor, stmt ast.Statement) error { return nil }

	r.Register(&ast.ConnectStmt{}, handler)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.Register(&ast.ConnectStmt{}, handler)
}

func TestRegistry_Dispatch_Success(t *testing.T) {
	r := emptyRegistry()
	called := false
	r.Register(&ast.ConnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		called = true
		if _, ok := stmt.(*ast.ConnectStmt); !ok {
			t.Fatalf("expected *ConnectStmt, got %T", stmt)
		}
		return nil
	})

	err := r.Dispatch(nil, &ast.ConnectStmt{Path: "/tmp/test.mpr"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRegistry_Dispatch_HandlerError(t *testing.T) {
	r := emptyRegistry()
	sentinel := errors.New("test error")
	r.Register(&ast.ConnectStmt{}, func(e *Executor, stmt ast.Statement) error {
		return sentinel
	})

	err := r.Dispatch(nil, &ast.ConnectStmt{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

func TestRegistry_Validate_Empty(t *testing.T) {
	r := emptyRegistry()

	knownTypes := []ast.Statement{
		&ast.ConnectStmt{},
		&ast.DisconnectStmt{},
	}
	err := r.Validate(knownTypes)
	if err == nil {
		t.Fatal("expected validation error for empty registry")
	}
}

func TestRegistry_Validate_Complete(t *testing.T) {
	r := emptyRegistry()
	noop := func(e *Executor, stmt ast.Statement) error { return nil }
	r.Register(&ast.ConnectStmt{}, noop)
	r.Register(&ast.DisconnectStmt{}, noop)

	knownTypes := []ast.Statement{
		&ast.ConnectStmt{},
		&ast.DisconnectStmt{},
	}
	err := r.Validate(knownTypes)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRegistry_Validate_Partial(t *testing.T) {
	r := emptyRegistry()
	noop := func(e *Executor, stmt ast.Statement) error { return nil }
	r.Register(&ast.ConnectStmt{}, noop)

	knownTypes := []ast.Statement{
		&ast.ConnectStmt{},
		&ast.DisconnectStmt{},
		&ast.StatusStmt{},
	}
	err := r.Validate(knownTypes)
	if err == nil {
		t.Fatal("expected validation error for partial registry")
	}
	// Should mention 2 missing types.
	if got := err.Error(); !strings.Contains(got, "2 unregistered") {
		t.Fatalf("expected '2 unregistered' in error, got: %s", got)
	}
}

func TestRegistry_HandlerCount(t *testing.T) {
	r := emptyRegistry()
	if r.HandlerCount() != 0 {
		t.Fatalf("expected 0, got %d", r.HandlerCount())
	}
	noop := func(e *Executor, stmt ast.Statement) error { return nil }
	r.Register(&ast.ConnectStmt{}, noop)
	if r.HandlerCount() != 1 {
		t.Fatalf("expected 1, got %d", r.HandlerCount())
	}
}
