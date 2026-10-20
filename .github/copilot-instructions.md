# RetiMCP public edition: contribution guidance

This is Rajesh Amara's independent RetiMCP edition of the GitHub MCP Server. Read `README.md`, `CONTRIBUTING.md`, and `NOTICE.md` for scope, contribution expectations, and upstream provenance. The original upstream introduction is preserved in `UPSTREAM_README.md`.

## Source layout

- `retimcp/cmd/gateway/`: the custom Go gateway and embedded landing page.
- `retimcp/internal/devtools/`: the owner-authenticated development endpoint and its four tools: `git_status`, `git_pull`, `run_unit_tests`, and `server_info`.
- `retimcp/config/devbox.example.json`: a nonfunctional example until the operator provides valid configuration.
- `cmd/github-mcp-server/`, `pkg/`, and `internal/`: the included upstream server implementation.
- `ui/`: the upstream browser UI.

Use the Go version in `go.mod`. Do not infer the current version, line count, capabilities, or deployment state from old upstream notes.

## Preserve the execution boundary

Keep owner-token validation, explicit numeric owner configuration, Git branch/origin checks, exact-commit checks, clean-worktree refusal, fast-forward-only updates, locking, deadlines, and bounded output intact.

The test runner exports only committed regular files from its fixed test folder and requires bubblewrap isolation. Do not introduce an unsandboxed fallback, arbitrary command execution, caller-selected paths, automatic package installation, browser credentials, or self-deployment.

The actual MCP registrations, `server_info` catalog, descriptions, documentation, and protocol tests must agree. Tool names in metadata do not establish that a client has registered those tools.

## Focused changes and validation

Keep changes small and explain the resulting behavior. Use meaningful tests for behavior changes; documentation-only edits do not require an application test run.

For custom Go changes, format the changed Go files and run:

```bash
go test ./retimcp/...
go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway
```

Some sandbox execution tests skip when the host cannot provide bubblewrap namespaces. Report skipped tests and other limitations explicitly. Do not claim that a skipped execution test passed.

For changes to upstream packages, run relevant package tests and applicable lint/build checks. Preserve existing exported API patterns. Tool schema snapshots live in `pkg/github/__toolsnaps__/`; update them only for intentional upstream schema changes. Consult `docs/tool-renaming.md` and `docs/toolsets-and-icons.md` for those upstream mechanisms. Custom RetiMCP tools use their own protocol/catalog tests.

For UI changes, use the scripts declared in `ui/package.json` and report the checks run. Use mocked responses and temporary repositories where appropriate; tests must not rely on private infrastructure or real credentials.

## Documentation and repository automation

Preserve the RetiMCP root README and the original upstream README separately. The included upstream `script/generate-docs` targets `README.md`; do not run it in a way that replaces the RetiMCP project overview. Update the relevant documentation deliberately.

This edition omits automatic GitHub Actions workflows and Dependabot configuration. There are no inherited CI, release, registry, container publishing, or automatic dependency-update guarantees. Remaining upstream helper files do not by themselves establish enabled automation.

## Public content

Use synthetic examples and explicit operator configuration. Never add credentials, local operator configuration, private application code, personal data, deployment state, live infrastructure identifiers, or original private Git history.

Keep the MIT license and upstream attribution. Use this repository's support, security, and conduct policies; do not direct reports about custom RetiMCP code to upstream maintainers as though they maintain this edition.
