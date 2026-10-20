package devtools

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{Worktree: "/srv/worktrees/agent-execution", Branch: "agent-execution", Remote: "https://github.com/example/project.git", GitHubUserID: 12345, StateDir: "/srv/retimcp-state", PublicURL: "https://mcp.example.invalid"}
}

func TestConfigurationRefusesUnsafeOrIncompleteValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Config)
	}{
		{"missing_owner", func(c *Config) { c.GitHubUserID = 0 }},
		{"missing_origin", func(c *Config) { c.Remote = "" }},
		{"relative_worktree", func(c *Config) { c.Worktree = "worktree" }},
		{"root_worktree", func(c *Config) { c.Worktree = "/" }},
		{"state_in_worktree", func(c *Config) { c.StateDir = c.Worktree + "/state" }},
		{"protected_branch", func(c *Config) { c.Branch = "main" }},
		{"option_branch", func(c *Config) { c.Branch = "--all" }},
		{"ref_range", func(c *Config) { c.Branch = "one..two" }},
		{"missing_hostname", func(c *Config) { c.PublicURL = "https://:443" }},
		{"missing_public_origin", func(c *Config) { c.PublicURL = "" }},
		{"external_plain_http", func(c *Config) { c.PublicURL = "http://example.invalid" }},
		{"embedded_credentials", func(c *Config) { c.PublicURL = "https://user:password@example.invalid" }},
		{"origin_path", func(c *Config) { c.PublicURL += "/untrusted" }},
		{"origin_query", func(c *Config) { c.PublicURL += "?secret=example" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := validConfig()
			tc.change(&c)
			if c.Validate() == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
	for _, origin := range []string{"https://mcp.example.invalid", "http://127.0.0.1:8082", "http://localhost:8082", "http://[::1]:8082"} {
		c := validConfig()
		c.PublicURL = origin
		if err := c.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOAuthMetadataUsesConfiguredOrigin(t *testing.T) {
	c := validConfig()
	r := httptest.NewRequest("GET", "http://untrusted.example/.well-known/oauth-protected-resource/dev", nil)
	w := httptest.NewRecorder()
	Metadata(c)(w, r)
	if !strings.Contains(w.Body.String(), "https://mcp.example.invalid/dev") || strings.Contains(w.Body.String(), "untrusted.example") {
		t.Fatal("metadata reflected request host")
	}
	w = httptest.NewRecorder()
	challenge(w, c)
	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "https://mcp.example.invalid/.well-known/oauth-protected-resource/dev") {
		t.Fatal("challenge does not use the configured origin")
	}
}
