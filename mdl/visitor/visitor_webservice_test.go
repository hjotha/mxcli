// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestCallWebServiceStatement(t *testing.T) {
	input := `create microflow Module.Test ()
begin
  $Root = call web service 'service-1'
  operation 'GetAccessGroups'
  send mapping 'send-1'
  receive mapping 'receive-1'
  timeout 30 on error rollback;
end;`

	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("parse error: %v", err)
		}
		return
	}
	mf, ok := prog.Statements[0].(*ast.CreateMicroflowStmt)
	if !ok {
		t.Fatalf("statement = %T, want *ast.CreateMicroflowStmt", prog.Statements[0])
	}
	call, ok := mf.Body[0].(*ast.CallWebServiceStmt)
	if !ok {
		t.Fatalf("body[0] = %T, want *ast.CallWebServiceStmt", mf.Body[0])
	}
	if call.OutputVariable != "Root" || call.ServiceID != "service-1" ||
		call.OperationName != "GetAccessGroups" || call.SendMappingID != "send-1" ||
		call.ReceiveMappingID != "receive-1" {
		t.Fatalf("unexpected web service call: %#v", call)
	}
	if call.ErrorHandling == nil || call.ErrorHandling.Type != ast.ErrorHandlingRollback {
		t.Fatalf("error handling = %#v, want rollback", call.ErrorHandling)
	}
}

func TestCallWebServiceRawStatement(t *testing.T) {
	input := `create microflow Module.Test ()
begin
  $Root = call web service raw 'AQID' on error rollback;
end;`

	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("parse error: %v", err)
		}
		return
	}
	mf := prog.Statements[0].(*ast.CreateMicroflowStmt)
	call, ok := mf.Body[0].(*ast.CallWebServiceStmt)
	if !ok {
		t.Fatalf("body[0] = %T, want *ast.CallWebServiceStmt", mf.Body[0])
	}
	if call.OutputVariable != "Root" || call.RawBSONBase64 != "AQID" {
		t.Fatalf("unexpected raw web service call: %#v", call)
	}
	if call.ErrorHandling == nil || call.ErrorHandling.Type != ast.ErrorHandlingRollback {
		t.Fatalf("error handling = %#v, want rollback", call.ErrorHandling)
	}
}
