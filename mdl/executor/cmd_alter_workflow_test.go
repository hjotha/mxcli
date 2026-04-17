// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"io"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

// makeWorkflowDoc builds a minimal workflow BSON document for testing.
// Activities are placed inside a Flow sub-document, matching real workflow structure.
func makeWorkflowDoc(activities ...bson.D) bson.D {
	actArr := bson.A{int32(3)} // Mendix array marker
	for _, a := range activities {
		actArr = append(actArr, a)
	}
	return bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: "Workflows$Workflow"},
		{Key: "Title", Value: "Test Workflow"},
		{Key: "WorkflowName", Value: bson.D{
			{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
			{Key: "$Type", Value: "Texts$Text"},
			{Key: "Text", Value: "Test Workflow"},
		}},
		{Key: "WorkflowDescription", Value: bson.D{
			{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
			{Key: "$Type", Value: "Texts$Text"},
			{Key: "Text", Value: "Original description"},
		}},
		{Key: "Flow", Value: bson.D{
			{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
			{Key: "$Type", Value: "Workflows$Flow"},
			{Key: "Activities", Value: actArr},
		}},
	}
}

// makeWorkflowActivity builds a minimal workflow activity BSON with a caption and name.
func makeWorkflowActivity(typeName, caption, name string) bson.D {
	return bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: typeName},
		{Key: "Caption", Value: caption},
		{Key: "Name", Value: name},
	}
}

// makeWorkflowActivityWithBoundaryEvents builds an activity with boundary events.
func makeWorkflowActivityWithBoundaryEvents(caption string, events ...bson.D) bson.D {
	evtArr := bson.A{int32(3)}
	for _, e := range events {
		evtArr = append(evtArr, e)
	}
	return bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: "Workflows$UserTask"},
		{Key: "Caption", Value: caption},
		{Key: "Name", Value: "task1"},
		{Key: "BoundaryEvents", Value: evtArr},
	}
}

func makeBoundaryEvent(typeName string) bson.D {
	return bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: typeName},
		{Key: "Caption", Value: ""},
	}
}

// --- SET DISPLAY tests ---

func TestSetWorkflowProperty_Display(t *testing.T) {
	doc := makeWorkflowDoc()

	op := &ast.SetWorkflowPropertyOp{Property: "DISPLAY", Value: "New Title"}
	if err := applySetWorkflowProperty(&doc, op); err != nil {
		t.Fatalf("SET DISPLAY failed: %v", err)
	}

	// Title should be updated
	if got := dGetString(doc, "Title"); got != "New Title" {
		t.Errorf("Title = %q, want %q", got, "New Title")
	}
	// WorkflowName.Text should be updated
	wfName := dGetDoc(doc, "WorkflowName")
	if wfName == nil {
		t.Fatal("WorkflowName is nil")
	}
	if got := dGetString(wfName, "Text"); got != "New Title" {
		t.Errorf("WorkflowName.Text = %q, want %q", got, "New Title")
	}
}

func TestSetWorkflowProperty_Display_NilSubDoc(t *testing.T) {
	// Build doc without WorkflowName to test auto-creation
	doc := bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: "Workflows$Workflow"},
		{Key: "Title", Value: "Old"},
		{Key: "Flow", Value: bson.D{
			{Key: "Activities", Value: bson.A{int32(3)}},
		}},
	}

	op := &ast.SetWorkflowPropertyOp{Property: "DISPLAY", Value: "Created Title"}
	if err := applySetWorkflowProperty(&doc, op); err != nil {
		t.Fatalf("SET DISPLAY with nil sub-doc failed: %v", err)
	}

	if got := dGetString(doc, "Title"); got != "Created Title" {
		t.Errorf("Title = %q, want %q", got, "Created Title")
	}

	wfName := dGetDoc(doc, "WorkflowName")
	if wfName == nil {
		t.Fatal("WorkflowName should have been auto-created")
	}
	if got := dGetString(wfName, "Text"); got != "Created Title" {
		t.Errorf("WorkflowName.Text = %q, want %q", got, "Created Title")
	}
	if got := dGetString(wfName, "$Type"); got != "Texts$Text" {
		t.Errorf("WorkflowName.$Type = %q, want %q", got, "Texts$Text")
	}
}

// --- SET DESCRIPTION tests ---

