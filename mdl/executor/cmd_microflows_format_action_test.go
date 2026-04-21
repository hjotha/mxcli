// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// =============================================================================
// formatAction — CRUD actions
// =============================================================================

func TestFormatAction_CreateObject_Simple(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateObjectAction{
		EntityQualifiedName: "MyModule.Customer",
		OutputVariable:      "NewCustomer",
	}
	got := e.formatAction(action, nil, nil)
	if got != "$NewCustomer = create MyModule.Customer;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CreateObject_WithMembers(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateObjectAction{
		EntityQualifiedName: "MyModule.Customer",
		OutputVariable:      "NewCustomer",
		InitialMembers: []*microflows.MemberChange{
			{AttributeQualifiedName: "MyModule.Customer.Name", Value: "'John'"},
			{AttributeQualifiedName: "MyModule.Customer.Age", Value: "25"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "$NewCustomer = create MyModule.Customer (Name = 'John', Age = 25);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_CreateObject_WithAssociationMember(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateObjectAction{
		EntityQualifiedName: "MyModule.Order",
		OutputVariable:      "NewOrder",
		InitialMembers: []*microflows.MemberChange{
			{AssociationQualifiedName: "MyModule.Order_Customer", Value: "$Customer"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "$NewOrder = create MyModule.Order (Order_Customer = $Customer);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_CreateObject_FallbackEntityID(t *testing.T) {
	e := newTestExecutor()
	entityNames := map[model.ID]string{mkID("e1"): "MyModule.Product"}
	action := &microflows.CreateObjectAction{
		EntityID:       mkID("e1"),
		OutputVariable: "NewProduct",
	}
	got := e.formatAction(action, entityNames, nil)
	if got != "$NewProduct = create MyModule.Product;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeObject(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeObjectAction{
		ChangeVariable: "Customer",
		Changes: []*microflows.MemberChange{
			{AttributeQualifiedName: "MyModule.Customer.Name", Value: "'Jane'"},
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "change $Customer (Name = 'Jane');" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeObject_NoChanges(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeObjectAction{ChangeVariable: "Obj"}
	got := e.formatAction(action, nil, nil)
	if got != "change $Obj;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_DeleteObject(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.DeleteObjectAction{DeleteVariable: "Customer"}
	got := e.formatAction(action, nil, nil)
	if got != "delete $Customer;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CommitObjects_WithEvents(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CommitObjectsAction{CommitVariable: "Order", WithEvents: true}
	got := e.formatAction(action, nil, nil)
	if got != "commit $Order with events;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CommitObjects_WithoutEvents(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CommitObjectsAction{CommitVariable: "Order"}
	got := e.formatAction(action, nil, nil)
	if got != "commit $Order;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CommitObjects_Refresh(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CommitObjectsAction{CommitVariable: "Order", RefreshInClient: true}
	got := e.formatAction(action, nil, nil)
	if got != "commit $Order refresh;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CommitObjects_WithEventsAndRefresh(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CommitObjectsAction{CommitVariable: "Order", WithEvents: true, RefreshInClient: true}
	got := e.formatAction(action, nil, nil)
	if got != "commit $Order with events refresh;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_RollbackObject(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RollbackObjectAction{RollbackVariable: "Order"}
	got := e.formatAction(action, nil, nil)
	if got != "rollback $Order;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_RollbackObject_Refresh(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RollbackObjectAction{RollbackVariable: "Order", RefreshInClient: true}
	got := e.formatAction(action, nil, nil)
	if got != "rollback $Order refresh;" {
		t.Errorf("got %q", got)
	}
}

// =============================================================================
// formatAction — Variable actions
// =============================================================================

func TestFormatAction_CreateVariable(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateVariableAction{
		VariableName: "Counter",
		DataType:     &microflows.IntegerType{},
		InitialValue: "0",
	}
	got := e.formatAction(action, nil, nil)
	if got != "declare $Counter Integer = 0;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_CreateVariable_NoInitial(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateVariableAction{
		VariableName: "Name",
		DataType:     &microflows.StringType{},
	}
	got := e.formatAction(action, nil, nil)
	if got != "declare $Name String = empty;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeVariable_Simple(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeVariableAction{
		VariableName: "Counter",
		Value:        "$Counter + 1",
	}
	got := e.formatAction(action, nil, nil)
	if got != "set $Counter = $Counter + 1;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeVariable_WithDollarPrefix(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeVariableAction{
		VariableName: "$Counter",
		Value:        "42",
	}
	got := e.formatAction(action, nil, nil)
	if got != "set $Counter = 42;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeVariable_XPath(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeVariableAction{
		VariableName: "$Product/Price",
		Value:        "9.99",
	}
	got := e.formatAction(action, nil, nil)
	if got != "change $Product (Price = 9.99);" {
		t.Errorf("got %q", got)
	}
}

// =============================================================================
// formatAction — List actions
// =============================================================================

func TestFormatAction_CreateList(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateListAction{
		EntityQualifiedName: "MyModule.Order",
		OutputVariable:      "OrderList",
	}
	got := e.formatAction(action, nil, nil)
	if got != "$OrderList = create list of MyModule.Order;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeList_Add(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeListAction{
		ChangeVariable: "OrderList",
		Type:           microflows.ChangeListTypeAdd,
		Value:          "$NewOrder",
	}
	got := e.formatAction(action, nil, nil)
	if got != "add $NewOrder to $OrderList;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeList_Remove(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeListAction{
		ChangeVariable: "OrderList",
		Type:           microflows.ChangeListTypeRemove,
		Value:          "$OldOrder",
	}
	got := e.formatAction(action, nil, nil)
	if got != "remove $OldOrder from $OrderList;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeList_Clear(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeListAction{
		ChangeVariable: "OrderList",
		Type:           microflows.ChangeListTypeClear,
	}
	got := e.formatAction(action, nil, nil)
	if got != "clear $OrderList;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ChangeList_Set(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeListAction{
		ChangeVariable: "OrderList",
		Type:           microflows.ChangeListTypeSet,
		Value:          "$OtherList",
	}
	got := e.formatAction(action, nil, nil)
	if got != "set $OrderList = $OtherList;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_AggregateList_Count(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.AggregateListAction{
		InputVariable:  "Orders",
		OutputVariable: "Total",
		Function:       microflows.AggregateFunctionCount,
	}
	got := e.formatAction(action, nil, nil)
	if got != "$Total = count($Orders);" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_AggregateList_Sum(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.AggregateListAction{
		InputVariable:          "Orders",
		OutputVariable:         "TotalAmount",
		Function:               microflows.AggregateFunctionSum,
		AttributeQualifiedName: "MyModule.Order.Amount",
	}
	got := e.formatAction(action, nil, nil)
	if got != "$TotalAmount = sum($Orders.Amount);" {
		t.Errorf("got %q", got)
	}
}

