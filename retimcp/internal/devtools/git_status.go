package devtools

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerGitStatus keeps the git_status MCP contract beside its implementation.
func registerGitStatus(server *mcp.Server, s *Service) {
	mcp.AddTool(server, &mcp.Tool{Name: "git_status", Description: "Read-only inspection of the configured execution host execution worktree on the configured execution branch. GitHub remains the source/code-authoring side. Use before pulling to inspect readiness and after pulling to verify the local commit and clean state. Does not fetch, author code, edit source files, or run tests.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyArgs) (*mcp.CallToolResult, Result, error) {
		r, e := s.Status(ctx)
		return nil, r, e
	})

}

// Status inspects the registered worktree under the shared lock and audit trail.
func (s *Service) Status(ctx context.Context) (Result, error) {
	return s.locked(ctx, "git_status", s.inspect)
}
