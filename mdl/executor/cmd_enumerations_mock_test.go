// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

// TestShowEnumerations_Mock demonstrates testing a handler with a MockBackend
// instead of a real .mpr file. The handler under test is showEnumerations,
// which calls ctx.Backend.ListEnumerations() and writes a table to ctx.Output.
func TestShowEnumerations_Mock(t *testing.T) {
	modID := model.ID("mod-1")
	enumID := model.ID("enum-1")

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) {
			return []*model.Enumeration{
				{
					BaseElement: model.BaseElement{ID: enumID},
					ContainerID: modID,
					Name:        "Color",
					Values: []model.EnumerationValue{
						{Name: "Red"},
						{Name: "Green"},
						{Name: "Blue"},
					},
				},
			}, nil
		},
	}

	// Pre-populate hierarchy so getHierarchy skips the e.reader path.
	hierarchy := &ContainerHierarchy{
		moduleIDs:       map[model.ID]bool{modID: true},
		moduleNames:     map[model.ID]string{modID: "MyModule"},
		containerParent: map[model.ID]model.ID{enumID: modID},
		folderNames:     map[model.ID]string{},
	}

	var buf bytes.Buffer
	ctx := &ExecContext{
		Context: context.Background(),
		Backend: mb,
		Output:  &buf,
		Format:  FormatTable,
		Cache:   &executorCache{hierarchy: hierarchy},
	}

	if err := showEnumerations(ctx, ""); err != nil {
		t.Fatalf("showEnumerations returned error: %v", err)
	}

	out := buf.String()

	// Verify table contains our enumeration data.
	if !strings.Contains(out, "MyModule.Color") {
		t.Errorf("expected qualified name 'MyModule.Color' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("expected value count '3' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "(1 enumerations)") {
		t.Errorf("expected summary '(1 enumerations)' in output, got:\n%s", out)
	}
}

// TestShowEnumerations_Mock_FilterByModule verifies that passing a module name
// filters the output to only that module's enumerations.
func TestShowEnumerations_Mock_FilterByModule(t *testing.T) {
	mod1 := model.ID("mod-1")
	mod2 := model.ID("mod-2")

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) {
			return []*model.Enumeration{
				{
					BaseElement: model.BaseElement{ID: model.ID("e1")},
					ContainerID: mod1,
					Name:        "Color",
					Values:      []model.EnumerationValue{{Name: "Red"}},
				},
				{
					BaseElement: model.BaseElement{ID: model.ID("e2")},
					ContainerID: mod2,
					Name:        "Size",
					Values:      []model.EnumerationValue{{Name: "S"}, {Name: "M"}},
				},
			}, nil
		},
	}

	hierarchy := &ContainerHierarchy{
		moduleIDs:       map[model.ID]bool{mod1: true, mod2: true},
		moduleNames:     map[model.ID]string{mod1: "Alpha", mod2: "Beta"},
		containerParent: map[model.ID]model.ID{},
		folderNames:     map[model.ID]string{},
	}

	var buf bytes.Buffer
	ctx := &ExecContext{
		Context: context.Background(),
		Backend: mb,
		Output:  &buf,
		Format:  FormatTable,
		Cache:   &executorCache{hierarchy: hierarchy},
	}

	if err := showEnumerations(ctx, "Beta"); err != nil {
		t.Fatalf("showEnumerations returned error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Alpha.Color") {
		t.Errorf("should not contain Alpha.Color when filtering by Beta:\n%s", out)
	}
	if !strings.Contains(out, "Beta.Size") {
		t.Errorf("expected Beta.Size in output:\n%s", out)
	}
	if !strings.Contains(out, "(1 enumerations)") {
		t.Errorf("expected 1 enumeration in summary:\n%s", out)
	}
}