// =============================================================================
// formatAction — Call actions
// =============================================================================

func TestFormatAction_MicroflowCall_WithResult(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.MicroflowCallAction{
		ResultVariableName: "Result",
		MicroflowCall: &microflows.MicroflowCall{
			Microflow: "MyModule.ProcessOrder",
			ParameterMappings: []*microflows.MicroflowCallParameterMapping{
				{Parameter: "MyModule.ProcessOrder.Order", Argument: "$Order"},
			},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "$Result = call microflow MyModule.ProcessOrder(Order = $Order);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_MicroflowCall_NoResult(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.MicroflowCallAction{
		MicroflowCall: &microflows.MicroflowCall{
			Microflow: "MyModule.DoSomething",
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "call microflow MyModule.DoSomething();" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_JavaActionCall(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.JavaActionCallAction{
		JavaAction:         "MyModule.SendEmail",
		ResultVariableName: "Success",
		ParameterMappings: []*microflows.JavaActionParameterMapping{
			{
				Parameter: "MyModule.SendEmail.To",
				Value: &microflows.ExpressionBasedCodeActionParameterValue{
					Expression: "$Customer/Email",
				},
			},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "$Success = call java action MyModule.SendEmail(To = $Customer/Email);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_CallExternal(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CallExternalAction{
		ConsumedODataService: "MyModule.OrderService",
		Name:                 "GetOrders",
		ResultVariableName:   "Orders",
	}
	got := e.formatAction(action, nil, nil)
	want := "$Orders = call external action MyModule.OrderService.GetOrders();"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// =============================================================================
// formatAction — UI actions
// =============================================================================

func TestFormatAction_ShowPage(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ShowPageAction{
		PageName: "MyModule.CustomerEdit",
	}
	got := e.formatAction(action, nil, nil)
	if got != "show page MyModule.CustomerEdit;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ShowPage_WithParams(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ShowPageAction{
		PageName: "MyModule.OrderDetail",
		PageParameterMappings: []*microflows.PageParameterMapping{
			{Parameter: "MyModule.OrderDetail.Order", Argument: "$Order"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "show page MyModule.OrderDetail($Order = $Order);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_ClosePage(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ClosePageAction{NumberOfPages: 1}
	got := e.formatAction(action, nil, nil)
	if got != "close page;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ClosePage_Multiple(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ClosePageAction{NumberOfPages: 3}
	got := e.formatAction(action, nil, nil)
	if got != "close page 3;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ShowMessage(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ShowMessageAction{
		Type: microflows.MessageTypeWarning,
		Template: &model.Text{
			Translations: map[string]string{"en_US": "Order saved"},
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "show message 'Order saved' type Warning;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ShowMessage_EscapesQuotes(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ShowMessageAction{
		Type: microflows.MessageTypeInformation,
		Template: &model.Text{
			Translations: map[string]string{"en_US": "It's done"},
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "show message 'It''s done' type Information;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ShowMessage_EscapesMultiline(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ShowMessageAction{
		Type: microflows.MessageTypeInformation,
		Template: &model.Text{
			Translations: map[string]string{"en_US": "Line 1\nLine 2\tTabbed"},
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "show message 'Line 1\\nLine 2\\tTabbed' type Information;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_ValidationFeedback(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ValidationFeedbackAction{
		ObjectVariable: "Customer",
		AttributeName:  "MyModule.Customer.Email",
		Template: &model.Text{
			Translations: map[string]string{"en_US": "Email is required"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "validation feedback $Customer/Email message 'Email is required';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_LogMessage(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.LogMessageAction{
		LogLevel:    microflows.LogLevelWarning,
		LogNodeName: "'OrderService'",
		MessageTemplate: &model.Text{
			Translations: map[string]string{"en_US": "Processing order"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "log warning node 'OrderService' 'Processing order';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_LogMessage_WithTemplateParams(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.LogMessageAction{
		LogLevel:    microflows.LogLevelInfo,
		LogNodeName: "'App'",
		MessageTemplate: &model.Text{
			Translations: map[string]string{"en_US": "Order {1} for {2}"},
		},
		TemplateParameters: []string{"$OrderNumber", "$CustomerName"},
	}
	got := e.formatAction(action, nil, nil)
	want := "log info node 'App' 'Order {1} for {2}' with ({1} = $OrderNumber, {2} = $CustomerName);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_LogMessage_EscapesMultiline(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.LogMessageAction{
		LogLevel:    microflows.LogLevelInfo,
		LogNodeName: "'App'",
		MessageTemplate: &model.Text{
			Translations: map[string]string{"en_US": "Line 1\nLine 2"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "LOG INFO NODE 'App' 'Line 1\\nLine 2';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_UnknownAction(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.UnknownAction{TypeName: "SomeNewAction"}
	got := e.formatAction(action, nil, nil)
	if got != "-- Unsupported action type: SomeNewAction" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_Nil(t *testing.T) {
	e := newTestExecutor()
	got := e.formatAction(nil, nil, nil)
	if got != "-- Empty action" {
		t.Errorf("got %q", got)
	}
}

// =============================================================================
// formatAction — Retrieve actions
// =============================================================================

func TestFormatAction_Retrieve_Database(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "Customers",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "MyModule.Customer",
		},
	}
	got := e.formatAction(action, nil, nil)
	if got != "retrieve $Customers from MyModule.Customer;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAction_Retrieve_WithXPath(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "ActiveCustomers",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "MyModule.Customer",
			XPathConstraint:     "[IsActive = true()]",
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "retrieve $ActiveCustomers from MyModule.Customer\n    where IsActive = true();"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_Retrieve_WithLimit(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "First",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "MyModule.Customer",
			Range:               &microflows.Range{RangeType: microflows.RangeTypeFirst},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "retrieve $First from MyModule.Customer\n    limit 1;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_Retrieve_WithSorting(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "Sorted",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "MyModule.Customer",
			Sorting: []*microflows.SortItem{
				{AttributeQualifiedName: "MyModule.Customer.Name", Direction: microflows.SortDirectionAscending},
				{AttributeQualifiedName: "MyModule.Customer.Age", Direction: microflows.SortDirectionDescending},
			},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "retrieve $Sorted from MyModule.Customer\n    sort by MyModule.Customer.Name asc, MyModule.Customer.Age desc;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_Retrieve_Association(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "Address",
		Source: &microflows.AssociationRetrieveSource{
			StartVariable:            "Customer",
			AssociationQualifiedName: "MyModule.Customer_Address",
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "retrieve $Address from $Customer/MyModule.Customer_Address;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
