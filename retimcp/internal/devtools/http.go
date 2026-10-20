package devtools

import (
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"strings"
	"time"
)

func Handler(c Config) (http.Handler, error) {
	return handler(c, &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }})
}
func handler(c Config, client *http.Client) (http.Handler, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	s := &Service{Config: c}
	server := mcp.NewServer(&mcp.Implementation{Name: "retimcp-dev", Version: serverVersion}, &mcp.ServerOptions{
		Instructions: "RetiMCP operates only on one operator-configured execution worktree. Author and commit source through your normal Git workflow, obtain the full SHA, and use git_pull(expected_commit) to fast-forward the configured branch to that exact remote commit. Verify git_status before executing run_unit_tests with the same SHA. Test execution exports only committed files from proof_of_work/chatgpt into a Linux bubblewrap sandbox with no network or host credentials. This server cannot author source, choose arbitrary paths or commands, install packages, automate browsers, refresh desktop plugins, or deploy itself. Tool names returned by server_info are metadata, not proof of client tool availability. Never claim a test ran or passed unless the tool result reports it.",
	})
	// Each tool owns its registration, schema, and implementation in its namesake file.
	registerGitStatus(server, s)
	registerGitPull(server, s)
	registerRunUnitTests(server, s)
	registerServerInfo(server, c)
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		value := r.Header.Get("Authorization")
		if !strings.HasPrefix(value, "Bearer ") || len(strings.TrimPrefix(value, "Bearer ")) < 10 {
			challenge(w, c)
			return
		}
		req, err := http.NewRequestWithContext(r.Context(), "GET", "https://api.github.com/user", nil)
		if err != nil {
			http.Error(w, "Authentication unavailable", 503)
			return
		}
		req.Header.Set("Authorization", value)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "retimcp-dev/0.1")
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Authentication unavailable", 503)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 401 {
			challenge(w, c)
			return
		}
		if resp.StatusCode != 200 {
			http.Error(w, "Authentication unavailable", 503)
			return
		}
		var user struct {
			ID int64 `json:"id"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, resp.Body, 65536)).Decode(&user) != nil {
			http.Error(w, "Authentication unavailable", 503)
			return
		}
		if user.ID != c.GitHubUserID {
			http.Error(w, "Account is not authorized for this development server", 403)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		transport.ServeHTTP(w, r)
	}), nil
}
func challenge(w http.ResponseWriter, c Config) {
	w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+strings.TrimRight(c.PublicURL, "/")+`/.well-known/oauth-protected-resource/dev"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Metadata uses the operator's validated origin, never an untrusted Host header.
func Metadata(c Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":                 strings.TrimRight(c.PublicURL, "/") + "/dev",
			"authorization_servers":    []string{"https://github.com/login/oauth"},
			"bearer_methods_supported": []string{"header"},
		})
	}
}
