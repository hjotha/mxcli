// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestParseAsyncAPI_Basic(t *testing.T) {
	yaml := `asyncapi: "2.2.0"
info:
  title: Order Service
  version: "1.0.0"
  description: Handles orders
channels:
  order/created:
    subscribe:
      operationId: receiveOrderCreated
      message:
        $ref: "#/components/messages/OrderCreated"
components:
  messages:
    OrderCreated:
      title: Order Created
      description: An order was created
      contentType: application/json
      payload:
        $ref: "#/components/schemas/OrderPayload"
  schemas:
    OrderPayload:
      type: object
      properties:
        orderId:
          type: string
        amount:
          type: number
          format: double
`

	doc, err := ParseAsyncAPI(yaml)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Version != "2.2.0" {
		t.Errorf("expected version 2.2.0, got %q", doc.Version)
	}
	if doc.Title != "Order Service" {
		t.Errorf("expected title Order Service, got %q", doc.Title)
	}
	if doc.DocVersion != "1.0.0" {
		t.Errorf("expected doc version 1.0.0, got %q", doc.DocVersion)
	}
	if doc.Description != "Handles orders" {
		t.Errorf("expected description, got %q", doc.Description)
	}

	// Messages
	if len(doc.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(doc.Messages))
	}
	msg := doc.Messages[0]
	if msg.Name != "OrderCreated" {
		t.Errorf("expected OrderCreated, got %q", msg.Name)
	}
	if msg.Title != "Order Created" {
		t.Errorf("expected title, got %q", msg.Title)
	}
	if len(msg.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(msg.Properties))
	}
	// Properties should be sorted alphabetically
	if msg.Properties[0].Name != "amount" {
		t.Errorf("expected first property 'amount' (sorted), got %q", msg.Properties[0].Name)
	}
	if msg.Properties[1].Name != "orderId" {
		t.Errorf("expected second property 'orderId' (sorted), got %q", msg.Properties[1].Name)
	}

	// Channels
	if len(doc.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(doc.Channels))
	}
	ch := doc.Channels[0]
	if ch.Name != "order/created" {
		t.Errorf("expected channel name, got %q", ch.Name)
	}
	if ch.OperationType != "subscribe" {
		t.Errorf("expected subscribe, got %q", ch.OperationType)
	}
	if ch.MessageRef != "OrderCreated" {
		t.Errorf("expected message ref OrderCreated, got %q", ch.MessageRef)
	}
}

func TestParseAsyncAPI_Empty(t *testing.T) {
	_, err := ParseAsyncAPI("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseAsyncAPI_InvalidYAML(t *testing.T) {
	_, err := ParseAsyncAPI("not: valid: yaml: [")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseAsyncAPI_MultipleChannels_Sorted(t *testing.T) {
	yaml := `asyncapi: "2.0.0"
info:
  title: Test
  version: "1.0"
channels:
  z/channel:
    publish:
      operationId: pub
      message:
        $ref: "#/components/messages/Msg"
  a/channel:
    subscribe:
      operationId: sub
      message:
        $ref: "#/components/messages/Msg"
components:
  messages:
    Msg:
      title: Test Message
`

	doc, err := ParseAsyncAPI(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(doc.Channels))
	}
	// Channels should be sorted by name
	if doc.Channels[0].Name != "a/channel" {
		t.Errorf("expected first channel a/channel (sorted), got %q", doc.Channels[0].Name)
	}
	if doc.Channels[1].Name != "z/channel" {
		t.Errorf("expected second channel z/channel (sorted), got %q", doc.Channels[1].Name)
	}
}

func TestFindMessage(t *testing.T) {
	doc := &AsyncAPIDocument{
		Messages: []*AsyncAPIMessage{
			{Name: "OrderCreated"},
			{Name: "OrderUpdated"},
		},
	}

	if got := doc.FindMessage("OrderCreated"); got == nil || got.Name != "OrderCreated" {
		t.Error("expected to find OrderCreated")
	}
	// Case-insensitive
	if got := doc.FindMessage("ordercreated"); got == nil {
		t.Error("expected case-insensitive match")
	}
	if got := doc.FindMessage("Missing"); got != nil {
		t.Error("expected nil for missing message")
	}
}

func TestParseAsyncAPI_InlinePayload(t *testing.T) {
	yaml := `asyncapi: "2.0.0"
info:
  title: Test
  version: "1.0"
channels: {}
components:
  messages:
    Inline:
      title: Inline Message
      payload:
        type: object
        properties:
          field1:
            type: string
          field2:
            type: integer
            format: int32
`

	doc, err := ParseAsyncAPI(yaml)
	if err != nil {
		t.Fatal(err)
	}
	msg := doc.Messages[0]
	if len(msg.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(msg.Properties))
	}
	// Sorted
	if msg.Properties[0].Name != "field1" {
		t.Errorf("expected field1 first, got %q", msg.Properties[0].Name)
	}
	if msg.Properties[1].Format != "int32" {
		t.Errorf("expected format int32, got %q", msg.Properties[1].Format)
	}
}
