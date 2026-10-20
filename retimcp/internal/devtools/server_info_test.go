package devtools

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Moving registrations out of the transport must not silently hide a tool.
// Exercise the authenticated MCP protocol rather than inspecting source layout.
func TestRegisteredToolCatalogMatchesServerInfo(t *testing.T) {
	c := Config{Worktree: "/unused", Branch: "agent-execution", Remote: "unused", StateDir: "/unused-state", GitHubUserID: 1, PublicURL: "https://mcp.example.invalid"}
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":1}`)), Header: make(http.Header)}, nil
	})}
	h, err := handler(c, client)
	if err != nil {
		t.Fatal(err)
	}
	// Keep this explicit catalog synchronized with the production registrations
	// and server_info output. Adding a tool without updating this list should fail
	// the test so the client-visible schema and server metadata cannot drift apart.
	expected := []string{"git_pull", "git_status", "run_unit_tests", "server_info"}
	call := func(body string) json.RawMessage {
		t.Helper()
		req := httptest.NewRequest("POST", "http://localhost/dev", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test-owner-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		payload := strings.TrimSpace(w.Body.String())
		if strings.HasPrefix(w.Header().Get("Content-Type"), "text/event-stream") {
			for _, line := range strings.Split(payload, "\n") {
				if strings.HasPrefix(line, "data: ") {
					payload = strings.TrimPrefix(line, "data: ")
					break
				}
			}
		}
		var envelope struct {
			Result json.RawMessage
			Error  json.RawMessage
		}
		if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
			t.Fatal(err)
		}
		if len(envelope.Error) != 0 {
			t.Fatalf("MCP error: %s", envelope.Error)
		}
		return envelope.Result
	}
	var list struct{ Tools []struct{ Name string } }
	if err := json.Unmarshal(call(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`), &list); err != nil {
		t.Fatal(err)
	}
	var actual []string
	for _, tool := range list.Tools {
		actual = append(actual, tool.Name)
	}
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("registered tools = %v; want %v", actual, expected)
	}
	var info struct {
		StructuredContent ServerInfo `json:"structuredContent"`
	}
	if err := json.Unmarshal(call(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"server_info","arguments":{}}}`), &info); err != nil {
		t.Fatal(err)
	}
	sort.Strings(info.StructuredContent.Tools)
	if !reflect.DeepEqual(info.StructuredContent.Tools, expected) {
		t.Fatalf("server_info catalog = %v; want %v", info.StructuredContent.Tools, expected)
	}
}

// Retain the existing metadata values through this organizational change.
func TestServerInfoVersion(t *testing.T) {
	if serverInfo(Config{}).Version != serverVersion {
		t.Fatal("server info version mismatch")
	}
}
