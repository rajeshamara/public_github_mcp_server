package devtools

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestAuthorization(t *testing.T) {
	c := Config{Worktree: "/unused", Branch: "agent-execution", Remote: "unused", StateDir: "/unused-state", GitHubUserID: 12345, PublicURL: "https://mcp.example.invalid"}
	for _, tc := range []struct {
		name, token, body string
		upstream, want    int
	}{
		{"missing", "", "", 200, 401},
		{"invalid", "Bearer invalid-test-token", "", 401, 401},
		{"other_user", "Bearer valid-test-token", `{"id":999}`, 200, 403},
		{"github_unavailable", "Bearer valid-test-token", "", 503, 503},
		{"owner", "Bearer valid-test-token", `{"id":12345}`, 200, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://api.github.com/user" {
					t.Fatal("unexpected token destination")
				}
				return &http.Response{StatusCode: tc.upstream, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			h, e := handler(c, client)
			if e != nil {
				t.Fatal(e)
			}
			body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`
			r := httptest.NewRequest("POST", "http://localhost/dev", strings.NewReader(body))
			r.Header.Set("Authorization", tc.token)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Accept", "application/json, text/event-stream")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestUnitTestToolProtocol(t *testing.T) {
	requireSandbox(t)
	s, _, _ := fixture(t)
	sha := commitTests(t, s, "import unittest\nclass Example(unittest.TestCase):\n def test_ok(self): self.assertEqual(2+3,5)\n")
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":1}`)), Header: make(http.Header)}, nil
	})}
	h, e := handler(s.Config, client)
	if e != nil {
		t.Fatal(e)
	}
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"run_unit_tests","arguments":{"expected_commit":"` + sha + `"}}}`,
	} {
		req := httptest.NewRequest("POST", "http://localhost/dev", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test-owner-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		if strings.Contains(body, "tools/list") {
			if !strings.Contains(w.Body.String(), "run_unit_tests") {
				t.Fatal(w.Body.String())
			}
		} else if !strings.Contains(w.Body.String(), `"status":"passed"`) {
			t.Fatal(w.Body.String())
		}
	}
}
