package piagent

import (
	"encoding/json"
	"testing"
)

func TestRPCEventParsesExtensionUIRequest(t *testing.T) {
	raw := []byte(`{
		"type":"extension_ui_request",
		"id":"uuid-1",
		"method":"select",
		"title":"Allow?",
		"options":["Allow","Block"],
		"timeout":10000
	}`)

	var event RPCEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatal(err)
	}
	if event.ID != "uuid-1" || event.Method != "select" || event.Title != "Allow?" {
		t.Fatalf("event = %#v", event)
	}
	if len(event.Options) != 2 || event.Options[1] != "Block" || event.Timeout != 10000 {
		t.Fatalf("event = %#v", event)
	}
}

func TestRPCCommandMarshalsExtensionUIResponse(t *testing.T) {
	confirmed := false
	command := RPCCommand{
		Type:      "extension_ui_response",
		ID:        "uuid-2",
		Confirmed: &confirmed,
		Cancelled: false,
	}

	data, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["type"] != "extension_ui_response" || raw["id"] != "uuid-2" {
		t.Fatalf("response = %s", data)
	}
	if value, ok := raw["confirmed"].(bool); !ok || value {
		t.Fatalf("response confirmed = %#v, want false", raw["confirmed"])
	}
}

func TestRPCEventParsesExtensionNotification(t *testing.T) {
	raw := []byte(`{
		"type":"extension_ui_request",
		"id":"uuid-3",
		"method":"notify",
		"message":"Command blocked",
		"notifyType":"warning"
	}`)

	var event RPCEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatal(err)
	}
	if event.Method != "notify" || string(event.Message) != `"Command blocked"` || event.NotifyType != "warning" {
		t.Fatalf("event = %#v", event)
	}
}