func TestSetWorkflowProperty_Description(t *testing.T) {
	doc := makeWorkflowDoc()

	op := &ast.SetWorkflowPropertyOp{Property: "DESCRIPTION", Value: "Updated desc"}
	if err := applySetWorkflowProperty(&doc, op); err != nil {
		t.Fatalf("SET DESCRIPTION failed: %v", err)
	}

	wfDesc := dGetDoc(doc, "WorkflowDescription")
	if wfDesc == nil {
		t.Fatal("WorkflowDescription is nil")
	}
	if got := dGetString(wfDesc, "Text"); got != "Updated desc" {
		t.Errorf("WorkflowDescription.Text = %q, want %q", got, "Updated desc")
	}
}

func TestSetWorkflowProperty_Description_NilSubDoc(t *testing.T) {
	doc := bson.D{
		{Key: "$ID", Value: primitive.Binary{Subtype: 0x04, Data: make([]byte, 16)}},
		{Key: "$Type", Value: "Workflows$Workflow"},
		{Key: "Title", Value: "Test"},
		{Key: "Flow", Value: bson.D{
			{Key: "Activities", Value: bson.A{int32(3)}},
		}},
	}

	op := &ast.SetWorkflowPropertyOp{Property: "DESCRIPTION", Value: "New desc"}
	if err := applySetWorkflowProperty(&doc, op); err != nil {
		t.Fatalf("SET DESCRIPTION with nil sub-doc failed: %v", err)
	}

	wfDesc := dGetDoc(doc, "WorkflowDescription")
	if wfDesc == nil {
		t.Fatal("WorkflowDescription should have been auto-created")
	}
	if got := dGetString(wfDesc, "Text"); got != "New desc" {
		t.Errorf("WorkflowDescription.Text = %q, want %q", got, "New desc")
	}
}

// --- SET unsupported property ---

func TestSetWorkflowProperty_UnsupportedProperty(t *testing.T) {
	doc := makeWorkflowDoc()

	op := &ast.SetWorkflowPropertyOp{Property: "UNKNOWN_PROP", Value: "x"}
	err := applySetWorkflowProperty(&doc, op)
	if err == nil {
		t.Fatal("Expected error for unsupported property")
	}
	if !strings.Contains(err.Error(), "unsupported workflow property") {
		t.Errorf("Error = %q, want to contain 'unsupported workflow property'", err.Error())
	}
}

// --- findActivityByCaption tests ---

func TestFindActivityByCaption_Found(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Approve", "task2")
	doc := makeWorkflowDoc(act1, act2)

	result, err := findActivityByCaption(doc, "Approve", 0)
	if err != nil {
		t.Fatalf("findActivityByCaption failed: %v", err)
	}
	if got := dGetString(result, "Caption"); got != "Approve" {
		t.Errorf("Caption = %q, want %q", got, "Approve")
	}
}

func TestFindActivityByCaption_ByName(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "ReviewTask")
	doc := makeWorkflowDoc(act1)

	result, err := findActivityByCaption(doc, "ReviewTask", 0)
	if err != nil {
		t.Fatalf("findActivityByCaption by name failed: %v", err)
	}
	if got := dGetString(result, "Name"); got != "ReviewTask" {
		t.Errorf("Name = %q, want %q", got, "ReviewTask")
	}
}

func TestFindActivityByCaption_NotFound(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	doc := makeWorkflowDoc(act1)

	_, err := findActivityByCaption(doc, "NonExistent", 0)
	if err == nil {
		t.Fatal("Expected error for missing activity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error = %q, want to contain 'not found'", err.Error())
	}
}

func TestFindActivityByCaption_Ambiguous(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Review", "task2")
	doc := makeWorkflowDoc(act1, act2)

	_, err := findActivityByCaption(doc, "Review", 0)
	if err == nil {
		t.Fatal("Expected error for ambiguous activity")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("Error = %q, want to contain 'ambiguous'", err.Error())
	}
}

func TestFindActivityByCaption_AtPosition(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Review", "task2")
	doc := makeWorkflowDoc(act1, act2)

	result, err := findActivityByCaption(doc, "Review", 2)
	if err != nil {
		t.Fatalf("findActivityByCaption @2 failed: %v", err)
	}
	if got := dGetString(result, "Name"); got != "task2" {
		t.Errorf("Name = %q, want %q", got, "task2")
	}
}

func TestFindActivityByCaption_AtPosition_OutOfRange(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	doc := makeWorkflowDoc(act1)

	_, err := findActivityByCaption(doc, "Review", 5)
	if err == nil {
		t.Fatal("Expected error for out-of-range position")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error = %q, want to contain 'not found'", err.Error())
	}
}

// --- DROP activity tests ---

