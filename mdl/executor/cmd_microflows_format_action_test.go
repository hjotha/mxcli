// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	mdltypes "github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"go.mongodb.org/mongo-driver/bson"
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

func TestFormatAction_CreateObject_MultilineMemberKeepsCommaOnNextLine(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CreateObjectAction{
		EntityQualifiedName: "MyModule.Token",
		OutputVariable:      "Token",
		InitialMembers: []*microflows.MemberChange{
			{AttributeQualifiedName: "MyModule.Token.Value", Value: "if $Count = 0\nthen $Fallback\nelse $Actual"},
			{AttributeQualifiedName: "MyModule.Token.TimeStamp", Value: "$Now"},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "$Token = create MyModule.Token (Value = if $Count = 0\nthen $Fallback\nelse $Actual\n, TimeStamp = $Now);"
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

func TestFormatAction_ChangeObject_EscapesRawControlCharsInExpressionLiterals(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeObjectAction{
		ChangeVariable: "ValidationFeedback",
		Changes: []*microflows.MemberChange{
			{
				AttributeQualifiedName: "MyModule.ValidationFeedback.FeedbackContent",
				Value:                  "$ValidationFeedback/FeedbackContent + '\n' + $ErrorMessage",
			},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "change $ValidationFeedback (FeedbackContent = $ValidationFeedback/FeedbackContent + '\\n' + $ErrorMessage);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_ChangeObject_MultilineMemberKeepsCommaOnNextLine(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ChangeObjectAction{
		ChangeVariable: "Response",
		Changes: []*microflows.MemberChange{
			{AttributeQualifiedName: "MyModule.Response.StatusCode", Value: "if ($Latest = empty)\nthen 500\nelse $Latest/StatusCode"},
			{AttributeQualifiedName: "MyModule.Response.ReasonPhrase", Value: "if ($Latest = empty)\nthen $latestError/Message\nelse $Latest/ReasonPhrase"},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "change $Response (StatusCode = if ($Latest = empty)\nthen 500\nelse $Latest/StatusCode\n, ReasonPhrase = if ($Latest = empty)\nthen $latestError/Message\nelse $Latest/ReasonPhrase);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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

func TestFormatAction_AggregateList_SumExpression(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.AggregateListAction{
		InputVariable:  "Orders",
		OutputVariable: "TotalTax",
		Function:       microflows.AggregateFunctionSum,
		UseExpression:  true,
		Expression:     "$currentObject/Amount * 0.21",
	}
	got := e.formatAction(action, nil, nil)
	if got != "$TotalTax = sum($Orders, $currentObject/Amount * 0.21);" {
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

func TestFormatAction_MicroflowCall_TrimsTrailingArgumentWhitespace(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.MicroflowCallAction{
		MicroflowCall: &microflows.MicroflowCall{
			Microflow: "MyModule.Handle",
			ParameterMappings: []*microflows.MicroflowCallParameterMapping{
				{Parameter: "MyModule.Handle.Path", Argument: "$Path\n"},
				{Parameter: "MyModule.Handle.Timestamp", Argument: "[%CurrentDateTime%]"},
			},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "call microflow MyModule.Handle(Path = $Path, Timestamp = [%CurrentDateTime%]);"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_MicroflowCall_CollapsesMultilineArgumentWhitespace(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.MicroflowCallAction{
		ResultVariableName: "RuntimesRoot",
		MicroflowCall: &microflows.MicroflowCall{
			Microflow: "SampleRuntimeApi.REST_RetrieveRuntimes",
			ParameterMappings: []*microflows.MicroflowCallParameterMapping{
				{
					Parameter: "SampleRuntimeApi.REST_RetrieveRuntimes.Offset",
					Argument:  "if $ListContext/TriggeredBySearch then '0'\nelse\ntoString($RuntimeListContext/PageSize * ($ListHeader/CurrentPageNumber-1))",
				},
			},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "$RuntimesRoot = call microflow SampleRuntimeApi.REST_RetrieveRuntimes(Offset = if $ListContext/TriggeredBySearch then '0' else toString($RuntimeListContext/PageSize * ($ListHeader/CurrentPageNumber - 1)));"
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

func TestFormatAction_DownloadFile(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.DownloadFileAction{
		FileDocument:      "GeneratedExcelDoc",
		ShowInBrowser:     true,
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
	}
	got := e.formatAction(action, nil, nil)
	want := "download file $GeneratedExcelDoc show in browser;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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

func TestFormatAction_ValidationFeedback_ObjectOnlyTarget(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ValidationFeedbackAction{
		ObjectVariable: "AdminCandidate",
		Template: &model.Text{
			Translations: map[string]string{"en_US": "Please select the requested user."},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "validation feedback $AdminCandidate message 'Please select the requested user.';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_ValidationFeedback_AssociationTarget(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.ValidationFeedbackAction{
		ObjectVariable:  "AdminCandidate",
		AssociationName: "SampleRequests.AdminCandidate_User",
		AttributeName:   "",
		Template: &model.Text{
			Translations: map[string]string{"en_US": "Please select the requested user."},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "validation feedback $AdminCandidate/SampleRequests.AdminCandidate_User message 'Please select the requested user.';"
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
	want := "log info node 'App' 'Line 1\\nLine 2';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_LogMessage_NodeExpression(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.LogMessageAction{
		LogLevel:    microflows.LogLevelInfo,
		LogNodeName: "@MyModule.SecurityLogNode",
		MessageTemplate: &model.Text{
			Translations: map[string]string{"en_US": "User added"},
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "log info node @MyModule.SecurityLogNode 'User added';"
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

func TestFormatAction_WebServiceCall(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.WebServiceCallAction{
		ServiceID:         "service-1",
		OperationName:     "FetchSampleItems",
		SendMappingID:     "send-1",
		ReceiveMappingID:  "receive-1",
		OutputVariable:    "Root",
		UseReturnVariable: true,
		TimeoutExpression: "30",
	}
	got := e.formatAction(action, nil, nil)
	want := "$Root = call web service 'service-1'\noperation 'FetchSampleItems'\nsend mapping 'send-1'\nreceive mapping 'receive-1'\ntimeout 30;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_WebServiceCallResolvesKnownReferences(t *testing.T) {
	moduleID := mkID("soap-module")
	serviceID := mkID("soap-service")
	sendMappingID := mkID("soap-send")
	receiveMappingID := mkID("soap-receive")
	serviceContents, err := bson.Marshal(bson.M{"Name": "OrderService"})
	if err != nil {
		t.Fatal(err)
	}

	backend := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListRawUnitsByTypeFunc: func(typePrefix string) ([]*mdltypes.RawUnit, error) {
			if typePrefix != "WebServices$ImportedWebService" {
				t.Fatalf("unexpected type prefix %q", typePrefix)
			}
			return []*mdltypes.RawUnit{{
				ID:          serviceID,
				ContainerID: moduleID,
				Type:        "WebServices$ImportedWebService",
				Contents:    serviceContents,
			}}, nil
		},
		ListExportMappingsFunc: func() ([]*model.ExportMapping, error) {
			return []*model.ExportMapping{{
				BaseElement: model.BaseElement{ID: sendMappingID},
				ContainerID: moduleID,
				Name:        "OrderRequest",
			}}, nil
		},
		ListImportMappingsFunc: func() ([]*model.ImportMapping, error) {
			return []*model.ImportMapping{{
				BaseElement: model.BaseElement{ID: receiveMappingID},
				ContainerID: moduleID,
				Name:        "OrderResponse",
			}}, nil
		},
	}
	h := mkHierarchy(&model.Module{BaseElement: model.BaseElement{ID: moduleID}, Name: "SampleSOAP"})
	ctx, _ := newMockCtx(t, withBackend(backend), withHierarchy(h))

	action := &microflows.WebServiceCallAction{
		ServiceID:         serviceID,
		OperationName:     "FetchOrders",
		SendMappingID:     sendMappingID,
		ReceiveMappingID:  receiveMappingID,
		OutputVariable:    "Root",
		UseReturnVariable: true,
	}
	got := formatAction(ctx, action, nil, nil)
	want := "$Root = call web service 'SampleSOAP.OrderService'\noperation 'FetchOrders'\nsend mapping 'SampleSOAP.OrderRequest'\nreceive mapping 'SampleSOAP.OrderResponse';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_WebServiceCallKeepsRawReferencesWhenUnknown(t *testing.T) {
	backend := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListRawUnitsByTypeFunc: func(typePrefix string) ([]*mdltypes.RawUnit, error) {
			return nil, nil
		},
		ListExportMappingsFunc: func() ([]*model.ExportMapping, error) {
			return nil, nil
		},
		ListImportMappingsFunc: func() ([]*model.ImportMapping, error) {
			return nil, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(backend), withHierarchy(mkHierarchy()))

	action := &microflows.WebServiceCallAction{
		ServiceID:        "dangling-service-id",
		OperationName:    "FetchOrders",
		SendMappingID:    "dangling-send-id",
		ReceiveMappingID: "dangling-receive-id",
		OutputVariable:   "Root",
	}
	got := formatAction(ctx, action, nil, nil)
	want := "$Root = call web service 'dangling-service-id'\noperation 'FetchOrders'\nsend mapping 'dangling-send-id'\nreceive mapping 'dangling-receive-id';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_WebServiceCallRaw(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.WebServiceCallAction{
		OutputVariable: "Root",
		RawBSON:        []byte{1, 2, 3},
	}
	got := e.formatAction(action, nil, nil)
	want := "$Root = call web service raw 'AQID';"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_WebServiceCallRawCanonicalizesEquivalentBSON(t *testing.T) {
	e := newTestExecutor()
	rawA, err := bson.Marshal(bson.D{
		{Key: "Z", Value: int32(1)},
		{Key: "Nested", Value: bson.D{
			{Key: "B", Value: "two"},
			{Key: "A", Value: "one"},
		}},
		{Key: "$Type", Value: "Microflows$CallWebServiceAction"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rawB, err := bson.Marshal(bson.D{
		{Key: "$Type", Value: "Microflows$CallWebServiceAction"},
		{Key: "Nested", Value: bson.D{
			{Key: "A", Value: "one"},
			{Key: "B", Value: "two"},
		}},
		{Key: "Z", Value: int32(1)},
	})
	if err != nil {
		t.Fatal(err)
	}

	gotA := e.formatAction(&microflows.WebServiceCallAction{OutputVariable: "Root", RawBSON: rawA}, nil, nil)
	gotB := e.formatAction(&microflows.WebServiceCallAction{OutputVariable: "Root", RawBSON: rawB}, nil, nil)
	if gotA != gotB {
		t.Fatalf("canonical raw statements differ:\nA=%s\nB=%s", gotA, gotB)
	}

	rawText := strings.TrimSuffix(strings.TrimPrefix(gotA, "$Root = call web service raw '"), "';")
	decoded, err := base64.StdEncoding.DecodeString(rawText)
	if err != nil {
		t.Fatal(err)
	}
	var canonical bson.D
	if err := bson.Unmarshal(decoded, &canonical); err != nil {
		t.Fatal(err)
	}
	if canonical[0].Key != "$Type" || canonical[1].Key != "Nested" || canonical[2].Key != "Z" {
		t.Fatalf("unexpected canonical key order: %#v", canonical)
	}
	nested := canonical[1].Value.(bson.D)
	if nested[0].Key != "A" || nested[1].Key != "B" {
		t.Fatalf("unexpected nested key order: %#v", nested)
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

func TestFormatAction_Retrieve_WithMultipleXPathPredicates(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.RetrieveAction{
		OutputVariable: "ActiveCustomers",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "MyModule.Customer",
			XPathConstraint:     "[IsActive=true()][LastLogin > '[%CurrentDateTime%]'][Owner = $CurrentUser]",
		},
	}
	got := e.formatAction(action, nil, nil)
	want := "retrieve $ActiveCustomers from MyModule.Customer\n    where IsActive=true() and LastLogin > '[%CurrentDateTime%]' and Owner = $CurrentUser;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatXPathConstraintForWhere_KeepsBracketMacrosInsideStrings(t *testing.T) {
	got := formatXPathConstraintForWhere("[CalculatedAt < '[%CurrentDateTime%]'][ArtifactId!=empty]")
	want := "CalculatedAt < '[%CurrentDateTime%]' and ArtifactId!=empty"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeExpressionSource_EscapesOnlyControlCharsInsideStringLiterals(t *testing.T) {
	got := normalizeExpressionSource("if $Flag then 'Line 1\nLine 2'\nelse $Other")
	want := "if $Flag then 'Line 1\\nLine 2'\nelse $Other"
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

func TestFormatAction_Retrieve_ReverseAssociationDatabaseSourceUsesCompactForm(t *testing.T) {
	e := newTestExecutor()
	e.backend = reverseAssociationBackend(t)
	action := &microflows.RetrieveAction{
		OutputVariable: "Domains",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "SampleRuntime.Domain",
			XPathConstraint:     "[SampleRuntime.Domain_Runtime = $Runtime]",
			Range:               &microflows.Range{RangeType: microflows.RangeTypeAll},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "retrieve $Domains from $Runtime/SampleRuntime.Domain_Runtime;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_Retrieve_ReverseAssociationRequiresSimpleAllRange(t *testing.T) {
	e := newTestExecutor()
	e.backend = reverseAssociationBackend(t)
	action := &microflows.RetrieveAction{
		OutputVariable: "Domains",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "SampleRuntime.Domain",
			XPathConstraint:     "[SampleRuntime.Domain_Runtime = $Runtime]",
			Range:               &microflows.Range{RangeType: microflows.RangeTypeFirst},
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "retrieve $Domains from SampleRuntime.Domain\n    where SampleRuntime.Domain_Runtime = $Runtime\n    limit 1;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAction_Retrieve_ReverseAssociationRequiresMatchingEntity(t *testing.T) {
	e := newTestExecutor()
	e.backend = reverseAssociationBackend(t)
	action := &microflows.RetrieveAction{
		OutputVariable: "Domains",
		Source: &microflows.DatabaseRetrieveSource{
			EntityQualifiedName: "SampleRuntime.Runtime",
			XPathConstraint:     "[SampleRuntime.Domain_Runtime = $Runtime]",
		},
	}

	got := e.formatAction(action, nil, nil)
	want := "retrieve $Domains from SampleRuntime.Runtime\n    where SampleRuntime.Domain_Runtime = $Runtime;"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseReverseAssociationXPathRejectsComplexPredicates(t *testing.T) {
	tests := []string{
		"[SampleRuntime.Domain_Runtime = $Runtime][Active = true]",
		"[SampleRuntime.Domain_Runtime != $Runtime]",
		"[SampleRuntime.Domain_Runtime = $Runtime/Other.Assoc]",
		"[SampleRuntime.Domain_Runtime = 'literal']",
		"SampleRuntime.Domain_Runtime = $Runtime",
	}

	for _, tt := range tests {
		if assoc, start, ok := parseReverseAssociationXPath(tt); ok {
			t.Fatalf("parseReverseAssociationXPath(%q) = %q, %q, true; want false", tt, assoc, start)
		}
	}
}

func reverseAssociationBackend(t *testing.T) *mock.MockBackend {
	t.Helper()
	moduleID := model.ID("clouddata-module")
	return &mock.MockBackend{
		GetModuleByNameFunc: func(name string) (*model.Module, error) {
			if name != "SampleRuntime" {
				return nil, nil
			}
			return &model.Module{
				BaseElement: model.BaseElement{ID: moduleID},
				Name:        "SampleRuntime",
			}, nil
		},
		GetDomainModelFunc: func(id model.ID) (*domainmodel.DomainModel, error) {
			if id != moduleID {
				return nil, nil
			}
			return &domainmodel.DomainModel{
				ContainerID: moduleID,
				Entities: []*domainmodel.Entity{
					{
						BaseElement: model.BaseElement{ID: "domain-entity"},
						Name:        "Domain",
					},
					{
						BaseElement: model.BaseElement{ID: "environment-entity"},
						Name:        "Runtime",
					},
				},
				Associations: []*domainmodel.Association{
					{
						Name:     "Domain_Runtime",
						ParentID: "domain-entity",
						ChildID:  "environment-entity",
						Type:     domainmodel.AssociationTypeReference,
					},
				},
			}, nil
		},
	}
}
