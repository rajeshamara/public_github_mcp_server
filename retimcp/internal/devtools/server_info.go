package devtools

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverVersion = "0.6.0-public.1"

type ServerInfo struct {
	Name                string       `json:"name"`
	Version             string       `json:"version"`
	Tools               []string     `json:"tools"`
	AuthoringBranch     string       `json:"authoring_branch"`
	ExecutionHost       string       `json:"execution_host"`
	Capabilities        []string     `json:"capabilities"`
	Limits              ServerLimits `json:"limits"`
	RefreshInstructions string       `json:"refresh_instructions"`
}
type ServerLimits struct {
	TestFolder         string   `json:"test_folder"`
	TestPattern        string   `json:"test_pattern"`
	TestTimeoutSeconds int      `json:"test_timeout_seconds"`
	OutputLimitBytes   int      `json:"output_limit_bytes"`
	SnapshotLimitBytes int      `json:"snapshot_limit_bytes"`
	SnapshotFileLimit  int      `json:"snapshot_file_limit"`
	Unsupported        []string `json:"unsupported"`
}

func serverInfo(c Config) ServerInfo {
	return ServerInfo{
		Name: "retimcp-dev", Version: serverVersion,
		Tools:           []string{"git_status", "git_pull", "run_unit_tests", "server_info"},
		AuthoringBranch: c.Branch, ExecutionHost: "configured-host",
		Capabilities: []string{
			"Inspect the configured execution worktree",
			"Fast-forward its configured branch to an exact remote commit",
			"Run isolated Python standard-library unit tests",
			"Report server capabilities",
		},
		Limits: ServerLimits{
			TestFolder: testFolder, TestPattern: "test_*.py", TestTimeoutSeconds: 60,
			OutputLimitBytes: 65536, SnapshotLimitBytes: 10485760, SnapshotFileLimit: 500,
			Unsupported: []string{
				"Source authoring",
				"Arbitrary commands or caller-selected test paths",
				"Package installation",
				"Arbitrary browser URLs, accounts, or credentials",
				"Browser automation",
				"Desktop plugin refresh and self-deployment",
				"Arbitrary service restart",
			},
		},
		RefreshInstructions: "After an operator deploys a new server build, refresh or reconnect your MCP client and inspect its actual callable tools. The public edition does not automate deployment or desktop refresh. A tool name in server_info does not itself register a callable tool.",
	}
}

// registerServerInfo keeps the server_info MCP contract beside its implementation.
func registerServerInfo(server *mcp.Server, c Config) {
	no := false
	mcp.AddTool(server, &mcp.Tool{Name: "server_info", Description: "Read-only information about the running RetiMCP Dev version, available tool names, supported capabilities, execution limits, and how to refresh ChatGPT when tools are missing. Does not access or modify the worktree, fetch code, run tests, or expose credentials. Listed names do not create callable tools in the client.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &no, IdempotentHint: true, OpenWorldHint: &no}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyArgs) (*mcp.CallToolResult, ServerInfo, error) {
		return nil, serverInfo(c), nil
	})

}
