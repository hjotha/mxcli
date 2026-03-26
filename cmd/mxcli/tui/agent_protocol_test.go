package tui

import (
	"encoding/json"
	"testing"
)

func TestAgentRequestParsing(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AgentRequest
		wantErr bool
	}{
		{
			name:  "exec action",
			input: `{"id":1,"action":"exec","mdl":"SHOW ENTITIES"}`,
			want:  AgentRequest{ID: 1, Action: "exec", MDL: "SHOW ENTITIES"},
		},
		{
			name:  "check action",
			input: `{"id":2,"action":"check"}`,
			want:  AgentRequest{ID: 2, Action: "check"},
		},
		{
			name:  "state action",
			input: `{"id":3,"action":"state"}`,
			want:  AgentRequest{ID: 3, Action: "state"},
		},
		{
			name:  "navigate action",
			input: `{"id":4,"action":"navigate","target":"MyModule.Customer"}`,
			want:  AgentRequest{ID: 4, Action: "navigate", Target: "MyModule.Customer"},
		},
		{
			name:  "delete action",
			input: `{"id":5,"action":"delete","target":"entity:MyModule.Customer"}`,
			want:  AgentRequest{ID: 5, Action: "delete", Target: "entity:MyModule.Customer"},
		},
		{
			name:  "create_module action",
			input: `{"id":6,"action":"create_module","name":"NewModule"}`,
			want:  AgentRequest{ID: 6, Action: "create_module", Name: "NewModule"},
		},
		{
			name:  "format action",
			input: `{"id":7,"action":"format","mdl":"SHOW ENTITIES"}`,
			want:  AgentRequest{ID: 7, Action: "format", MDL: "SHOW ENTITIES"},
		},
		{
			name:  "describe action",
			input: `{"id":8,"action":"describe","target":"entity:MyModule.Customer"}`,
			want:  AgentRequest{ID: 8, Action: "describe", Target: "entity:MyModule.Customer"},
		},
		{
			name:  "list action",
			input: `{"id":9,"action":"list","target":"entities"}`,
			want:  AgentRequest{ID: 9, Action: "list", Target: "entities"},
		},
		{
			name:    "invalid json",
			input:   `{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AgentRequest
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAgentResponseSerialization(t *testing.T) {
	tests := []struct {
		name string
		resp AgentResponse
		want string
	}{
		{
			name: "success response",
			resp: AgentResponse{ID: 1, OK: true, Result: "3 entities found"},
			want: `{"id":1,"ok":true,"result":"3 entities found"}`,
		},
		{
			name: "error response",
			resp: AgentResponse{ID: 2, OK: false, Error: "syntax error at line 1"},
			want: `{"id":2,"ok":false,"error":"syntax error at line 1"}`,
		},
		{
			name: "response with mode",
			resp: AgentResponse{ID: 3, OK: true, Mode: "browser"},
			want: `{"id":3,"ok":true,"mode":"browser"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.resp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAgentRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     AgentRequest
		wantErr string
	}{
		{
			name:    "missing id",
			req:     AgentRequest{Action: "check"},
			wantErr: "missing request id",
		},
		{
			name: "valid check",
			req:  AgentRequest{ID: 1, Action: "check"},
		},
		{
			name: "valid state",
			req:  AgentRequest{ID: 2, Action: "state"},
		},
		{
			name:    "exec without mdl",
			req:     AgentRequest{ID: 3, Action: "exec"},
			wantErr: "exec action requires mdl field",
		},
		{
			name: "valid exec",
			req:  AgentRequest{ID: 4, Action: "exec", MDL: "SHOW ENTITIES"},
		},
		{
			name:    "navigate without target",
			req:     AgentRequest{ID: 5, Action: "navigate"},
			wantErr: "navigate action requires target field",
		},
		{
			name: "valid navigate",
			req:  AgentRequest{ID: 6, Action: "navigate", Target: "MyModule.Entity"},
		},
		{
			name:    "unknown action",
			req:     AgentRequest{ID: 7, Action: "destroy"},
			wantErr: `unknown action: "destroy"`,
		},
		// New action validations
		{
			name:    "delete without target",
			req:     AgentRequest{ID: 8, Action: "delete"},
			wantErr: "delete action requires target field",
		},
		{
			name: "valid delete",
			req:  AgentRequest{ID: 9, Action: "delete", Target: "entity:MyModule.Customer"},
		},
		{
			name:    "create_module without name",
			req:     AgentRequest{ID: 10, Action: "create_module"},
			wantErr: "create_module action requires name field",
		},
		{
			name: "valid create_module",
			req:  AgentRequest{ID: 11, Action: "create_module", Name: "NewModule"},
		},
		{
			name:    "format without mdl",
			req:     AgentRequest{ID: 12, Action: "format"},
			wantErr: "format action requires mdl field",
		},
		{
			name: "valid format",
			req:  AgentRequest{ID: 13, Action: "format", MDL: "CREATE ENTITY Mod.Foo"},
		},
		{
			name:    "describe without target",
			req:     AgentRequest{ID: 14, Action: "describe"},
			wantErr: "describe action requires target field",
		},
		{
			name: "valid describe",
			req:  AgentRequest{ID: 15, Action: "describe", Target: "entity:MyModule.Customer"},
		},
		{
			name:    "list without target",
			req:     AgentRequest{ID: 16, Action: "list"},
			wantErr: "list action requires target field",
		},
		{
			name: "valid list",
			req:  AgentRequest{ID: 17, Action: "list", Target: "entities"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantType  string
		wantQName string
	}{
		{
			name:      "type with qualified name",
			target:    "entity:Module.Entity",
			wantType:  "entity",
			wantQName: "Module.Entity",
		},
		{
			name:      "type only without colon",
			target:    "entities",
			wantType:  "entities",
			wantQName: "",
		},
		{
			name:      "type with scope",
			target:    "entities:Mod",
			wantType:  "entities",
			wantQName: "Mod",
		},
		{
			name:      "empty string",
			target:    "",
			wantType:  "",
			wantQName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotQName := parseTarget(tt.target)
			if gotType != tt.wantType {
				t.Errorf("type: got %q, want %q", gotType, tt.wantType)
			}
			if gotQName != tt.wantQName {
				t.Errorf("qname: got %q, want %q", gotQName, tt.wantQName)
			}
		})
	}
}

func TestBuildAgentDescribeCmd(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		want    string
		wantErr string
	}{
		{
			name:   "entity describe",
			target: "entity:Module.Entity",
			want:   "DESCRIBE ENTITY Module.Entity",
		},
		{
			name:   "javaaction describe",
			target: "javaaction:Module.Action",
			want:   "DESCRIBE JAVA ACTION Module.Action",
		},
		{
			name:   "imagecollection describe",
			target: "imagecollection:Module.Icons",
			want:   "DESCRIBE IMAGE COLLECTION Module.Icons",
		},
		{
			name:    "unsupported type",
			target:  "security:Module.X",
			wantErr: `unsupported describe type: "security"`,
		},
		{
			name:    "missing qualified name",
			target:  "entity",
			wantErr: "describe target must be type:QualifiedName (e.g. entity:Module.Entity)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAgentDescribeCmd(tt.target)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildListCmd(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		want    string
		wantErr string
	}{
		{
			name:   "entities",
			target: "entities",
			want:   "SHOW ENTITIES",
		},
		{
			name:   "entities with scope",
			target: "entities:Mod",
			want:   "SHOW ENTITIES IN Mod",
		},
		{
			name:   "imagecollections",
			target: "imagecollections",
			want:   "SHOW IMAGE COLLECTIONS",
		},
		{
			name:   "javaactions with scope",
			target: "javaactions:Mod",
			want:   "SHOW JAVA ACTIONS IN Mod",
		},
		{
			name:   "microflows",
			target: "microflows",
			want:   "SHOW MICROFLOWS",
		},
		{
			name:   "modules",
			target: "modules",
			want:   "SHOW MODULES",
		},
		{
			name:   "nanoflows",
			target: "nanoflows",
			want:   "SHOW NANOFLOWS",
		},
		{
			name:   "layouts",
			target: "layouts",
			want:   "SHOW LAYOUTS",
		},
		{
			name:   "snippets",
			target: "snippets",
			want:   "SHOW SNIPPETS",
		},
		{
			name:   "nanoflows with scope",
			target: "nanoflows:Mod",
			want:   "SHOW NANOFLOWS IN Mod",
		},
		{
			name:    "unsupported type",
			target:  "foobar",
			wantErr: `unsupported list type: "foobar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildListCmd(tt.target)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