func TestDropActivity_ByCaption(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Approve", "task2")
	act3 := makeWorkflowActivity("Workflows$UserTask", "Finalize", "task3")
	doc := makeWorkflowDoc(act1, act2, act3)

	op := &ast.DropActivityOp{ActivityRef: "Approve"}
	if err := applyDropActivity(doc, op); err != nil {
		t.Fatalf("DROP ACTIVITY failed: %v", err)
	}

	flow := dGetDoc(doc, "Flow")
	activities := dGetArrayElements(dGet(flow, "Activities"))
	if len(activities) != 2 {
		t.Fatalf("Expected 2 activities after drop, got %d", len(activities))
	}

	name0 := dGetString(activities[0].(bson.D), "Caption")
	name1 := dGetString(activities[1].(bson.D), "Caption")
	if name0 != "Review" {
		t.Errorf("First activity caption = %q, want %q", name0, "Review")
	}
	if name1 != "Finalize" {
		t.Errorf("Second activity caption = %q, want %q", name1, "Finalize")
	}
}

func TestDropActivity_NotFound(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	doc := makeWorkflowDoc(act1)

	op := &ast.DropActivityOp{ActivityRef: "NonExistent"}
	err := applyDropActivity(doc, op)
	if err == nil {
		t.Fatal("Expected error for dropping nonexistent activity")
	}
}

// --- DROP BOUNDARY EVENT tests ---

func TestDropBoundaryEvent_Single(t *testing.T) {
	evt := makeBoundaryEvent("Workflows$InterruptingTimerBoundaryEvent")
	act := makeWorkflowActivityWithBoundaryEvents("Review", evt)
	doc := makeWorkflowDoc(act)

	op := &ast.DropBoundaryEventOp{ActivityRef: "Review"}
	if err := applyDropBoundaryEvent(io.Discard, doc, op); err != nil {
		t.Fatalf("DROP BOUNDARY EVENT failed: %v", err)
	}

	actDoc, _ := findActivityByCaption(doc, "Review", 0)
	events := dGetArrayElements(dGet(actDoc, "BoundaryEvents"))
	if len(events) != 0 {
		t.Errorf("Expected 0 boundary events after drop, got %d", len(events))
	}
}

func TestDropBoundaryEvent_Multiple_DropsFirst(t *testing.T) {
	evt1 := makeBoundaryEvent("Workflows$InterruptingTimerBoundaryEvent")
	evt2 := makeBoundaryEvent("Workflows$NonInterruptingTimerBoundaryEvent")
	act := makeWorkflowActivityWithBoundaryEvents("Review", evt1, evt2)
	doc := makeWorkflowDoc(act)

	op := &ast.DropBoundaryEventOp{ActivityRef: "Review"}
	if err := applyDropBoundaryEvent(io.Discard, doc, op); err != nil {
		t.Fatalf("DROP BOUNDARY EVENT failed: %v", err)
	}

	actDoc, _ := findActivityByCaption(doc, "Review", 0)
	events := dGetArrayElements(dGet(actDoc, "BoundaryEvents"))
	if len(events) != 1 {
		t.Fatalf("Expected 1 boundary event after drop, got %d", len(events))
	}

	remaining := events[0].(bson.D)
	if got := dGetString(remaining, "$Type"); got != "Workflows$NonInterruptingTimerBoundaryEvent" {
		t.Errorf("Remaining event type = %q, want NonInterruptingTimerBoundaryEvent", got)
	}
}

func TestDropBoundaryEvent_NoEvents(t *testing.T) {
	act := makeWorkflowActivityWithBoundaryEvents("Review") // no events
	doc := makeWorkflowDoc(act)

	op := &ast.DropBoundaryEventOp{ActivityRef: "Review"}
	err := applyDropBoundaryEvent(io.Discard, doc, op)
	if err == nil {
		t.Fatal("Expected error when dropping from activity with no boundary events")
	}
	if !strings.Contains(err.Error(), "no boundary events") {
		t.Errorf("Error = %q, want to contain 'no boundary events'", err.Error())
	}
}

// --- findActivityIndex tests ---

func TestFindActivityIndex_Basic(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Approve", "task2")
	doc := makeWorkflowDoc(act1, act2)

	idx, activities, flow, err := findActivityIndex(doc, "Approve", 0)
	if err != nil {
		t.Fatalf("findActivityIndex failed: %v", err)
	}
	if idx != 1 {
		t.Errorf("index = %d, want 1", idx)
	}
	if len(activities) != 2 {
		t.Errorf("activities length = %d, want 2", len(activities))
	}
	if flow == nil {
		t.Error("flow should not be nil")
	}
}

func TestFindActivityIndex_NoFlow(t *testing.T) {
	doc := bson.D{
		{Key: "$Type", Value: "Workflows$Workflow"},
	}

	_, _, _, err := findActivityIndex(doc, "Review", 0)
	if err == nil {
		t.Fatal("Expected error for doc without Flow")
	}
	if !strings.Contains(err.Error(), "no Flow") {
		t.Errorf("Error = %q, want to contain 'no Flow'", err.Error())
	}
}

// --- collectAllActivityNames tests ---

func TestCollectAllActivityNames(t *testing.T) {
	act1 := makeWorkflowActivity("Workflows$UserTask", "Review", "ReviewTask")
	act2 := makeWorkflowActivity("Workflows$UserTask", "Approve", "ApproveTask")
	doc := makeWorkflowDoc(act1, act2)

	names := collectAllActivityNames(doc)
	if !names["ReviewTask"] {
		t.Error("Expected ReviewTask in names")
	}
	if !names["ApproveTask"] {
		t.Error("Expected ApproveTask in names")
	}
	if names["NonExistent"] {
		t.Error("NonExistent should not be in names")
	}
}

func TestCollectAllActivityNames_NoFlow(t *testing.T) {
	doc := bson.D{{Key: "$Type", Value: "Workflows$Workflow"}}

	names := collectAllActivityNames(doc)
	if len(names) != 0 {
		t.Errorf("Expected empty names map, got %d entries", len(names))
	}
}

// --- SET EXPORT_LEVEL / DUE_DATE ---

func TestSetWorkflowProperty_ExportLevel(t *testing.T) {
	doc := makeWorkflowDoc()
	// ExportLevel must exist in the doc for dSet to update it
	doc = append(doc, bson.E{Key: "ExportLevel", Value: "Usable"})

	op := &ast.SetWorkflowPropertyOp{Property: "EXPORT_LEVEL", Value: "Hidden"}
	if err := applySetWorkflowProperty(&doc, op); err != nil {
		t.Fatalf("SET EXPORT_LEVEL failed: %v", err)
	}

	if got := dGetString(doc, "ExportLevel"); got != "Hidden" {
		t.Errorf("ExportLevel = %q, want %q", got, "Hidden")
	}
}

// --- applySetActivityProperty tests ---

func TestSetActivityProperty_DueDate(t *testing.T) {
	act := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	act = append(act, bson.E{Key: "DueDate", Value: ""})
	doc := makeWorkflowDoc(act)

	op := &ast.SetActivityPropertyOp{
		ActivityRef: "Review",
		Property:    "DUE_DATE",
		Value:       "${PT48H}",
	}
	if err := applySetActivityProperty(doc, op); err != nil {
		t.Fatalf("SET DUE_DATE failed: %v", err)
	}

	actDoc, _ := findActivityByCaption(doc, "Review", 0)
	if got := dGetString(actDoc, "DueDate"); got != "${PT48H}" {
		t.Errorf("DueDate = %q, want %q", got, "${PT48H}")
	}
}

func TestSetActivityProperty_UnsupportedProperty(t *testing.T) {
	act := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	doc := makeWorkflowDoc(act)

	op := &ast.SetActivityPropertyOp{
		ActivityRef: "Review",
		Property:    "INVALID",
		Value:       "x",
	}
	err := applySetActivityProperty(doc, op)
	if err == nil {
		t.Fatal("Expected error for unsupported activity property")
	}
	if !strings.Contains(err.Error(), "unsupported activity property") {
		t.Errorf("Error = %q, want to contain 'unsupported activity property'", err.Error())
	}
}

// --- applyDropOutcome tests ---

func TestDropOutcome_NotFound(t *testing.T) {
	act := makeWorkflowActivity("Workflows$UserTask", "Review", "task1")
	// Add empty Outcomes array
	act = append(act, bson.E{Key: "Outcomes", Value: bson.A{int32(3)}})
	doc := makeWorkflowDoc(act)

	op := &ast.DropOutcomeOp{ActivityRef: "Review", OutcomeName: "NonExistent"}
	err := applyDropOutcome(doc, op)
	if err == nil {
		t.Fatal("Expected error for dropping nonexistent outcome")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error = %q, want to contain 'not found'", err.Error())
	}
}

// --- bsonArrayMarker constant test ---

func TestBsonArrayMarkerConstant(t *testing.T) {
	if bsonArrayMarker != int32(3) {
		t.Errorf("bsonArrayMarker = %v, want int32(3)", bsonArrayMarker)
	}
}
